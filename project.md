## Planejamento 

### 0. Manutenções
* Revisar log billing e planner
* Verificar que no caso do plano pacote o número de items seja menor que a quantidade de aulas


No subprojeto billing:
- Caso a invoice tenha vindo de um sistema externo (api), ou seja, ExternalCode não seja nil, não possibilitar editar ou apagar pelo htmx do billing, mas sim somente pelo sistema externo que a gerou, ou seja pela api.
- Criar um serviço e a devida api para cancelamento de invoices, ou seja, alterando assim a data de cancelamento (CancellationDate)
- Fazer com que o botão de cancelamento do htmx de invoices faça uso dessa api de cancelamento, ao inves de utilizar a update com a data de cancelamento

No subprojeto planner:
- Voltar a permitir novamente a editar e apagar pré faturas que estejam faturadas (ReferenceInvoiceID não seja nil):
    - No caso de apagar a pré-fatura, deve-se enviar utilizar a api de cancelamento descrita no item anterior com a data atual
    - No caso de atualização da pré-fatura, deve-se cancelar a api (utilizando a api de cancelamento de billing) e criar uma nova com os dados atualizados
    




                    