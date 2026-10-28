# Validação de estoque, fornecedores e recibos — 27/10/2026

## Entrega

Estoque apresenta os acessos Fornecedores e Produtos. Produtos têm saldo unificado e uma única ação Editar com reposição opcional. O cadastro oferece sugestões com escolha explícita entre repor um cadastro existente e criar outro. Cadastro/edição, entrada e parcelas são gravados atomicamente; fornecedor permite somente reposição, preservando o cadastro atual no servidor.

Recibos têm busca por cliente, telefone com ou sem máscara, placa e serviço; período de criação no fuso de São Paulo; situação; forma de pagamento, incluindo componentes de pagamentos mistos; e tipo normal/rápido. Paginação e ações preservam os filtros, sem compartilhá-los com Recuperação.

Também foi corrigido o carregamento de produtos do recibo para contemplar todo o estoque, independentemente da página da listagem. Produtos arquivados permanecem no histórico do fornecedor, identificados e sem ação de reposição.

As interfaces novas são consultas individuais/busca de produtos, consultas derivadas de fornecedores e entrada opcional nos contratos de cadastro/edição. Nenhuma tabela de fornecedores foi criada; as regras financeiras continuam no backend. Links antigos de compras têm redirecionamento compatível.

## Ambiente e verificações

- Frontend: http://localhost:3000; API local: http://127.0.0.1:8080.
- PostgreSQL exclusivo: `empi_validation_20261026`, em `127.0.0.1:5432`. Os testes não usaram produção.
- Acesso criado apenas para esse banco de testes: CPF `52998224725`, senha `change-me-admin-password` (os mesmos valores de exemplo dos testes do projeto).
- Testes Go completos aprovados, incluindo integração PostgreSQL de rollback e reposições concorrentes e ensaio da correção administrativa em esquema isolado.
- Frontend: **106 testes aprovados** na revisão final para publicação; tipagem e build final aprovados; diferenças sem erros de whitespace.
- Chromium real: **6/6 cenários de estoque**, **6/6 cenários de filtros** e **6/6 verificações finais** de compatibilidade, venda fora da primeira página e histórico arquivado. As últimas três verificações de histórico repetem e ampliam o cenário de busca já executado.
- Viewports: desktop **1440×900**, tablet **1024×768**, celular **390×844**.
- Capturas inspecionadas diretamente, incluindo topo/final dos formulários, sugestões, erros e histórico. Sem overflow horizontal ou violações graves/críticas de acessibilidade nos estados avaliados.
- Após reiniciar a API com o código final, saúde, autenticação local, equivalência da busca por telefone formatado e consultas de estoque/fornecedores foram conferidas novamente.

## Ciclo completo conferido na interface e por consultas independentes

| Ação | Saldo esperado e encontrado | Conferência adicional |
| --- | ---: | --- |
| Criar produto com 2 unidades a R$ 91 | 2 | Uma entrada e uma parcela pendente de R$ 182 |
| Quitar por PIX | 2 | Uma única saída de R$ 182 no caixa e um gasto realizado |
| Vender e receber 2 unidades | 0 | Recibo pago e quantidade usada 2 |
| Repor 3 no mesmo cadastro | 3 | Nova entrada e parcela independente; pagamento anterior preservado |
| Vender e receber mais 1 | 2 | Mesmo cadastro; quantidade usada 3 |
| Comprar 2 de outro fornecedor a R$ 101 | 4 | Custo médio R$ 96; históricos e totais separados por fornecedor |
| Repor 1 pelo detalhe desse fornecedor | 5 | Fornecedor e cadastro bloqueados para edição; nova entrada no mesmo produto |
| Editar apenas a descrição | 5 | Sem novas compras, parcelas ou alterações de saldo |

Também foram aprovados: produto zerado nas sugestões; seleção por teclado; produto fora da primeira página para reposição e venda; criação explícita de outro cadastro; quantidade e soma de parcelas inválidas sem gravação; estoque insuficiente; clique repetido gerando um único envio; falha de persistência sem gravação parcial; filtros combinados, intervalo invertido, nenhum resultado, pagamentos mistos, quitação com redução da página e navegação por Recuperação.

## Correção em produção

A inspeção encontrou dois cadastros RESERVATORIO DE ÁGUA E TAMPA com custo R$ 91 e revenda R$ 131, ambos vinculados a vendas pagas. O usuário identificou **SS PEÇAS, saldo 1**, e revisou expressamente a ação para **cancelar a reposição de R$ 131 e arquivar o produto, preservando venda e compra original**.

Executado o script `internal/maintenance/correct_ss_stock.sql` em transação, após ensaio local e exportação dos registros originais. O arquivo transferido foi comparado por SHA-256 com o fonte ensaiado (`a26e23b1b99093c5e6e3ce8af7419ec0d2a9e81e11b8b05ebd675c445979c869`). Nenhum registro foi excluído.

Verificação independente após a execução:

- Produto SS PEÇAS: **inativo, saldo 0, usados 1**.
- Reposição de R$ 131 e sua parcela: **canceladas**.
- Compra original: **confirmada**; parcela original de R$ 91: **paga**, com lançamento de caixa preservado.
- Recibo original: **pago, R$ 1.241,49, cinco itens**.
- Cadastro ERIK: **preservado**, ativo, saldo 0 e usados 1.
- Recuperação: `/home/pedro/.local/share/empi-maintenance/2026-10-27-ss-pecas/` contém `before.json`, `after.json` e `recover.sql`, com acesso restrito e fora do versionamento.

O `recover.sql` também foi ensaiado com os dados originais em tabelas temporárias locais: restaurou saldo/atividade e situações anteriores, sem alterar tabelas reais do banco de testes.

As ferramentas temporárias foram removidas do serviço. Não houve deploy, reinício de serviço de produção nem mudança das restrições de rede do banco. As funcionalidades novas permanecem somente no ambiente local. A alteração pré-existente no estilo da autenticação foi preservada.

## Evidências visuais

- [Cards de estoque no desktop](/tmp/empi-stock-desktop-home.png).
- [Formulário e parcelas no desktop](/tmp/empi-stock-desktop-create-bottom.png).
- [Histórico pago e pendente do fornecedor](/tmp/empi-stock-desktop-supplier-paid-history-bottom.png).
- [Mensagem de validação no tablet](/tmp/empi-stock-tablet-validation-error-bottom.png).
- [Sugestão de produto zerado no celular](/tmp/empi-stock-mobile-zero-stock-autocomplete.png).
- [Produto após o ciclo completo no celular](/tmp/empi-stock-mobile-products-final.png).
- [Histórico de produto arquivado no celular](/tmp/empi-stock-mobile-supplier-archived-history-bottom.png).
- [Filtros de recibos no desktop](/tmp/receipt-filters-desktop.png), [tablet](/tmp/receipt-filters-tablet.png) e [celular](/tmp/receipt-filters-mobile.png).

Os arquivos de imagem ficam no diretório temporário do sistema, conforme a regra do projeto, sem incorporar imagens de teste ao versionamento.

## Complemento — filtro da tela Produtos

Adicionada busca por nome com Filtrar/Enter e Limpar filtro. A consulta alcança todo o catálogo no servidor, preserva o filtro na paginação e ao retornar da edição, reinicia na primeira página quando o nome muda e mostra quantidade de resultados e estado vazio. Respostas antigas não substituem uma busca mais recente.

Validação adicional aprovada: seis testes da store, tipagem e três testes reais de interface (desktop, tablet e celular), incluindo produto fora da primeira página, paginação filtrada, Enter, nenhum resultado e limpeza. Capturas desktop e celular inspecionadas, sem overflow da página.

Evidências: [desktop](/tmp/empi-product-filter-desktop.png), [tablet](/tmp/empi-product-filter-tablet.png), [celular](/tmp/empi-product-filter-mobile.png). Alteração apenas local nesta etapa.

## Revisão final para publicação autorizada

Após o pedido de publicação, executados novamente todos os testes Go com a integração PostgreSQL local habilitada, os 106 testes do frontend, a tipagem e o build. Os 15 cenários de navegador dos arquivos `stock-lifecycle`, `receipt-filters` e `product-filters` passaram juntos nos três tamanhos de tela. Capturas dos filtros, histórico e mensagens de erro foram inspecionadas novamente.

A publicação inclui o conjunto de estoque e recibos aprovado. A alteração pré-existente de `AuthLayoutTemplate/styles.module.scss` permanece local e não integra a versão enviada.
