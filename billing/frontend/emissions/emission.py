import requests
from requests.exceptions import ConnectionError, Timeout

# Configurações
endpoint = 'http://localhost:8080'
page = 1
page_size = 1000

# emit
def send (vendor, start_date, end_date, emission_date):
  json_data = {'vendor': vendor, 'invoice_start_date': start_date, 'invoice_end_date': end_date, 'emission_date': emission_date}
  try:
    resposta = requests.post(f'{endpoint}/emission/send', json=json_data, timeout=5)
  except ConnectionError as e:
    return f"Erro: A conexão foi recusada pelo servidor remoto. Detalhes: {e}"
  except Timeout as e:
    return f"Erro: A requisição excedeu o tempo limite estabelecido. {e}"
  except requests.exceptions.RequestException as e:
    return f"Ocorreu um erro genérico no requests: {e}"
  json_data = resposta.json()
  return f'{json_data["status"]} - {json_data["message"]}'