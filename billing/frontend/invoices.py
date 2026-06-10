import requests
import pandas as pd
from tabulate import tabulate
from requests.exceptions import ConnectionError, Timeout


# Configurações
endpoint = 'http://localhost:8080'
page = 1
page_size = 1000

# get
def get(vendor, customer, invoicing, due, email_sent, whatsapp_sent, tax):
    # build request payload
    json_data = {'vendor': vendor, 'page': page, 'page_size': page_size}
    if customer and customer != "":
        json_data['customer'] = customer
    if invoicing and invoicing != "":
        json_data['invoicing'] = invoicing
    if due and due != "":
        json_data['due'] = due
    if email_sent and email_sent != "":
        json_data['email_sent'] = email_sent
    if whatsapp_sent and whatsapp_sent != "":
        json_data['whatsapp_sent'] = whatsapp_sent
    if tax and tax != "":
        json_data['tax'] = tax
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
    itens_list = []
    for invoices in json_data['invoices']:
        for items in invoices['items']:
            itens_list.append({'id': invoices['id'], 'item_id': items['id'], 
                               'description': items['description'], 'quantity': items['quantity'], 
                               'price': items['price']})           
        del invoices['items']
    # create dataframes and merge
    df_invoices = pd.DataFrame(json_data['invoices'])
    df_invoices = df_invoices.fillna('-')
    df_itens = pd.DataFrame(itens_list)
    df = pd.merge(df_invoices, df_itens, left_on='id', right_on='id', how='left')
    # fillna after merge    
    return tabulate(df, headers='keys', tablefmt='grid', showindex=False)    
    

# insert
def insert(vendor, customer, invoicing, due, items):
    json_data = {"items": [{'vendor': vendor, 'customer': customer, 'invoicing': invoicing, 'due': due, 'items': items}]}
    try:
        resposta = requests.post(f'{endpoint}/invoice/create', json=json_data, timeout=5)
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