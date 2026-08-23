# Compras, contas a pagar, caixa multimeios e alertas

## Decisões aprovadas

- O cadastro do produto não gera gasto, pagamento ou saldo sem origem.
- Entradas de estoque são registradas por compras rastreáveis, com vários produtos e fornecedor em texto livre.
- A compra aumenta o estoque na confirmação e usa custo médio ponderado.
- Parcelas futuras não afetam Gastos nem Caixa; somente a quitação integral gera saída financeira.
- Pagamentos de estoque são saídas separadas, excluídas dos gastos operacionais e do cálculo duplicado do lucro.
- Compras ficam em Estoque; parcelas ficam em Contas a pagar; Caixa deixa de hospedar esses formulários.
- O Caixa resume Dinheiro, PIX, Débito e Crédito. Apenas Dinheiro exige caixa aberto e participa da conferência física.
- Boleto é uma forma prevista; na quitação registra-se o meio efetivo usado.
- Alertas permanecem até quitação/cancelamento e cobrem atraso, hoje, amanhã e antecipação em 30 dias.
- Antecipação usa apenas dinheiro físico e preserva o fundo inicial, priorizando vencimentos próximos.
- Compra confirmada só pode ser cancelada sem pagamentos e com unidades ainda disponíveis.
- Não haverá pagamento parcial, cadastro de fornecedores, contas bancárias ou notificações externas nesta versão.

## Critérios funcionais

1. Uma compra multiproduto atualiza estoque e cria parcelas numa única transação.
2. A soma das parcelas deve ser igual ao total dos itens.
3. Uma parcela só pode ser quitada uma vez.
4. Pagamento em dinheiro exige caixa aberto; outros meios podem ser registrados com o caixa fechado.
5. Gastos operacionais não incluem pagamentos de estoque.
6. Alertas são derivados das parcelas, sem uma tabela duplicada de notificações.
7. Datas financeiras seguem `America/Sao_Paulo`.
