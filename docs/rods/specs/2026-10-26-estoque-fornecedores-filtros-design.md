# Estoque por fornecedores e produtos e filtros de recibos

Plano aprovado pelo usuário em 26/10/2026.

## Comportamento aprovado

- Estoque apresenta dois cards: Fornecedores e Produtos.
- Produtos têm saldo unificado, independentemente do fornecedor; fornecedores exibem compras, produtos, quantidades, custos, parcelas e totais pagos/pendentes.
- Uma ação Editar reúne cadastro e reposição opcional; saldo atual e quantidade adicional são separados.
- Cadastro sugere produtos de todo o estoque, inclusive saldo zero, priorizando fornecedor. O usuário escolhe explicitamente repor um cadastro existente ou criar outro.
- Cadastro/edição com entrada e parcelas é atômico. Reposição preserva custo médio e gera novas contas, sem alterar pagamentos anteriores.
- Recibos têm busca por cliente/telefone/placa/serviço, período de criação, situação, forma de pagamento e tipo normal/rápido. Filtros são combinados antes da paginação e preservados após ações.
- Data final inclui o dia inteiro no fuso America/Sao_Paulo; formas de pagamento contemplam recibos mistos.
- Fornecedores usam nome e histórico existentes, sem cadastro fiscal ou contatos nesta entrega.

## Produção

Somente a correção administrativa do produto RESERVATORIO DO ÁGUA E TAMPA, custo 9100 centavos e revenda 13100 centavos, está autorizada. Localizar ID único e confirmar ausência de recibos e compras compartilhadas; produzir backup e prévia, ensaiar localmente, excluir todos os vínculos exclusivos em transação e conferir resumos. Não publicar as funcionalidades automaticamente. O usuário informa que não há vendas vinculadas.

### Revisão aprovada em 27/10/2026 após inspeção

A produção contém dois cadastros com o nome real RESERVATORIO DE ÁGUA E TAMPA e os mesmos preços, ambos vinculados a vendas pagas. O usuário identificou o cadastro SS PEÇAS (saldo 1) e escolheu expressamente **cancelar a reposição de R$ 131 e arquivar o produto, preservando venda e compra original**. Essa decisão substitui a exclusão integral inicialmente prevista. O cadastro ERIK permanece intacto.

A correção foi ensaiada em PostgreSQL local e executada em transação pelo script `internal/maintenance/correct_ss_stock.sql`, com prévia padrão e validação rígida dos IDs, valores, saldo e situações. Resultado conferido: produto inativo/saldo zero, reposição e parcela canceladas, compra original confirmada/pagamento de R$ 91 preservado com lançamento de caixa, recibo pago de R$ 1.241,49 e seus cinco itens preservados. Nenhum registro foi excluído. Backup de recuperação local em `/home/pedro/.local/share/empi-maintenance/2026-10-27-ss-pecas/` (fora do versionamento).

## Aceite

Validar em PostgreSQL e interface locais: criar 2 unidades, quitar compra, vender/receber 2 (saldo zero), repor 3 no mesmo cadastro, vender 1 (saldo 2), repor por outro fornecedor e conferir históricos. Cobrir edição sem entrada, sugestões, criação explícita, paginação, filtros combinados, falhas sem gravação parcial, estoque insuficiente e envio repetido. Inspecionar desktop 1440x900, tablet 1024x768 e celular 390x844, incluindo teclado e erros. Executar testes Go/frontend, integração PostgreSQL, tipagem e build; registrar evidências e lacunas. Preservar alteração pré-existente no estilo de AuthLayoutTemplate.
