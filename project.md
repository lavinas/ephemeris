## Planejamento 

### 0. Manutenções
* Revisar log billing e planner
* erificar que no caso do plano pacote o número de items seja menor que a quantidade de aulas


No subprojeto billing:
- Na tabela invoice:
    * foi adicionado o campo vendor_id
    * foi adicionado o campo external_code
    * foi retirada o campo whatsapp_sent_date
    * foi retirado o campo whatsapp_receipt_date
    * foi substituido o campo email_sent_date por sent_date
    * foi substituido o campo email_receipt_date por receipt_date
- Os campos whatsapp_sent_date e whatsapp_receipt_date não estavam sendo utilizados, então devem ser retirados do dominio
- Os campos email_sent_date e email_receipt_date, mudar para sent_date e receipt_date respectivamente no dominio, serviços, dto, etc.
- O campo external_code deve estar no dominio e também deve ser recebido pelo dto e tratado pelo serviços (create, list, update), etc para ser armazenado se recebido
- O campo vendor_id estava também na tabela customer_id e pode estar duplicado na tabela invoice, porém por uma questão de performance vai se repetir...então nos serviço de criação da invoice, providenciar a inserção do campo vendor_id (que já vem no dto) também no dominio/tabela invoice
- Deve-se fazer a validação do external_code para que não seja permitido duplicidade para a tupla (vendor_id, external_code), ou seja, o mesmo vendedor não pode ter o mesmo external_code....caso tenha, deve-se retornar um erro de duplicadade na api de criação ou alteração.
- Na api de criação da invoice /invoice/api/create, no dto e serviço correspondente, retornar na chamada além dos dados base o id da invoice criada
- Também tratar o external_code no htmx de gestão de invoices, adicionando na inserção, alteração e listagem do campo (na listagem colocar entre envio nota e itens & total, com o nome da campo na tabela 'Ext. Cod.')


No subprojeto planner:
- Fazer um serviço que irá faturar as pre faturas, isto é, irá enviar para o subsistema billing as invoices selecionadas para criação das faturas:
    * Para isto, utilizar as apis que foram criadas no subprojeto billing /invoice/api/create enviando
        - vendor: invoice.plan.customer.vendor.nickname
        - customer: invoice.plan.customer.nickname
        - invoicing: invoice.invoiceDate
        - due: invoice.dueDate
        - invoiceItems: invoice.items: [
            - description: invoice.items.description
            - price: invoice.items.price
            - quantity: invoice.items.invoicedQuantity
        ]
    * O id recebido da chamada da api, armazenar no campo ReferenceInvoiceID de cada pre fatura.
    * Para o processo de chamada da api do billing, como se tratando de um serviço externo, preparar um adapter que interagirá com este sistema externo, podendo assim substituir a chamada para um outro serviço sem alterar o core da aplicação, no melhor modelo hexagonal/ddd.
    * As invoices que já tem uma chamada prévea, ou seja, com o ReferenceInvoiceID da aplicação prenchido, não deve ser enviado, retornando assim um bad request no serviço do planner.
    * Ainda no serviço de listagem da pré-fatura, criar na interface (dto) um status de faturamento (Invoiced bool), true no caso do campo ReferenceInvoiceID diferente de nulo e false no caso contrário
    * No htmx da pré-fatura, deve-se inserir nas ações da listagem de pré-faturas um botão para faturar a invoice, ou seja, chamar o serviço externo de faturamento.
    * Ainda no htmx de pré-fatura deve-se 
        

     



                    

                    