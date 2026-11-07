Sub-sitema accounting:
- O subsistema accounting será responsável por gerenciar os dados financeiros e contábeis de um vendor.
- Utilizeramos a linguagem golang e sua estrutura de microserviços.
- Pode-se utilizar como consulta outro subsistema como billing e planner. 
- Também utilizaremos a arquitetura hexagonal, bem como as melhores práticas de clean code e SOLID.
- A estrutura de diretórios já está criada seguindo a arquitetura hexagonal e a estrutura de diretórios nativa do golang:
    - cmd - código de ponto de entrada da aplicação
    - internal
        - adapters - camada responsável por receber requisições externas, como HTTP e filas, iteralçao com banco de dados e outros serviços externos
        - domain - camada responsável por conter as entidades e regras de negócio do sistema. 
        - dto - camada responsável por conter as Data Transfer Objects do sistema.
        - ports - camada responsável por conter as interfaces do sistema.
        - service - camada responsável por conter a lógica de negócio do sistema.
    - web
        - static - arquivos web estáticos
        - templates - templates html e htmx
    - migration - estrutura de dados proposto
    - docker - configurações da aplicação para o docker
    - docs - onde estão os documentos de especificações que serão utilizadas para a criação do projeto e para os agentes de IA. 
    - scripts - onde estarão arquivos shell e ações em banco de dados
    - log - onde os serviços logarão suas informações
    - tmp - onde estarão os arquivos do air
- Este sistema terá interação com outros subsistemas como billing e planner através de apis e filas caso necessário
- Também irá expor apis e também expor interfaces web atraavés de htmx
- As macro funcionalidades iniciais serão:
    - banking - buscar extratos bancários através de apis ou fazendo upload de arquivos ofx de vários bancos (inicialmente Santander e Banco do Brasil)
    - accounting - terá a contabilidade gerencial, ou seja, irá receber os lançamentos financeiros do banking e irá realizar a contabilidade gerencial. Também poderá receber lançamentos contábeis manuais através de apis ou do htmx
- O accouting irá interagir com o subsistema billing onde ele poderá receber os farturamentos previstos para registro e também poderá dar baixa no billing de pagamentos reconhecidos.
- O banking irá ter também jobs que de tempos em tempos, de forma configurada, irá buscar dados em uma determinada api ou arquivos ofx em uma conta de whatsapp (inicialmente apis)
- Iremos iniciar com o banking e depois as demais funcionalidades
- No subdiretório migration, existe o arquivo com as tabelas previstas do banking para apreciação
- Iremos utilizar o banco de dados e a aplicação em dockers.
- No ambiente de desenvolvimento, utilizaremos o air (verificar as estrutura de billing e planner)

Conseguiria criar a estrutura inicial já com o air e com o docker, já importando a estrutura de migração?



