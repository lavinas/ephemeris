import xml.etree.ElementTree as ET
from datetime import datetime
import csv

# function to extract transactions from OFX files
def ofx_extract_transaction(files, init_date=None, end_date=None):
    if init_date:
        init_date = datetime.strptime(init_date, "%d/%m/%Y")
    if end_date:
        end_date = datetime.strptime(end_date, "%d/%m/%Y")
    stat = []
    for file in files:
        with open(file, 'r', encoding='utf-8') as f:
            conteudo_ofx = f.read()
        # Encontra onde começa a estrutura XML para ignorar o cabeçalho SGML
        pos_ofx = conteudo_ofx.find('<OFX>')
        if pos_ofx == -1:
            print("Estrutura <OFX> não encontrada.")
            return
        xml_puro = conteudo_ofx[pos_ofx:]
        # Faz o parse do XML
        raiz = ET.fromstring(xml_puro)
        # Itera sobre todas as transações (<STMTTRN>)
        for stmttrn in raiz.findall('.//STMTTRN'):
            tipo = stmttrn.find('TRNTYPE').text
            data_bruta = stmttrn.find('DTPOSTED').text[:8] # Pega AAAAMMDD
            valor = float(stmttrn.find('TRNAMT').text)
            descricao = stmttrn.find('MEMO').text.strip()
            # Converte a data bruta para objeto datetime
            data_obj = datetime.strptime(data_bruta, "%Y%m%d")

            # Filtra pelas datas inicial e final, se fornecidas
            if init_date and data_obj < init_date:
                continue
            if end_date and data_obj > end_date:
                continue

            # Formata a data para DD/MM/AAAA
            data_formatada = data_obj.strftime("%d/%m/%Y")
            stat.append({
                "data": data_formatada,
                "tipo": tipo,
                "valor": valor,
                "descricao": descricao
            })
    stat.sort(key=lambda x: (datetime.strptime(x["data"], "%d/%m/%Y"), x["descricao"]))
    return stat

# function to extract transactions from CSV files
def csv_extract_transaction(file_path, init_date=None, end_date=None):
    if init_date:
        init_date = datetime.strptime(init_date, "%d/%m/%Y")
    if end_date:
        end_date = datetime.strptime(end_date, "%d/%m/%Y")
    stat = []
    with open(file_path, 'r', encoding='utf-8') as f:
        leitor = csv.reader(f)
        next(leitor)  # Skip the header row
        for line in leitor:    
            try:
                data_obj = datetime.strptime(line[0], "%d/%m/%Y")
                valor_str = line[4].replace(',', '.')
                valor_str = valor_str.replace('"', '').strip()
                valor = float(valor_str)
            except ValueError:
                continue            
            # Filtra pelas datas inicial e final, se fornecidas
            if init_date and data_obj < init_date:
                continue
            if end_date and data_obj > end_date:
                continue
            # append the transaction to the list
            stat.append({
                "data": data_obj.strftime("%d/%m/%Y"),
                "conta": line[1].strip(),
                "tipo": line[2].strip(),
                "valor": valor,
                "descricao": line[3].strip()
            })
            
            
    # Sort the transactions by date before returning
    stat.sort(key=lambda x: (datetime.strptime(x["data"], "%d/%m/%Y"), x["descricao"]))
    return stat

# diff 
def transaction_diff(transactions1, transactions2):
    set1 = set((t['data'], t['valor'], t['descricao'][:8].upper()) for t in transactions1)
    set2 = set((t['data'], t['valor'], t['descricao'][:8].upper()) for t in transactions2)
    diff1 = [t for t in set1 if t not in set2]
    diff2 = [t for t in set2 if t not in set1]
    diff1.sort(key=lambda x: (datetime.strptime(x[0], "%d/%m/%Y"), x[2]))
    diff2.sort(key=lambda x: (datetime.strptime(x[0], "%d/%m/%Y"), x[2]))
    return diff1, diff2