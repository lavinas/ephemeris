from requests import status_codes
import requests
import pandas as pd
from tabulate import tabulate
from requests.exceptions import ConnectionError, Timeout
from os import path

# Configurações
# endpoint = 'http://192.168.1.138:8081'
endpoint = 'http://localhost:8082/api'

page = 1
page_size = 1000

# get
def get(nickname, legal_name, trading_name, document, email, whatsapp):
    # build request payload
    json_data = {'page': page, 'page_size': page_size}
    if nickname and nickname != "":
        json_data['nickname'] = nickname
    if legal_name and legal_name != "":
        json_data['legal_name'] = legal_name
    if trading_name and trading_name != "":
        json_data['trading_name'] = trading_name
    if document and document != "":
        json_data['document'] = document
    if email and email != "":
        json_data['email'] = email
    if whatsapp and whatsapp != "":
        json_data['whatsapp'] = whatsapp
    # make API call with error handling
    try:
        resposta = requests.get(f'{endpoint}/vendor/list', json=json_data, timeout=5)
    except ConnectionError as e:
        return f"Erro: A conexão foi recusada pelo servidor remoto. Detalhes: {e}", 0
    except Timeout as e:
        return f"Erro: A requisição excedeu o tempo limite estabelecido. {e}", 0
    except requests.exceptions.RequestException as e:
        return f"Ocorreu um erro genérico no requests: {e}", 0
    # verify response status code
    if resposta.status_code != 200:
        return f'Erro na chamada da API: {resposta.status_code} - {resposta.text}', 0
    # processing response data
    json_data = resposta.json()
    if 'vendors' not in json_data or len(json_data['vendors']) == 0:
        return 'Nenhum usuario encontrado.', 0
    # exibir usuarios
    users = json_data['vendors']
    df = pd.DataFrame(users)
    df = df.fillna('-')
    df = df.sort_values(by=['nickname'])
    return tabulate(df, headers='keys', tablefmt='grid', showindex=False), len(users)

# insert
def insert(nickname, legal_name, trading_name, document, tax_document, account_bank, 
account_agency, account_number, pix_token, pix_name, pix_city, logo_name, email, 
whatsapp, last_rps, smtp_host, smtp_port, smtp_user, smtp_password):
    json_data = {'nickname': nickname, 'legal_name': legal_name, 'trading_name': trading_name, 
                'document': document, 'tax_document': tax_document, 'account_bank': account_bank, 
                'account_agency': account_agency, 'account_number': account_number, 'pix_token': pix_token, 
                'pix_name': pix_name, 'pix_city': pix_city, 'logo_name': logo_name, 'email': email, 
                'whatsapp': whatsapp, 'last_rps': last_rps, 'smtp_host': smtp_host, 'smtp_port': smtp_port, 
                'smtp_user': smtp_user, 'smtp_password': smtp_password}
    try:
        resposta = requests.post(f'{endpoint}/vendor/create', json=json_data, timeout=5)
    except ConnectionError as e:
        return f"Erro: A conexão foi recusada pelo servidor remoto. Detalhes: {e}"
    except Timeout as e:
        return f"Erro: A requisição excedeu o tempo limite estabelecido. {e}"
    except requests.exceptions.RequestException as e:
        return f"Ocorreu um erro genérico no requests: {e}"
    if resposta.status_code != 200:
        return f'Erro na chamada da API: {resposta.status_code} - {resposta.text}'
    json_data = resposta.json()
    return f'{json_data["status"]} - {json_data["message"]} '

# update
def update(nickname, legal_name, trading_name, document, tax_document, account_bank, 
account_agency, account_number, pix_token, pix_name, pix_city, logo_name, email, 
whatsapp, last_rps, smtp_host, smtp_port, smtp_user, smtp_password):
    json_data = {'nickname': nickname}
    if legal_name and legal_name != "":
        json_data['legal_name'] = legal_name
    if trading_name and trading_name != "":
        json_data['trading_name'] = trading_name
    if document and document != "":
        json_data['document'] = document
    if tax_document and tax_document != "":
        json_data['tax_document'] = tax_document
    if account_bank and account_bank != "":
        json_data['account_bank'] = account_bank
    if account_agency and account_agency != "":
        json_data['account_agency'] = account_agency
    if account_number and account_number != "":
        json_data['account_number'] = account_number
    if pix_token and pix_token != "":
        json_data['pix_token'] = pix_token
    if pix_name and pix_name != "":
        json_data['pix_name'] = pix_name
    if pix_city and pix_city != "":
        json_data['pix_city'] = pix_city
    if logo_name and logo_name != "":
        json_data['logo_name'] = logo_name
    if email and email != "":
        json_data['email'] = email
    if whatsapp and whatsapp != "":
        json_data['whatsapp'] = whatsapp
    if last_rps and last_rps != "":
        json_data['last_rps'] = last_rps
    if smtp_host and smtp_host != "":
        json_data['smtp_host'] = smtp_host
    if smtp_port and smtp_port != "":
        json_data['smtp_port'] = smtp_port
    if smtp_user and smtp_user != "":
        json_data['smtp_user'] = smtp_user
    if smtp_password and smtp_password != "":
        json_data['smtp_password'] = smtp_password
    try:
        resposta = requests.patch(f'{endpoint}/vendor/update', json=json_data, timeout=5)
    except ConnectionError as e:
        return f"Erro: A conexão foi recusada pelo servidor remoto. Detalhes: {e}"
    except Timeout as e:
        return f"Erro: A requisição excedeu o tempo limite estabelecido. {e}"
    except requests.exceptions.RequestException as e:
        return f"Ocorreu um erro genérico no requests: {e}"
    if resposta.status_code != 200:
        return f'Erro na chamada da API: {resposta.status_code} - {resposta.text}'
    json_data = resposta.json()
    return f'{json_data["status"]} - {json_data["message"]}'
 