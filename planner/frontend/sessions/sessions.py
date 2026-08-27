import requests
import pandas as pd
from tabulate import tabulate
from requests.exceptions import ConnectionError, Timeout
from os import path as os_path
from base64 import b64decode
import matplotlib.pyplot as plt


# Configurações
endpoint = 'http://localhost:8083/api/session'
page = 1
page_size = 1000

# Get the list of sessions
def get(nickname, date_start, date_end, minutes, service, status, comments):  
    params = {
        "page": page,
        "page_size": page_size,
    }
    if nickname != "":
        params["nickname"] = nickname
    if date_start != "":
        params["date_start"] = date_start
    if date_end != "":
        params["date_end"] = date_end
    if minutes != "":
        params["minutes"] = minutes
    if service != "":
        params["service"] = service
    if status != "":
        params["status"] = status
    if comments != "":
        params["comments"] = comments
    try:
        response = requests.get(f"{endpoint}/list", json=params, timeout=10)
        response = response.json()
    except (ConnectionError, Timeout) as e:
        return(f"Error connecting to the server: {e}")
    except requests.exceptions.RequestException as e:
        return f"An error occurred with the request: {e}"
    if response.get("http_code") != 200:
        return f"Error: Received status code {response.get('http_code')} with message: {response.get('message')}"

    if not response.get("sessions"):
        return "No sessions found."
    df = pd.DataFrame(response["sessions"])
    df = df.fillna('-')
    # df = df.sort_values(by=['customer', 'invoicing'])
    return tabulate(df, headers='keys', tablefmt='grid', showindex=False)


# create creates a new session with the provided details.
def create(nickname, date, minutes, service, status, comments):
    payload = {
        "nickname": nickname,
        "date": date,
        "minutes": minutes,
        "service": service,
        "status": status,
        "comments": comments
    }
    try:
        response = requests.post(f"{endpoint}/create", json=payload, timeout=10)
        response = response.json()
    except (ConnectionError, Timeout) as e:
        return(f"Error connecting to the server: {e}")
    except requests.exceptions.RequestException as e:
        return f"An error occurred with the request: {e}"
    if response.get("http_code") != 200:
        return f"Error: Received status code {response.get('http_code') } with message: {response.get('message')}"
    return f"Session created successfully with ID: {response.get('session_id')}"

# delete deletes a session with the provided session ID.
def delete(session_id):
    try:
        response = requests.delete(f"{endpoint}/delete", json={"id": session_id}, timeout=10)
        response = response.json()
    except (ConnectionError, Timeout) as e:
        return(f"Error connecting to the server: {e}")
    except requests.exceptions.RequestException as e:
        return f"An error occurred with the request: {e}"
    if response.get("http_code") != 200:
        return f"Error: Received status code {response.get('http_code') } with message: {response.get('message')}"
    return f"Session with ID: {session_id} deleted successfully."