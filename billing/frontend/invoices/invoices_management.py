import requests
import pandas as pd
from tabulate import tabulate
from requests.exceptions import ConnectionError, Timeout
from os import path as os_path
from base64 import b64decode


# Configurações
endpoint = 'http://localhost:8080'
page = 1
page_size = 1000


# get_df 
def get_df(vendor, customer, invoicing, due, payment, email_sent, whatsapp_sent, email_receipt, whatsapp_receipt, tax, cancellation):
   # build request payload
    json_data = {'vendor': vendor, 'page': page, 'page_size': page_size}
    if customer and customer != "":
        json_data['customer'] = customer
    if invoicing and invoicing != "":
        json_data['invoicing'] = invoicing
    if due and due != "":
        json_data['due'] = due
    if payment and payment != "":
        json_data['payment'] = payment
    if cancellation and cancellation != "":
        json_data['cancellation'] = cancellation
    if email_sent and email_sent != "":
        json_data['email_sent'] = email_sent
    if whatsapp_sent and whatsapp_sent != "":
        json_data['whatsapp_sent'] = whatsapp_sent
    if email_receipt and email_receipt != "":
        json_data['email_receipt'] = email_receipt
    if whatsapp_receipt and whatsapp_receipt != "":
        json_data['whatsapp_receipt'] = whatsapp_receipt
    if tax and tax != "":
        json_data['tax'] = tax
    if cancellation and cancellation != "":
        json_data['cancellation'] = cancellation
    # make API call with error handling
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
    df_invoices = pd.DataFrame(json_data['invoices'])
    df_invoices = df_invoices.fillna('-')
    lb = lambda items: '\n'.join([f"{item['description']} (qtty: {item['quantity']}, price: {item['price']})" 
                                  for item in items]) if isinstance(items, list) else '-'
    df_invoices['items'] = df_invoices['items'].apply(lb)
    df_invoices = df_invoices.sort_values(by=['customer', 'invoicing'])
    return df_invoices

# get
def get(vendor, customer, invoicing, due, payment, email_sent, whatsapp_sent, email_receipt, whatsapp_receipt, tax, cancellation):
    df_invoices = get_df(vendor, customer, invoicing, due, payment, email_sent, whatsapp_sent, email_receipt, whatsapp_receipt, tax, cancellation)
    if isinstance(df_invoices, str):
        return df_invoices, 0, 0, 0
    if len(df_invoices) == 0:
        return 'Nenhuma fatura encontrada.', 0, 0, 0
    return tabulate(df_invoices, headers='keys', tablefmt='grid', showindex=False), \
        len(df_invoices), \
        sum(df_invoices['amount'].replace('-', 0).astype(float)), \
        df_invoices.loc[df_invoices['payment'] != '-', 'amount'].replace('-', 0).astype(float).sum()

def get_overdue(vendor):
    df = get_df(vendor, '', '', '', 'null', '', '', '', '', '', '')
    if isinstance(df, str):
        return df, 0, 0, 0
    if len(df) == 0:
        return 'Nenhuma fatura encontrada.', 0, 0, 0
    df['due'] = pd.to_datetime(df['due'], errors='coerce')
    df['overdue'] = df['due'] < pd.Timestamp.now().replace(hour=0, minute=0, second=0, microsecond=0)
    df_overdue = df[df['overdue']]
    if len(df_overdue) == 0:
        return 'Nenhuma fatura vencida encontrada.', 0, 0, 0
    return tabulate(df_overdue, headers='keys', tablefmt='grid', showindex=False), \
        len(df_overdue), \
        sum(df_overdue['amount'].replace('-', 0).astype(float)), \
        df_overdue.loc[df_overdue['payment'] != '-', 'amount'].replace('-', 0).astype(float).sum()

# insert
def insert(vendor, customer, invoicing, due, payment, cancellation, notes, items):
    invoice = {'vendor': vendor, 'customer': customer, 'invoicing': invoicing,
        'due': due, 'items': items
    }
    if payment and payment != "":
        invoice['payment'] = payment
    if cancellation and cancellation != "":
        invoice['cancellation'] = cancellation
    if notes and notes != "":
        invoice['notes'] = notes
     
    json_data = {"items": [invoice]}
    try:
        resposta = requests.post(f'{endpoint}/invoice/create', json=json_data, timeout=5)
    except ConnectionError as e:
        return f"Erro: A conexão foi recusada pelo servidor remoto. Detalhes: {e}"
    except Timeout as e:
        return f"Erro: A requisição excedeu o tempo limite estabelecido. {e}"
    except requests.exceptions.RequestException as e:
        return f"Ocorreu um erro genérico no requests: {e}"
    resp = resposta.json()
    return f'{resp["status"]} - {resp["message"]}'


# update
def update(vendor, id, invoicing, due, payment, email_sent, whatsapp_sent, tax, cancellation):
    invoice = {'vendor': vendor, 'id': id}
    if invoicing and invoicing != "":
        invoice['invoicing'] = invoicing
    if due and due != "":
        invoice['due'] = due
    if payment and payment != "":
        invoice['payment'] = payment
    if cancellation and cancellation != "":
        invoice['cancellation'] = cancellation
    if email_sent and email_sent != "":
        invoice['email_sent'] = email_sent
    if whatsapp_sent and whatsapp_sent != "":
        invoice['whatsapp_sent'] = whatsapp_sent
    if tax and tax != "":
        invoice['tax'] = tax
    try:
        resposta = requests.patch(f'{endpoint}/invoice/update', json=invoice, timeout=5)
    except ConnectionError as e:
        return f"Erro: A conexão foi recusada pelo servidor remoto. Detalhes: {e}"
    except Timeout as e:
        return f"Erro: A requisição excedeu o tempo limite estabelecido. {e}"
    except requests.exceptions.RequestException as e:
        return f"Ocorreu um erro genérico no requests: {e}"
    resp = resposta.json()
    return f'{resp["status"]} - {resp["message"]}'

# insertCSV
def insert_csv(vendor, csv_file):
    try:
        df = pd.read_csv(csv_file)
        df = df.fillna('')
        df = df.map(lambda x: x.strip() if isinstance(x, str) else x)
    except FileNotFoundError:
        return [f"Erro: O arquivo {csv_file} não foi encontrado."]
    except pd.errors.EmptyDataError:
        return [f"Erro: O arquivo {csv_file} está vazio."]
    except pd.errors.ParserError as e:
        return [f"Erro ao analisar o arquivo CSV: {e}"]
    except Exception as e:
        return [f"Ocorreu um erro ao ler o arquivo CSV: {e}"]

    responses = []
    descs = df.filter(like='description').columns
    qttys = df.filter(like='quantity').columns
    prices = df.filter(like='price').columns
    for _, row in df.iterrows():
        customer = row.get('customer', '')
        invoicing = format_date(row.get('invoicing', ''))
        due = format_date(row.get('due', ''))
        payment = format_date(row.get('payment', ''))
        cancellation = format_date(row.get('cancellation', ''))
        notes = row.get('notes', '')
        items = []
        for desc_col, qtty_col, price_col in zip(descs, qttys, prices):
            if row.get(desc_col, '') != '' and row.get(qtty_col, 0) != 0 and row.get(price_col, 0.0) != 0.0:
                description = row.get(desc_col, '')
                qtty = int(row.get(qtty_col, 0))
                price = float(row.get(price_col, 0.0))
                items.append({'description': description, 'quantity': qtty, 'price': price})
        payload = {'vendor': vendor, 'customer': customer, 'invoicing': invoicing,
                   'due': due, 'payment': payment, 'cancellation': cancellation, 
                   'notes': notes, 'items': items}    
        resp = insert(vendor, customer, invoicing, due, payment, cancellation, notes, items)
        responses.append(f"Payload '{payload}'.Resposta: {resp}")
    return responses

# format_date
def format_date(date_str):
    if date_str in [None, '', '-']:
        return date_str
    try:
        return pd.to_datetime(date_str, format='%d/%m/%Y').strftime('%Y-%m-%d')
    except Exception as e:
        return date_str
    
# send receipt
def send_bill(vendor, invoiceID, doc, email):
    action = 0
    if email != '':
        action = 1
    json_data = {'vendor': vendor, 'doc': doc, 'invoice_id': invoiceID, 'action': action, 'email': email}
    try:
        resposta = requests.post(f'{endpoint}/invoice/bill', json=json_data, timeout=15)
    except ConnectionError as e:
        return f"Erro: A conexão foi recusada pelo servidor remoto. Detalhes: {e}"
    except Timeout as e:
        return f"Erro: A requisição excedeu o tempo limite estabelecido. {e}"
    except requests.exceptions.RequestException as e:
        return f"Ocorreu um erro genérico no requests: {e}"
    if resposta.status_code != 200:
        return f'Erro na chamada da API: {resposta.status_code} - {resposta.text}'
    resp = resposta.json()
    return f'{resp["status"]} - {resp["message"]}'


def save_bill(vendor, invoiceID, doc, path):
    json_data = {'vendor': vendor, 'doc': doc, 'invoice_id': invoiceID, 'action': 2}
    try:
        resposta = requests.post(f'{endpoint}/invoice/bill', json=json_data, timeout=5)
    except ConnectionError as e:
        return f"Erro: A conexão foi recusada pelo servidor remoto. Detalhes: {e}"
    except Timeout as e:
        return f"Erro: A requisição excedeu o tempo limite estabelecido. {e}"
    except requests.exceptions.RequestException as e:
        return f"Ocorreu um erro genérico no requests: {e}"
    if resposta.status_code != 200:
        return f'Erro na chamada da API: {resposta.status_code} - {resposta.text}'
    resp = resposta.json()
    file_path = os_path.join(path, resp['document_name'])
    file_bin = b64decode(resp['document_base64'])
    with open(file_path, 'wb') as f:
        f.write(file_bin)
    return f'Arquivo salvo com sucesso em: {file_path}'