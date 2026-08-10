import requests
from requests.exceptions import ConnectionError, Timeout
from os import path as os_path
from base64 import b64decode, b64encode

# Configurações
endpoint = 'http://localhost:8080'
page = 1
page_size = 1000

# send
def generate (vendor, start_date, end_date, emission_date, path):
  json_data = {'vendor': vendor, 'invoice_start_date': start_date, 'invoice_end_date': end_date, 'emission_date': emission_date}
  try:
    resposta = requests.post(f'{endpoint}/tax/generate', json=json_data, timeout=5)
  except ConnectionError as e:
    return f"Erro: A conexão foi recusada pelo servidor remoto. Detalhes: {e}"
  except Timeout as e:
    return f"Erro: A requisição excedeu o tempo limite estabelecido. {e}"
  except requests.exceptions.RequestException as e:
    return f"Ocorreu um erro genérico no requests: {e}"
  if resposta.status_code != 200:
    return f"Erro: A requisição retornou um status code diferente de 200. Status code: {resposta.status_code} => {resposta.text}"
  resp = resposta.json()
  file_path = os_path.join(path, resp['document_name'])
  file_bin = b64decode(resp['document_base64'])  
  with open(file_path, 'wb') as f:
    f.write(file_bin)
  return f'{resp["status"]} - {resp["message"]} - id: {resp["emission_id"]} - qtde: {resp["emission_quantity"]} - valor: {resp["emission_amount"]}'

# receive
def clear (vendor, id, source):
  with open(source, "rb") as file:
    base64_bytes = b64encode(file.read())
    base64_string = base64_bytes.decode("utf-8")
  json_data = {'vendor': vendor, 'emission_id': id, 'source': base64_string}
  try:
    resposta = requests.post(f'{endpoint}/tax/clear', json=json_data, timeout=5)
  except ConnectionError as e:
    return f"Erro: A conexão foi recusada pelo servidor remoto. Detalhes: {e}"
  except Timeout as e:
    return f"Erro: A requisição excedeu o tempo limite estabelecido. {e}"
  except requests.exceptions.RequestException as e:
    return f"Ocorreu um erro genérico no requests: {e}"
  json_data = resposta.json()
  return f'{json_data["http_code"]} - {json_data["status"]} - {json_data["message"]}'