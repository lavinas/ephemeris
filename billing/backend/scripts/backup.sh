#!/bin/bash

# ==========================================
# CONFIGURAÇÕES DO BANCO E CONTAINER
# ==========================================
CONTAINER_NAME="billing_db"
DB_USER="root"
DB_NAME="ephemeris"
DB_PASSWORD="root"

# ==========================================
# CONFIGURAÇÃO DE DIRETÓRIO NO HOST (SUA MÁQUINA)
# ==========================================
# Salva na pasta /home/seu_usuario/backups/postgres
BACKUP_DIR="$HOME/backups/postgres"
DATA_HORA=$(date +"%Y%m%d_%H%M%S")
NOME_ARQUIVO="backup_${DB_NAME}_${DATA_HORA}.dump"
CAMINHO_FINAL="${BACKUP_DIR}/${NOME_ARQUIVO}"

# Garante que a pasta de destino existe na sua máquina de fora
mkdir -p "$BACKUP_DIR"

echo "🔄 Iniciando o backup do banco [${DB_NAME}] de dentro do container..."

# ==========================================
# EXECUÇÃO DO COMANDO MÁGICO
# ==========================================
# Explicando o truque: O 'docker exec -i' (sem o -t) injeta a senha e extrai os dados.
# O operador '>' faz o fluxo de dados do container escrever no disco da sua máquina física.
docker exec -i "$CONTAINER_NAME" env PGPASSWORD="$DB_PASSWORD" pg_dump \
    -U "$DB_USER" \
    -F c \
    -b \
    -v \
    "$DB_NAME" > "$CAMINHO_FINAL"

# ==========================================
# VALIDAÇÃO DO SUCESSO
# ==========================================
if [ $? -eq 0 ]; then
    echo "✅ Backup concluído com sucesso!"
    echo "📂 Salvo em de fora do container: $CAMINHO_FINAL"
    echo "⚖️  Tamanho do arquivo: $(du -sh "$CAMINHO_FINAL" | cut -f1)"
else
    echo "❌ Erro ao realizar o backup."
    # Remove o arquivo vazio caso tenha falhado
    rm -f "$CAMINHO_FINAL"
fi