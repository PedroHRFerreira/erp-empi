# Migração segura dos dados legados de produção

## Objetivo

Preparar a aplicação nova para preservar integralmente os dados da versão anterior, sem criar movimentos retroativos no caixa e sem permitir operação sobre uma migração incompleta.

## Estado confirmado de produção

A versão atual possui apenas usuários, produtos, recibos, itens de recibo, gastos e metas mensais. Não existem tabelas de sessões de caixa, movimentos de caixa, compras de estoque, itens de compra ou parcelas a pagar.

## Regras aprovadas

### Gastos legados

- Cada gasto ativo recebe uma única parcela quitada.
- Valor e data da quitação usam `amount_cents` e `spent_at` originais.
- A forma de pagamento é `legacy`, apresentada como "Forma não registrada — legado".
- Não é criada sessão nem entrada de caixa retroativa.
- O gasto continua compondo os relatórios financeiros históricos.
- Gastos arquivados permanecem arquivados e não recebem parcela ativa.

### Estoque legado

- Cada produto recebe uma compra histórica própria.
- A quantidade original é `quantity + used_quantity`.
- O custo total histórico é `cost_cents × quantidade original`.
- A compra recebe uma parcela quitada com forma `legacy`.
- Não é criada entrada de caixa retroativa.
- Quantidade disponível, quantidade usada, custo, markup, preço e estado ativo permanecem inalterados.

### Recibos legados

- Status, valores e `paid_at` permanecem inalterados.
- Recibos pagos continuam aparecendo nos relatórios pela data original.
- Não são criadas sessões ou entradas de caixa retroativas.

### Dados novos

- Após a migração, novos gastos e compras seguem normalmente por Contas a pagar.
- Somente quitações novas geram movimentos de caixa.

## Segurança e execução

- A migração é versionada, transacional e idempotente.
- Uma tabela de controle impede execução duplicada.
- Preflight valida esquema, referências, quantidades e valores antes de escrever.
- Pós-validação compara contagens e totais; qualquer divergência causa rollback e bloqueia a inicialização.
- O `AutoMigrate` não deve depender de dados parcialmente convertidos.
- O ensaio usa uma restauração local isolada do backup de produção.
- Nenhuma escrita em produção será feita durante desenvolvimento ou ensaio.
- O PostgreSQL pago do Render fornece a camada de recuperação recente por PITR.
- Antes de cada deploy importante, um export lógico é criado no painel do Render e baixado para armazenamento local durável.
- O pré-deploy exige esse arquivo, valida `pg_restore --list`, tamanho e SHA-256 e bloqueia a migração se qualquer verificação falhar.
- Uma cópia externa diária no Cloudflare R2 será criada pelo Cron Job dedicado do Render, com retenção diária e mensal configurada no bucket.

## Critérios de aceite

- Todos os registros preexistentes continuam acessíveis.
- Totais históricos de receitas e gastos permanecem iguais antes e depois.
- Nenhum registro legado cria movimento ou sessão de caixa.
- Cada gasto ativo possui exatamente uma parcela legada paga.
- Cada produto possui exatamente uma compra e parcela histórica paga.
- Saldos e custos atuais de estoque não mudam.
- Executar a migração novamente não altera contagens ou totais.
- Uma falha intermediária não deixa alterações parciais.
- Há relatório de ensaio e roteiro de backup, deploy, validação e rollback.
