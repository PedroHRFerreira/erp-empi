# Estoque e compras unificados

## Objetivo

Transformar Estoque na única interface para cadastrar produtos, registrar reposições e consultar as informações de compra, eliminando a necessidade de uma tela separada de Compras.

## Escopo aprovado

- Manter uma linha por produto na tabela de estoque.
- Mostrar na linha a quantidade atual, o último fornecedor, o último custo, a data da última entrada, a situação financeira e ações úteis.
- Cadastrar a primeira entrada e suas parcelas junto com um novo produto.
- Disponibilizar `Adicionar estoque` para produtos existentes, sugerindo fornecedor e custo da última entrada e permitindo alterações.
- Somar cada reposição à quantidade atual e gerar suas próprias parcelas em Contas a pagar.
- Disponibilizar `Ver histórico` com todas as entradas do produto, incluindo data, fornecedor, quantidade, custo, total e situação das parcelas.
- Substituir o cancelamento por `Remover produto`, removendo o produto, suas entradas e parcelas ainda pendentes.
- Bloquear a remoção quando qualquer parcela relacionada ao produto já tiver sido paga, preservando Caixa, Gastos e o histórico financeiro.
- Remover Compras como destino de navegação separado, mantendo compatibilidade controlada para URLs antigas.
- Apresentar os produtos em tabela no desktop e cartões legíveis no celular.

## Regras funcionais

1. Toda criação ou reposição de produto representa uma entrada de estoque vinculada a fornecedor e pagamento.
2. Uma entrada pode ser à vista ou parcelada; a soma das parcelas deve ser igual ao total da entrada.
3. Fornecedor e custo anteriores são somente sugestões e permanecem editáveis.
4. A tabela é agregada por produto; o histórico preserva as entradas individuais.
5. A situação financeira resumida deriva das parcelas das entradas do produto.
6. Nenhum pagamento futuro afeta Caixa ou Gastos antes da quitação.
7. Uma parcela quitada continua aparecendo uma única vez como saída de estoque.
8. A exclusão é permitida apenas quando não existe parcela paga associada a nenhuma entrada do produto.
9. A exclusão permitida deve remover atomicamente o produto, suas compras/entradas e seus compromissos pendentes.

## Responsabilidades por camada

- Domínio Go: garantir criação/reposição, histórico por produto, exclusão atômica e bloqueio por pagamento realizado.
- Transporte/contratos: expor os dados agregados e históricos necessários sem transferir regras financeiras para o frontend.
- Nuxt: concentrar o fluxo em Estoque, apresentar sugestões e confirmações, e explicar claramente bloqueios retornados pelo domínio.

## Critérios de aceitação

- Um produto novo pode ser cadastrado com fornecedor, quantidade, custo e parcelamento no mesmo fluxo.
- Uma reposição preenche os dados da última entrada, aceita alterações, soma o estoque e cria novas parcelas.
- A tabela mostra dados úteis da última entrada e situação financeira sem depender da tela Compras.
- O histórico lista corretamente todas as entradas do produto.
- Produto sem pagamentos realizados pode ser removido com todos os vínculos pendentes.
- Produto com ao menos uma parcela paga não pode ser removido e apresenta mensagem explicativa.
- Contas a pagar, Caixa e Gastos mantêm sua consistência contábil.
- URLs antigas de Compras não deixam o usuário em uma tela quebrada.
- Desktop e celular não apresentam perda de conteúdo ou overflow horizontal.
