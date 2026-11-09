# Pluggy Open Finance - Aplicação Piloto em Go

Aplicação em **Golang** integrada com a plataforma **Pluggy Open Finance**, desenvolvida para coletar, sincronizar e consolidar extratos de conta corrente, faturas de cartão de crédito e carteiras de investimento em banco de dados relacional **PostgreSQL** (banco `pluggy` no container `accounting_db_dev`).

A aplicação inclui uma interface web interativa em **HTMX + Tailwind CSS** com widget **Pluggy Connect**, sincronização automática em tempo real e receptor de **Webhooks** preparado para exposição pública a partir de rede caseira.

---

## 🚀 Funcionalidades Principais

1. **APIs Pluggy Integradas**:
   - **Bank Statements**: Consulta e baixa extratos de contas correntes e poupanças (`/accounts`, `/transactions`).
   - **Credit Cards**: Consulta contas de cartão de crédito, limites disponíveis, faturas fechadas/abertas (`/bills`) e lançamentos na fatura.
   - **Investments**: Consulta carteira de aplicações financeiras (renda fixa, fundos, tesouro direto, ações) e histórico de movimentações (`/investments`, `/investments/{id}/transactions`).
2. **PostgreSQL Relacional**:
   - Script SQL estruturado e versionado em [`./migration/pluggy.sql`](file:///home/lavinas/projects/ephemeris/pluggy/migration/pluggy.sql).
   - Tabelas normalizadas com suporte a dados enriquecidos (`merchant`, `CNPJ`, `payment_data`, etc.).
   - Views otimizadas para consulta rápida: `vw_bank_statements`, `vw_credit_card_statements` e `vw_portfolio_summary`.
3. **Interface Web HTMX**:
   - **Dashboard**: Resumo patrimonial consolidado (saldo em contas, total gasto em cartões, total investido, instituições ativas) e últimas transações.
   - **Extratos Bancários**: Filtros por conta e busca instantânea com paginação via HTMX.
   - **Cartões de Crédito**: Cards visuais com dados de bandeira, limite, fatura atual, vencimento e detalhamento de lançamentos.
   - **Investimentos**: Detalhamento da carteira com taxas, prazos de vencimento e modal com movimentações.
   - **Conexões (Items)**: Listagem de bancos conectados, status da sincronização e botões para forçar sync ou desconectar.
   - **Webhooks & Túnel**: Guia de exposição de rede caseira e auditoria em tempo real dos payloads recebidos.
4. **Widget Pluggy Connect Embutido**:
   - Botão **"Conectar Conta"** que gera o `connectToken` e abre o modal oficial do Pluggy para conectar instituições em ambiente Sandbox (ex: Banco Sandbox, Itaú Sandbox, Nubank Sandbox, XP Sandbox).
5. **Receptor de Webhooks**:
   - Endpoint `POST /api/webhook` para escutar eventos disparados pela Pluggy (`item/created`, `item/updated`, `transactions/created`, etc.) e atualizar o banco automaticamente em background.

---

## 📁 Estrutura do Projeto

```text
pluggy/
├── agent.md                        # Requisitos e especificações do piloto
├── config.yaml                     # Arquivo de configuração (credenciais Pluggy e DB)
├── config.example.yaml             # Exemplo limpo para versionamento
├── go.mod / go.sum                 # Módulo Go e dependências (pgx/v5, yaml.v3)
├── migration/
│   └── pluggy.sql                  # DDL de criação das tabelas, índices e views
├── cmd/
│   └── server/
│       └── main.go                 # Ponto de entrada da aplicação
├── internal/
│   ├── config/config.go            # Carregamento de configurações e variáveis de ambiente
│   ├── db/
│   │   ├── db.go                   # Conexão e repositório PostgreSQL (upserts e queries)
│   │   └── models.go               # Structs de domínio correspondentes às tabelas
│   ├── pluggy/
│   │   ├── client.go               # Cliente HTTP Pluggy (Auth, Items, Accounts, Txs, Bills, Inv, Webhooks)
│   │   └── types.go                # DTOs da API Pluggy
│   ├── service/
│   │   └── syncer.go               # Orquestrador de sincronização e mapeamento de dados
│   └── handler/
│       ├── handlers.go             # Roteador HTTP Go 1.22+
│       ├── web.go                  # Renderização de views e fragmentos HTMX
│       ├── api.go                  # Endpoints REST e integrações (connect-token, sync, etc.)
│       └── webhook.go              # Listener de webhooks com processamento assíncrono
├── templates/                      # Templates HTML com Tailwind e HTMX
│   ├── layout.html                 # Layout base com navbar e Pluggy Connect JS
│   ├── dashboard.html              # Visão geral e KPIs
│   ├── statements.html             # Tabela e filtros de extrato bancário
│   ├── credit_cards.html           # Cartões, limites, faturas e lançamentos
│   ├── investments.html            # Carteira e modal de movimentações
│   ├── connections.html            # Gestão das instituições conectadas
│   └── webhooks.html               # Configuração do túnel e histórico de eventos
└── scripts/
    └── tunnel.sh                   # Script para expor o webhook local para a internet
```

---

## ⚙️ Configuração (`config.yaml`)

As configurações de porta, banco de dados e credenciais da Pluggy estão centralizadas em `config.yaml`:

```yaml
server:
  port: 8085

pluggy:
  base_url: "https://api.pluggy.ai"
  client_id: "1fe09f72-ca31-4139-904b-5fa9172767ce"
  client_secret: "VzBM_pWK1Z7hPtqu_SYfiXJoAhvIS4Cw0xZ1c5kSX1I"
  webhook_url: ""

database:
  host: "127.0.0.1"
  port: 5435
  user: "root"
  password: "root"
  dbname: "pluggy"
  sslmode: "disable"
```

*(Todas as opções também podem ser sobrescritas por variáveis de ambiente como `PLUGGY_CLIENT_ID`, `DB_HOST`, `DB_PORT`, etc.)*

---

## 🐳 Como Executar a Aplicação via Docker

A aplicação conta com Dockerfile de desenvolvimento (`docker/Dockerfile.dev` com **Air** para hot-reload automático) e de produção (`docker/Dockerfile.prod` otimizado em múltiplos estágios).

### 1. Ambiente de Desenvolvimento (com Hot Reload)

Na pasta `pluggy`:

```bash
docker compose -f docker-compose.dev.yml up --build -d
```

- **Logs do container**: `docker logs -f pluggy_api_dev`
- **Parar container**: `docker compose -f docker-compose.dev.yml down`
- O container conecta-se diretamente à rede Docker `accounting_net-accounting-dev` para falar com o PostgreSQL (`accounting_db:5432`) no banco `pluggy`.

### 2. Ambiente de Produção

```bash
docker compose -f docker-compose.prod.yml up --build -d
```

---

## 🛠️ Como Executar Nativamente (Sem Docker)

### 1. Iniciar o servidor Go

Na pasta `pluggy`:

```bash
cd pluggy
go run ./cmd/server
```

Ou compilar o binário:

```bash
go build -o bin/server ./cmd/server
./bin/server
```

O servidor iniciará na porta **8085** (tanto no Docker quanto nativo):
- Interface Web HTMX: **[http://localhost:8085](http://localhost:8085)**
- Endpoint Webhook: **`http://localhost:8085/api/webhook`**

---

## 🌐 Como Expor o Webhook (Com ou Sem Docker em Rede Wi-Fi Caseira)

Como a porta **8085** do container está mapeada para a porta **8085 do seu host** (`ports: - "8085:8085"`), expor o webhook funciona exatamente da mesma forma simples e direta com Docker ou nativo:

### Passo a Passo da Exposição:

1. **Inicie o container ou a aplicação nativa** na porta 8085.
2. **Abra o túnel seguro** a partir da sua máquina:

```bash
./scripts/tunnel.sh 8085
```

O comando (utilizando `localtunnel`, já instalado) exibirá a URL pública:
```text
your url is: https://exemplo-pluggy.loca.lt
```

3. **URL final do Webhook**:
```text
https://exemplo-pluggy.loca.lt/api/webhook
```

4. **Registrar na Pluggy**:
Acesse **[http://localhost:8085/webhooks](http://localhost:8085/webhooks)**, cole a URL pública gerada no formulário e clique em **"Registrar na Pluggy"**. A Pluggy enviará os eventos (`item/updated`, `transactions/created`, etc.) diretamente para o seu container!
2. No formulário **"Registrar Webhook na Plataforma Pluggy"**, cole a URL pública gerada (ex: `https://neat-badger-42.loca.lt/api/webhook`).
3. Clique em **"Registrar na Pluggy"**. A aplicação fará a requisição `POST /webhooks` diretamente na API da Pluggy.

---

## 🏦 Como Testar a Conexão de Contas (Sandbox)

1. Acesse o painel em **[http://localhost:8085](http://localhost:8085)**.
2. Clique no botão **"Conectar Conta"** no topo da página.
3. O modal oficial do **Pluggy Connect** será aberto com conectores do tipo **Sandbox** habilitados.
4. Escolha **"Banco Sandbox"** (ou use as credenciais de teste pré-preenchidas fornecidas pelo Pluggy).
5. Conclua a conexão. O widget chamará o callback `onSuccess` e nosso backend automaticamente:
   - Salvará a instituição em `pluggy_item`.
   - Baixará as contas bancárias e cartões de crédito em `pluggy_account`.
   - Baixará as transações bancárias e de cartão em `pluggy_transaction`.
   - Baixará faturas de cartão em `pluggy_credit_card_bill`.
   - Baixará os investimentos em `pluggy_investment` e operações em `pluggy_investment_transaction`.
6. A página atualizará exibindo os extratos e saldos em tempo real!

