import psycopg2
from psycopg2 import errors

host = 'localhost'
port = 5432
database = 'ephemeris'
user = 'root'
password = 'root'
schema = 'billing'
database_url = f'postgresql://{user}:{password}@{host}:{port}/{database}'        

# Repository class to manage database connection and operations
class Repository:
    # Repository class to manage database connection and operations
    def __init__(self):
        self.connection = psycopg2.connect(database_url)
    # close the database connection
    def close(self):
        self.connection.close()
    def commit(self) -> str:
        self.connection.commit()
    def rollback(self) -> str:
        self.connection.rollback()
    # insert a new customer into the database and return the result
    def insert_customer(self, name, nickname, status, document, email, wapp) -> str:
        if self.connection is None:
            raise Exception("Database connection is not established")
        cursor = self.connection.cursor()
        cursor.execute(f"""
            INSERT INTO {schema}.customer (name, nickname, status, document, email, whatsapp)
            VALUES (%s, %s, %s, %s, %s, %s)
            RETURNING id
        """, (name, nickname, status, document, email, wapp))
        customer_id = cursor.fetchone()[0]
        return customer_id
    # find a customer by document and return the result
    def find_customer_by_document(self, document) -> tuple[str, str, str, str, str, str, str, str]:
        if self.connection is None:
            raise Exception("Database connection is not established")
        cursor = self.connection.cursor()
        cursor.execute(f"""
            SELECT id, name, nickname, status, email, whatsapp, created_at, updated_at
            FROM {schema}.customer
            WHERE document = %s
        """, (document,))
        return cursor.fetchone()
    # find a customer by nickname and return the result
    def find_customer_by_nickname(self, nickname) -> tuple[str, str, str, str, str, str, str, str]:
        if self.connection is None:
            raise Exception("Database connection is not established")
        cursor = self.connection.cursor()
        cursor.execute(f"""
            SELECT id, name, nickname, status, email, whatsapp, created_at, updated_at
            FROM {schema}.customer
            WHERE nickname = %s
        """, (nickname,))
        return cursor.fetchone()
    # find all active customers and return the result
    def find_active_customers(self) -> list[tuple[str, str, str, str, str]]:
        if self.connection is None:
            raise Exception("Database connection is not established")
        cursor = self.connection.cursor()
        cursor.execute(f"""
            SELECT id, name, nickname, document, email, whatsapp
            FROM {schema}.customer
            WHERE status = 'active'
        """)
        return cursor.fetchall()