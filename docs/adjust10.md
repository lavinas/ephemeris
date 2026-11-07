Como resposta ao ultimo comanda do agente, quanto ao ajuste da emissão da nfe para o leiout 2 da prefeitura de são paulo, estruturar o novo xml do leiaute 2 e atualizar a rotina rps, assumindo os seguintes parametros e alterações adicionais:
- Quanto ao código nbs para uso no leiaute 2:
    * alterei a tabela billing.invoice_item para conter o código nbs que será usado para envio.
    * Ajustar dominio, serviços e adapters para estes ajustes.
    * Caso o nbs não esteja presente (ou seja nul), não permitir o envio de nota fiscal
    * Alterar as apis e htmx de inclusão e alteração da invoice para receber este parametro de forma não obrigatório. Caso não informado, assumir nulo.
    * ainda no htmx, não apresentar o botão de emissão de nota fiscal, caso o nbs seja nulo.
    * No serviço de notificar pagamento, caso o nbs seja nulo, não fazer o passo da emissão, apenas o registro do pagamento e o envio do recibo.
- Quanto a parametrização do simples nacional para os campos ibs/cbs:
    * Assumir em configuração que o modelo padrão é o código de situação tributária específico do Simples.
    * Posteriormente haverá uma configuração no vendor para isto, mas por enquanto deixe o psdrão configurado
- Quanto ao endereço:
    * Para as pessoas fisicas, não é necessario o endereço, apenas para pessoa juridica, desta forma, ainda não ireo configurar endereço para os clientes, dado que todos são pf.
    * Para as pessoas físicas (documento menor  a 14 digitos), emitir a nota com os dados de endereço em branco
    * Para as pessoas juridicas (documento maiior que 14 digitos), não permitir a emissão.
    * No htmx retirar o botão de emissão para o caso de pessoa juridica e também não fazer o passo de emissão da nota no serviço de pagamento (apenas para pessoa juridica).
- Adicionalmente quanto ao código nbs no subsistema planner:
    * adicionei o campo nbs na tabela planner.service
    * ajuster dominio, serviços e adapters para estes ajustes.
    * No serviço de emissão da invoice (envio para billing), considerar este campos ao enviar/criar invoice item.

É necessária mais alguma informação, caso contrário seguir com os ajustes.
