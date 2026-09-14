#!/usr/bin/env bash
#
# backup_prod.sh
#
# Gera um dump do banco PostgreSQL "billing" e envia para uma pasta do
# Google Drive (via rclone).
#
# Pré-requisitos:
#   - pg_dump instalado
#   - rclone instalado e configurado com um remote chamado "gdrive"
#     autenticado na conta lavinas@gmail.com
#     (configure com: rclone config)
#
# Uso: ./backup_prod.sh

set -euo pipefail

# ---- Configurações do banco ----
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="root"
DB_NAME="billing"
DB_ENV="production" # "production" ou "development"
export PGPASSWORD="root"

# ---- Configurações do Google Drive ----
RCLONE_REMOTE="gdrive"
DRIVE_FOLDER_ID="1OpV3fB6gPuKD1lSszczJ7BsoZ6Q5eN1F"

# ---- Configurações do backup ----


# ---- Diretório temporário ----
BACKUP_DIR="$(mktemp -d)"
trap 'rm -rf "$BACKUP_DIR"' EXIT

TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
DUMP_FILE="${BACKUP_DIR}/${DB_NAME}_${DB_ENV}_${TIMESTAMP}.dump"

echo ">> Gerando dump do banco '${DB_NAME}' em ${DB_HOST}:${DB_PORT}..."
pg_dump \
  --host="$DB_HOST" \
  --port="$DB_PORT" \
  --username="$DB_USER" \
  --format=custom \
  --compress=9 \
  --file="$DUMP_FILE" \
  "$DB_NAME"

echo ">> Dump criado: $DUMP_FILE ($(du -h "$DUMP_FILE" | cut -f1))"

echo ">> Enviando para o Google Drive (${RCLONE_REMOTE}, pasta ${DRIVE_FOLDER_ID})..."
rclone copy "$DUMP_FILE" "${RCLONE_REMOTE}:" \
  --drive-root-folder-id="$DRIVE_FOLDER_ID" \
  --progress

echo ">> Backup concluído com sucesso: ${DB_NAME}_${TIMESTAMP}.dump"