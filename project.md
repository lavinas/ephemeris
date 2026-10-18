## Planejamento 

### 0. Manutenções
* Revisar log billing e planner
* Verificar que no caso do plano pacote o número de items seja menor que a quantidade de aulas


No subprojeto planner:
- No htmx de Pré-faturas e correspodente handler:
    - Não permitir (e não aprsentar os botões nas ações) apagar e editar pré-faturas caso a Pré-fatura já esteja Faturada (enviada para o billing)
    - Da mesma forma que existe embaixo do nome do cliente o indicativo de "Faturada", colocar pras não faturadas o indicativo de "Pendente" em outra cor propicia
    - Colocar um filtro (liga e deslisga) chamado Pendente, clicado como default, que apresentará apenas as pré-faturas que não foram fatrudas (Pendentes). Caso o botão seja desclicado, mostra todas.
- No htmx de Sessões, colocar um botão mais discreto para o botão de sincronismo em ações, que não se sobresaia aos demais botões

No subprojeto billing:
- No htmx de Faturas, alinhar o layout da tabela de faturas, pois as datas estão sendo apresentadas em duas linhas junto com o gif de feito....pode aproximar mais os Itens e Total das ações e dar mais espaço para os campos de datas



                    

                    