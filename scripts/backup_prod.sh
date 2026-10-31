#!/usr/bin/env bash
#
# backup_prod.sh
#
# Gera dumps dos bancos PostgreSQL ("billing", "planner" e "signup") e envia
# para uma pasta do Google Drive (via rclone).
#
# Pré-requisitos:
#   - pg_dump instalado
#   - rclone instalado e configurado com um remote chamado "gdrive"
#     autenticado na conta lavinas@gmail.com
#     (configure com: rclone config)
#
# Uso: ./backup_prod.sh

set -euo pipefail

# ---- Configurações gerais do banco ----
DB_HOST="localhost"
DB_USER="root"
DB_ENV="production" # "production" ou "development"
export PGPASSWORD="root"

# Lista de bancos no formato "nome_do_banco:porta"
DATABASES=(
  "billing:5432"
  "planner:5433"
  "signup:5434"
)

# ---- Configurações do Google Drive ----
RCLONE_REMOTE="gdrive"
DRIVE_FOLDER_ID="1OpV3fB6gPuKD1lSszczJ7BsoZ6Q5eN1F"

# ---- Configurações do backup ----
KEEP_LAST=3

# ---- Diretório temporário ----
BACKUP_DIR="$(mktemp -d)"
trap 'rm -rf "$BACKUP_DIR"' EXIT

TIMESTAMP="$(date +%Y%m%d_%H%M%S)"

backup_database() {
  local db_name="$1"
  local db_port="$2"
  local dump_file="${BACKUP_DIR}/${db_name}_${DB_ENV}_${TIMESTAMP}.dump"

  echo "============================================================"
  echo ">> Gerando dump do banco '${db_name}' em ${DB_HOST}:${db_port}..."
  pg_dump \
    --host="$DB_HOST" \
    --port="$db_port" \
    --username="$DB_USER" \
    --format=custom \
    --compress=9 \
    --file="$dump_file" \
    "$db_name"

  echo ">> Dump criado: $dump_file ($(du -h "$dump_file" | cut -f1))"

  echo ">> Enviando para o Google Drive (${RCLONE_REMOTE}, pasta ${DRIVE_FOLDER_ID})..."
  rclone copy "$dump_file" "${RCLONE_REMOTE}:" \
    --drive-root-folder-id="$DRIVE_FOLDER_ID" \
    --progress

  echo ">> Mantendo apenas os ${KEEP_LAST} backups mais recentes de '${db_name}' no Google Drive..."
  local files_to_delete
  files_to_delete=$(rclone lsf "${RCLONE_REMOTE}:" \
    --drive-root-folder-id="$DRIVE_FOLDER_ID" \
    --include "${db_name}_${DB_ENV}_*.dump" \
    | sort -r \
    | tail -n +$((KEEP_LAST + 1)) || true)

  if [ -n "$files_to_delete" ]; then
    while IFS= read -r file; do
      if [ -n "$file" ]; then
        echo ">> Removendo backup antigo: $file"
        rclone deletefile "${RCLONE_REMOTE}:$file" \
          --drive-root-folder-id="$DRIVE_FOLDER_ID"
      fi
    done <<< "$files_to_delete"
  else
    echo ">> Nenhum backup antigo de '${db_name}' para remover (total <= ${KEEP_LAST})."
  fi

  echo ">> Backup concluído com sucesso: ${db_name}_${DB_ENV}_${TIMESTAMP}.dump"
}

for entry in "${DATABASES[@]}"; do
  IFS=":" read -r db_name db_port <<< "$entry"
  backup_database "$db_name" "$db_port"
done

echo "============================================================"
echo ">> Todos os backups foram concluídos com sucesso!"