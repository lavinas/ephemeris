import requests
import pandas as pd
import json
from requests.exceptions import ConnectionError, Timeout

# Configurações
endpoint = 'http://localhost:8080'
page = 1
page_size = 1000


# get_sessions
def get_sessions(invoicing, file): 
    df = pd.read_excel(file, sheet_name=0, usecols=range(1, 5))
    df['status'] = df['Status das aulas'].apply(lambda x: 'realizada' if x in ('done', 'realizada') else ('reposicao' if x in ('reposicao',) else ('cancelada/cobrar' if x in ('faltou', 'missed') else 'ignorar')))
    df['Data'] = pd.to_datetime(df['Data'], errors='coerce')
    ano_ref, mes_ref = map(int, invoicing.split('-'))
    this_month = (ano_ref, mes_ref)
    last_month = (ano_ref, mes_ref - 1) if mes_ref > 1 else (ano_ref - 1, 12)
    df = df[
        df['Data'].apply(
            lambda date: (date.year, date.month) in (this_month, last_month)
            if pd.notna(date) else False
        )
    ]
    df['minutes'] = df['Service'].apply(lambda x: 60 if x in ('canto_60', 'piano_60', 'Canto') else (30 if x in ('canto_30', 'piano_30') else 45 if x in ('canto_45', 'canto_60, canto_45') else 0))    
    df['service'] = df['Service'].apply(lambda x: 'aula/piano' if x in ('piano_60', 'piano_30', 'piano_45') else 'aula/canto')
    df['status_row'] = df.apply(lambda row: 'error' if row['minutes'] == 0 or row['status'] == 'ignorar' else 'ok', axis=1)
    df['status_type'] = df.apply(lambda row: 'realizada/reposicao' if row['status'] in ['realizada', 'reposicao'] else row['status'], axis=1)
    df['month'] = df['Data'].dt.strftime('%Y-%m')
    df = df.drop(columns=['Status das aulas', 'Service'])
    df = df.rename(columns={'nome cliente': 'customer'})
    df_sum = (
        df[['customer', 'service', 'minutes', 'status_row', 'status_type', 'month']]
        .value_counts(dropna=False)
        .rename('done')
        .reset_index()
    )
    return df_sum

# get_invoices
def get_invoices(vendor, invoicing):
   # build request payload
    json_data = {'vendor': vendor, 'page': page, 'page_size': page_size, 'invoicing': invoicing, 'cancellation': 'null'}
    try:
        resposta = requests.get(f'{endpoint}/invoice/list', json=json_data, timeout=5)
    except ConnectionError as e:
        return f"Erro: A conexão foi recusada pelo servidor remoto. Detalhes: {e}"
    except Timeout as e:
        return f"Erro: A requisição excedeu o tempo limite estabelecido. {e}"
    except requests.exceptions.RequestException as e:
        return f"Ocorreu um erro genérico no requests: {e}"
    if resposta.status_code != 200:
        return f'Erro na chamada da API: {resposta.status_code} - {resposta.text}'
    # processing response data
    json_data = resposta.json()
    if 'invoices' not in json_data or len(json_data['invoices']) == 0:
        return 'Nenhuma fatura encontrada.'
    # get items from invoices and merge with invoices data
    df_invoices = pd.DataFrame(json_data['invoices'], columns=['customer', 'invoicing', 'payment', 'amount', 'items'])
    df_invoices = df_invoices.fillna('-')

    if 'items' in df_invoices.columns:
        def parse_items(items):
            if isinstance(items, str):
                try:
                    items = json.loads(items)
                except json.JSONDecodeError:
                    return []
            return items if isinstance(items, list) else []

        df_invoices['items'] = df_invoices['items'].apply(parse_items)
        df_invoices = df_invoices.explode('items', ignore_index=True)
        items_df = pd.json_normalize(df_invoices.pop('items')).reindex(columns=['description', 'quantity', 'price'])
        df_invoices = pd.concat([df_invoices, items_df], axis=1)

        import re

        meses_map = {
            'janeiro': 1, 'fevereiro': 2, 'março': 3, 'abril': 4,
            'maio': 5, 'junho': 6, 'julho': 7, 'agosto': 8,
            'setembro': 9, 'outubro': 10, 'novembro': 11, 'dezembro': 12
        }

        def extract_mes_referencia(description):
            if not isinstance(description, str):
                return float('nan')
            match = re.search(r'\b(' + '|'.join(meses_map.keys()) + r')\s+de\s+(\d{4})\b', description, re.IGNORECASE)
            if match:
                mes_nome = match.group(1).lower()
                ano = match.group(2)
                return f'{ano}-{meses_map[mes_nome]:02d}'
            match = re.search(r'\b(\d{2})/(\d{2})/(\d{4})\b', description)
            if match:
                return f'{match.group(3)}-{match.group(2)}'
            return float('nan')

        def extract_minutos(description):
            if not isinstance(description, str):
                return float('nan')
            match = re.search(r'\b(\d+)\s*minutos\b', description, re.IGNORECASE)
            if match:
                return int(match.group(1))
            return float('nan')

        def extract_tipo_servico(description):
            if not isinstance(description, str):
                return float('nan')
            desc_lower = description.lower()
            if 'aula' in desc_lower and 'canto' in desc_lower:
                return 'aula/canto'
            if 'aula' in desc_lower and 'piano' in desc_lower:
                return 'aula/piano'
            return float('nan')

        df_invoices['month'] = df_invoices['description'].apply(extract_mes_referencia)     
        df_invoices['minutes'] = df_invoices['description'].apply(extract_minutos)
        df_invoices['service'] = df_invoices['description'].apply(extract_tipo_servico)
        df_invoices['payment'] = df_invoices['payment'].apply(lambda x: 'paid' if pd.notna(x) and (x != '') else 'not paid')
        df_invoices['status_row'] = df_invoices[['month', 'minutes', 'service']].isna().any(axis=1).map({True: 'error', False: 'ok'})
        df_invoices = df_invoices.rename(columns={'quantity': 'preview'})

    return df_invoices

# reconciliate
def reconciliate(df_invoices, df_sessions):
    df_merged = df_invoices.merge(
        df_sessions,
        how='left',
        left_on=['customer', 'service', 'month', 'minutes'],
        right_on=['customer', 'service', 'month', 'minutes'],
        suffixes=('_invoice', '_session')
    )
    # Both dataframes contain ``status_row``; pandas adds suffixes during the
    # merge, so the unsuffixed column is not available here.
    df_merged = df_merged[[
        'status_row_invoice', 'month', 'customer', 'service', 'minutes',
        'payment', 'preview', 'done', 
    ]].rename(columns={'status_row_invoice': 'status_row'})
    df_merged['preview'] = pd.to_numeric(df_merged['preview'], errors='coerce')
    df_merged['done'] = pd.to_numeric(df_merged['done'], errors='coerce').fillna(0).astype(int)
    df_merged['balance'] = df_merged['preview'] - df_merged['done']
    return df_merged

# extras
def extras(df_invoices, df_sessions, invoicing_month):
    df_sessions = df_sessions.reset_index()
    df_sessions = df_sessions[df_sessions['month'] == invoicing_month]
    df_merged = df_sessions.merge(
        df_invoices,
        how='left',
        left_on=['customer', 'service', 'minutes'],
        right_on=['customer', 'service', 'minutes'],
        suffixes=('_session', '_invoice'),
        indicator=True
    )
    df_extras = df_merged[df_merged['_merge'] == 'left_only']
    df_extras = df_extras[['customer', 'service', 'minutes', 'done']]
    return df_extras