# Autosugestão no recibo e histórico detalhado do estoque

## Objetivo

Reduzir a digitação repetitiva na criação de recibos e transformar o histórico de entradas de um produto em uma página própria, clara e responsiva.

## Escopo aprovado

- Buscar clientes existentes por nome ou telefone durante a criação de um recibo.
- Aplicar sugestões somente após seleção explícita do usuário.
- Preencher nome, telefone, modelo, ano e placa com base no recibo mais recente do cliente selecionado.
- Manter todos os dados sugeridos editáveis e submetidos às validações normais do recibo.
- Informar a origem da sugestão e permitir limpar a seleção.
- Não copiar serviços, produtos, valores, descontos, pagamento, gastos ou observações do atendimento anterior.
- Manter a margem configurada no perfil como sugestão segura no cadastro de produto novo.
- Manter último fornecedor e custo como sugestões na reposição de estoque.
- Fazer `Ver histórico` navegar para uma página própria do produto.
- Mostrar resumo do produto, entradas e parcelas na página de histórico.

## Regras funcionais

1. Digitar nome ou telefone não altera silenciosamente o restante do formulário.
2. Somente a seleção de uma sugestão dispara o carregamento do histórico do cliente.
3. O recibo mais recente é determinado pela data de criação mais nova disponível no detalhe do cliente.
4. A ausência de recibo anterior mantém os campos do veículo vazios e não bloqueia o cadastro.
5. Alterar nome ou telefone depois da seleção desfaz a associação visual da sugestão, sem apagar campos que o usuário já validou.
6. A tela de histórico do estoque é somente leitura.
7. Entradas canceladas permanecem identificadas, e cada entrada apresenta fornecedor, data, quantidade, custo unitário e total.
8. Parcelas apresentam número, valor, vencimento, forma prevista, forma efetiva quando houver e situação.

## Responsabilidades

- Stores existentes: carregar clientes, detalhe do cliente, produtos e compras.
- Componentes do recibo: controlar busca, seleção explícita e feedback de dados sugeridos.
- Página de histórico: compor informações já retornadas pelos contratos de compras, sem persistência adicional.

## Critérios de aceitação

- A busca encontra clientes por nome ou telefone e funciona com teclado e ponteiro.
- Selecionar um cliente preenche os cinco campos aprovados usando o recibo mais recente.
- O usuário consegue alterar qualquer sugestão antes de avançar.
- Dados específicos do atendimento atual permanecem vazios.
- A tela de histórico abre por produto e exibe entradas e parcelas completas.
- Estados de carregamento, vazio, erro e produto inexistente são compreensíveis.
- Desktop e celular não apresentam overflow horizontal.
