- Gostaria de fazer um programa em golang que se faça interface com a plataforma pluggy por apis, que baixe os extratos financeiros das contas configuradas em uma determinada conta de aplicacattivo configurada no pluggy.
- Gostaria de salvar as informações de extrato de conta corrente, cartão de crédito e de aplicações financeiras da conta.
- Salvar a informação em banco de dados...no postgres do container accounting_db_dev, criei o database pluggy, para armazenar os dados coletados da aplicação
- O programa deve ser capaz de fazer requisições nas seguintes APIs:
    - bank_statements
    - credit_cards
    - investments
- O pluggy fornece a seguinte documentação para os agentes: https://mcp.pluggy.ai/mcp
- Por favor estruturar uma base relacional postgres para armazenar os dados baixados para consulta
- Salvar as scripts sql na pasta ./migration/pluggy.sql
- Gostaria de ter uma interface htmx para consultar as informações baixadas e também com algum botão para sincronizar os dados automaticamente.
- Tenho as seguintes credenciais para acesso a plataforma de desenvolvimento que poderia estar em arquivo de configuração.
    * Client ID: 1fe09f72-ca31-4139-904b-5fa9172767ce
    * Client Secret: VzBM_pWK1Z7hPtqu_SYfiXJoAhvIS4Cw0xZ1c5kSX1I
- Gostaria de desenvolver também uma api para receber do webhook os eventos gerados pelo pluggy. 
- Gostaria de expor esta api webhook da minha máquina que está dentro de uma rede interna atras de uma rede wi-fi caseira normal
- Estruturar o aplicativo da forma mais estruturada possível, porém simples. Este app será para fazer um piloto de interface cujos algoritmos serão refeitos dentro de um projeto mais estruturado
- O projeto fica dentro da pasta pluggy e pode criar as pastas necessárias para a estruturação do código.

- Adicionalmente consegue estrutura o docker para que eu possa rodar o aplicativo em um container golang e ter acesso a interface html bem como expondo o webhook (especificando como deve ser feito o expor)