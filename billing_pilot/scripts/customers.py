import repository
import pandas as pd

# Script to create customers in the database
def create_customer(name, nickname, document, email, wapp) -> str:
    try:
        repo = repository.Repository()
        client = repo.find_customer_by_document(document)
        if client is not None:
            return f"Cliente com documento {document} já existe"
        client = repo.find_customer_by_nickname(nickname)
        if client is not None:
            return f"Cliente com apelido {nickname} já existe"
        id = repo.insert_customer(name, nickname, 'active', document, email, wapp)
        repo.commit()
        return "Cliente criado com sucesso com ID {}".format(id)
    except Exception as e:
        repo.rollback()
        return f"Erro ao criar cliente: {e}"
    finally:
        repo.close()
# find actives customers in the database
def find_active_customers() -> pd.DataFrame:
    try:
        repo = repository.Repository()
        customers = repo.find_active_customers()
        return pd.DataFrame(customers, columns=['id', 'nome', 'apelido', 'documento', 'email', 'whatsapp'])
    except Exception as e:
        return f"Erro ao buscar clientes ativos: {e}"
    finally:
        repo.close()