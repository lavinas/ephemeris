#!/usr/bin/env bash
# ==============================================================================
# Helper Script: Expor Webhook da máquina local (Wi-Fi caseira) para a Internet
# ==============================================================================

PORT=${1:-8085}

echo "=================================================================="
echo " Iniciando túnel para o Webhook da Pluggy na porta $PORT"
echo "=================================================================="

if command -v lt >/dev/null 2>&1; then
    echo ">> Usando localtunnel (lt)..."
    echo ">> Copie a URL gerada abaixo (ex: https://xxx.loca.lt)"
    echo ">> A sua URL final de webhook será: https://<URL_GERADA>/api/webhook"
    echo ">> Em seguida, registre no painel Webhook da aplicação ou no painel Pluggy."
    echo ""
    lt --port "$PORT"
elif command -v cloudflared >/dev/null 2>&1; then
    echo ">> Usando cloudflared tunnel..."
    cloudflared tunnel --url "http://localhost:$PORT"
elif command -v ngrok >/dev/null 2>&1; then
    echo ">> Usando ngrok..."
    ngrok http "$PORT"
else
    echo "Nenhuma ferramenta de túnel encontrada (lt, cloudflared ou ngrok)."
    echo "Para instalar o localtunnel: npm install -g localtunnel"
    exit 1
fi

