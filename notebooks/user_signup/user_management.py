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
def get(vendor, status, username, name, email, whatsapp):
    # build request payload
    json_data = {'vendor': vendor, 'page': page, 'page_size': page_size}
    if username and username != "":
        json_data['username'] = username
    if name and name != "":
        json_data['name'] = name
    if email and email != "":
        json_data['email'] = email
    if whatsapp and whatsapp != "":
        json_data['whatsapp'] = whatsapp
    if status and status != "":
        if status not in ['0', '1', '-1']:
            return "Erro: Status deve ser 0 (ativo), 1 (inativo) ou -1 (todos).", 0
        json_data['status'] = int(status)
    # make API call with error handling
    try:
        resposta = requests.get(f'{endpoint}/user/list', json=json_data, timeout=5)
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
    if 'users' not in json_data or len(json_data['users']) == 0:
        return 'Nenhum usuario encontrado.', 0
    # exibir usuarios
    users = json_data['users']
    df = pd.DataFrame(users)
    df = df.fillna('-')
    df = df.sort_values(by=['username'])
    return tabulate(df, headers='keys', tablefmt='grid', showindex=False), len(users)

# insert
def insert(vendor, username, password, name, email, whatsapp):
    json_data = {'vendor': vendor, 'username': username, 'password': password, 'name': name, 
                 'email': email, 'whatsapp': whatsapp}
    try:
        resposta = requests.post(f'{endpoint}/user/create', json=json_data, timeout=5)
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

# update
def update(id, vendor, username, password, name, email, whatsapp, status):
    json_data = {'vendor': vendor}
    if username and username != "":
        json_data['username'] = username
    if password and password != "":
        json_data['password'] = password
    if name and name != "":
        json_data['name'] = name
    if email and email != "":
        json_data['email'] = email
    if whatsapp and whatsapp != "":
        json_data['whatsapp'] = whatsapp
    if status and status != "":
        if status not in ['0', '1', '-1']:
            return "Erro: Status deve ser 0 (ativo), 1 (inativo) ou -1 (excluído)."
        json_data['status'] = int(status)
    try:
        resposta = requests.patch(f'{endpoint}/user/update', json=json_data, timeout=5)
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
 
# insert csv
def insert_csv (vendor, file_path):
    if not path.isfile(file_path):
        return f"Erro: O arquivo '{file_path}' não existe."
    try:
        df = pd.read_csv(file_path)
        df = df.fillna('')
        df = df.map(lambda x: x.strip() if isinstance(x, str) else x)
    except Exception as e:
        return f"Erro ao ler o arquivo CSV: {e}"
    resp = ''
    for index, row in df.iterrows():
        item = {
            'username': row.get('username', ''),
            'name': row.get('name', ''),
            'email': row.get('email', ''),
            'whatsapp': row.get('whatsapp', '')
        }
        print(f"Processando linha {index + 1}: {item}")
        if not item['username'] or not item['name']:
            resp += f"Erro: Linha {index + 1} - Os campos 'username' e 'name' são obrigatórios. Dados: {item}\n"
            continue
        resp = insert(vendor, item['username'], item['name'], item['email'], item['whatsapp'])
        print(f"Resposta da API para linha {index + 1}: {resp}")
    return resp
