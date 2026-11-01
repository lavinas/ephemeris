No subprojeto billing:
- Os serviços de criação e alteração de invoices via api ou htmx não permite atualmente que se passe um valor total para a invoice, ignorando valores na linha de itens. 
- Permitir que se envie um valor total na criação e alteração da invoice que caso seja enviado, sobrescreva o valor calculado dos itens.
- Caso o valor total não seja enviado, o valor calculado dos itens será usado.
- Tirar as validações de que o preço ou quantidade devem ser maior que zero nos itens....a quanitidade pode ser maior ou igual a zero e o preço pode ser menor, igaual ou maior que zero.

No subprojeto planner:
- Na funcionalidade de envio da fatura para o billing, enviar o valor total da invoice na chamada da api de billing

