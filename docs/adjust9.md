No subprojeto planning:
- Alterei as tabelas invoice e invoice_item, deslocando as colunas period_start e period_end da tabela invoice para invoice_item, fazendo assim que cada item tenha seu período específico.
- Desta forma, é necessário ajustar dominio, serviços, dto, etc para se adapatar a esta nova caracteristica.
- A geração dos planos em invoices se mantém como antes, porém registrando o período nos itens ao inves da tabela principal.
- No htmx das pré-faturas, retirar o período que aparece embaixo do tipo plano
- Criar serviços, apis e ações no htmx para adicionar um item de fatura avulso para adicionar ou descontar serviços de um determinado período na pré-fatura, registrando o serviço, o período, o preço e a quantidade. No caso de adicão, o valor final será positivo e caso desconto o valor será negativo. Permitir também que se possa apagar a inclusão deste item.
- A criação ou o apagar deste novo item deve também integrar-se ao banco de horas, adicionando ou subtraindo horas conforme o caso.
- Também estes itens deverão ser integrados no envio da invoice ao billing.
- Manter o padrão do projeto, mantendo as melhores praticas da arquitetura hexagonal e de codigo e arqchitetura limpos.
