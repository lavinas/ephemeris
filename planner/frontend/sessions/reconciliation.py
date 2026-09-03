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
    df = df[(df['Data'].dt.year == ano_ref) & (df['Data'].dt.month == mes_ref)]
    df['minutes'] = df['Service'].apply(lambda x: 60 if x in ('canto_60', 'piano_60', 'Canto') else (30 if x in ('canto_30', 'piano_30') else 45 if x in ('canto_45', 'canto_60, canto_45') else 0))    
    df['service'] = df['Service'].apply(lambda x: 'aula/piano' if x in ('piano_60', 'piano_30', 'piano_45') else 'aula/canto')
    df['status_row'] = df.apply(lambda row: 'error' if row['minutes'] == 0 or row['status'] == 'ignorar' else 'ok', axis=1)
    df = df.drop(columns=['Status das aulas', 'Service'])
    return df

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
    df_invoices = pd.DataFrame(json_data['invoices'], columns=['customer', 'invoicing', 'amount', 'items'])
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
                return f'{meses_map[mes_nome]:02d}/{ano}'
            match = re.search(r'\b(\d{2})/(\d{2})/(\d{4})\b', description)
            if match:
                return f'{match.group(2)}/{match.group(3)}'
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

        df_invoices['mes_referencia'] = df_invoices['description'].apply(extract_mes_referencia)
        df_invoices['minutos'] = df_invoices['description'].apply(extract_minutos)
        df_invoices['tipo_servico'] = df_invoices['description'].apply(extract_tipo_servico)
        df_invoices['status_row'] = df_invoices[['mes_referencia', 'minutos', 'tipo_servico']].isna().any(axis=1).map({True: 'error', False: 'ok'})

    return df_invoices
