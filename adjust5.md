No subprojeto 'planner':
- No htmx de sessoes, na funcionalidade de inserir uma sessao ou alterar uma sessão, no campo de escolha de serviços:
    - Sugerir os servilos de acordo com o cliente (nickname) da seguinte forma:
        - Caso o cliente tenha um ou mais planos ativos (isto é, se a data atual esteja dentro do período do plano (data incial e data final)), apresentar apenas os serviços que estão disponíveis naquele plano, ordenados pelo nome do plano.
        - Caso o cliente não tenha nenhum plano ativo, apresentar todos os serviços disponíveis.
    - Ajustar o backend para se ter os serviços, apis, etc dentro do padrão do projeto e dentro do modelo de arquitetura hexagonal e dentro das boas práticas de código e arquitetura limpos