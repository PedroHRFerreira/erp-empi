# Parcelamento e histórico de gastos

## Objetivo

Permitir que todo gasto seja lançado com uma ou mais parcelas, acompanhado em Contas a pagar e reconhecido financeiramente apenas quando cada parcela for quitada.

## Fluxo aprovado

- O cadastro de gasto oferece quantidade de parcelas, primeiro vencimento, valores e datas ajustáveis, seguindo a experiência do cadastro de compras de estoque.
- Inclusive um gasto em uma parcela nasce pendente e é encaminhado para Contas a pagar.
- A quitação acontece em Contas a pagar, com forma e data de pagamento.
- Somente parcelas quitadas compõem Gastos operacionais e afetam o caixa.
- Compras de estoque permanecem separadas de Gastos operacionais, evitando dupla contabilização.

## Histórico e navegação

- Todo gasto possui a rota `/expenses/:id/history`.
- A tela mostra os dados do lançamento, parcelas, vencimentos, pagamentos e situação.
- A tabela de gastos usa a ação "Ver histórico" para todos os registros, substituindo a navegação inválida "Ver compra".
- Gastos legados sem parcelas continuam acessíveis e exibem explicitamente a ausência de histórico parcelado.

## Edição e exclusão

- Qualquer parcela paga bloqueia a edição e a exclusão do gasto.
- Sem parcelas pagas, alterações em valor, quantidade ou primeiro vencimento recalculam todas as parcelas pendentes.
- Excluir um gasto sem pagamentos remove também suas parcelas pendentes.

## Critérios de aceite

- O total das parcelas deve ser exatamente igual ao valor do gasto.
- Parcelas pendentes aparecem em Contas a pagar sem alterar Gastos operacionais nem caixa.
- Ao quitar uma parcela, somente seu valor passa a integrar Gastos operacionais e a saída de caixa correspondente.
- O histórico reflete corretamente parcelas pendentes, vencidas e pagas.
- Gastos com pagamentos não podem ser alterados nem excluídos.
- A ação de histórico nunca aponta para uma rota inexistente.
- Testes cobrem criação, redistribuição, quitação, reconhecimento financeiro, bloqueios e navegação.
