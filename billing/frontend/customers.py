import requests
import pandas as pd
from tabulate import tabulate
from requests.exceptions import ConnectionError, Timeout

# Configurações
endpoint = 'http://localhost:8080'
page = 1
page_size = 10

# get
def get(vendor, status, nickname, name, document, email, whatsapp):
    json_data = {'status': status,'page': page, 'page_size': page_size, 'vendor': vendor, 
                'nickname': nickname, 'name': name, 'document': document, 'email': email, 
                'whatsapp': whatsapp}
    # make API call with error handling
    try:
        resposta = requests.get(f'{endpoint}/customer/list', json=json_data, timeout=5)
    except ConnectionError as e:
        return f"Erro: A conexão foi recusada pelo servidor remoto. Detalhes: {e}"
    except Timeout as e:
        return f"Erro: A requisição excedeu o tempo limite estabelecido. {e}"
    except requests.exceptions.RequestException as e:
        return f"Ocorreu um erro genérico no requests: {e}"
    # verify response status code
    if resposta.status_code != 200:
        return f'Erro na chamada da API: {resposta.status_code}'
    # processing response data
    json_data = resposta.json()
    if 'customers' not in json_data or len(json_data['customers']) == 0:
        return 'Nenhum cliente encontrado.'
    # exibir clientes
    customers = json_data['customers']
    df = pd.DataFrame(customers)
    return tabulate(df, headers='keys', tablefmt='grid', showindex=False)

# insert
def insert(vendor, nickname, name, document, email, whatsapp):
    json_data = {'vendor': vendor, 'items': [{'nickname': nickname, 'name': name, 
                                              'document': document, 'email': email, 
                                              'whatsapp': whatsapp}]}
    try:
        resposta = requests.post(f'{endpoint}/customer/create', json=json_data, timeout=5)
    except ConnectionError as e:
        return f"Erro: A conexão foi recusada pelo servidor remoto. Detalhes: {e}"
    except Timeout as e:
        return f"Erro: A requisição excedeu o tempo limite estabelecido. {e}"
    except requests.exceptions.RequestException as e:
        return f"Ocorreu um erro genérico no requests: {e}"
    json_data = resposta.json()
    return f'{json_data["status"]} - {json_data["message"]}'
 