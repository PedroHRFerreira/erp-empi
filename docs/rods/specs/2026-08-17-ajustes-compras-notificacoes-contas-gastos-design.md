# Compras, notificações, contas a pagar e gastos realizados

## Objetivo

Tornar o ciclo de compras e pagamentos compreensível de ponta a ponta: registrar mercadoria, programar parcelas, acompanhar vencimentos, quitar compromissos e analisar somente as saídas efetivamente realizadas.

## Escopo aprovado

- Ao selecionar um produto na compra, preencher o custo unitário cadastrado no estoque, permitindo edição manual antes da confirmação.
- Ao abrir a compra por `?productId=...`, preencher esse custo automaticamente após carregar o estoque.
- Permitir remover parcelas geradas. Ao remover, redistribuir o total entre as restantes, preservando datas e meios previstos; a última parcela não pode ser removida.
- Redesenhar Compras de estoque e Contas a pagar com KPIs e filtros, tabela profissional no desktop e cartões compactos no celular.
- Substituir o pop-up de notificações por um drawer lateral direito, com backdrop, rolagem, fechamento explícito/ESC e informações financeiras completas.
- Transformar Gastos no painel de saídas realizadas, com filtros Todos, Operacionais e Estoque. Parcelas pendentes ficam exclusivamente em Contas a pagar e não entram nos totais realizados.
- Manter a separação contábil: compra de estoque quitada é saída de estoque, não gasto operacional; a visão Todos soma ambas sem duplicidade.

## Regras funcionais

1. O custo cadastrado é um valor inicial da linha de compra, não um campo bloqueado.
2. A soma das parcelas deve permanecer igual ao total da compra após geração, edição ou remoção.
3. Remover uma parcela atualiza a quantidade exibida e distribui os centavos restantes deterministicamente.
4. Nenhuma parcela futura afeta Caixa ou Gastos antes da quitação.
5. Uma parcela quitada aparece uma única vez em Gastos, com origem Estoque, na data do pagamento efetivo.
6. Gastos operacionais continuam editáveis/removíveis; saídas de estoque derivadas de pagamentos são somente leitura.
7. Ações de pagamento continuam aceitando Dinheiro, PIX, Débito e Crédito; Boleto permanece apenas como forma prevista.
8. O drawer deriva os alertas de parcelas e não cria uma segunda tabela de notificações.

## APIs e padrões existentes permitidos

- `GET /stock/purchases`, `POST /stock/purchases`, `DELETE /stock/purchases/:id` por `usePurchasesStore`.
- `GET /payables`, `GET /payables/alerts` e `POST /payables/:id/pay` por `usePurchasesStore`.
- `IStockItem.costCents` e `stock.items` por `useStockStore` para o custo inicial.
- `PurchaseDraft` com itens e parcelas em `frontend/app/stores/usePurchasesStore.ts`.
- `CashService.CreatePurchase`, `PendingInstallments`, `PayInstallment` e `PayableAlerts` como regras de domínio atuais.
- `GET /expenses` para CRUD operacional e `GET /financial/summary` para o resumo atual.
- Tokens, cards, alertas, tipografia e responsividade do design WattVision já definidos no sistema.

## Antipadrões proibidos

- Não criar gasto ou saída ao cadastrar produto ou confirmar compra ainda não paga.
- Não inserir pagamentos de estoque na tabela de gastos operacionais.
- Não somar `Expense` e seu `CashEntry` correspondente duas vezes.
- Não permitir editar/excluir pelo painel analítico uma saída de estoque derivada de parcela.
- Não manter o painel de notificações preso à posição do menu lateral.
- Não comprimir tabelas largas no celular; usar apresentação em cartões.
- Não inventar novos métodos de pagamento ou parâmetros fora dos contratos existentes.

## Plano de implementação

### Fase 0 — documentação e contratos

**Implementar**

- Confirmar as assinaturas e relações em `internal/domain/entities/cash.go`, `internal/domain/cash/services/cash_service.go`, `internal/domain/financial/services/financial_service.go`, handlers, rotas e `frontend/server/contracts/types.ts`.
- Definir o DTO de saídas realizadas antes de alterar frontend ou persistência.

**Referências**

- `docs/rods/specs/2026-08-17-compras-contas-pagar-caixa-design.md`.
- `internal/domain/cash/services/cash_service_test.go` para a garantia de que somente a quitação gera saída.
- `internal/domain/expenses/services/expense_service_test.go` para o resumo financeiro.

**Verificação**

- Lista documentada de campos, status, meios de pagamento e filtros aceitos.
- Confirmação de que o feed unificado não exige nova tabela.

### Fase 1 — formulário de compra e parcelas

**Implementar**

- Em `frontend/app/pages/stock/purchases/new.vue`, copiar `costCents` do produto selecionado para `costInput` usando `formatCentsAsCurrency`.
- Aplicar o mesmo comportamento ao produto preselecionado pela query.
- Criar helper único de divisão em centavos e reutilizá-lo na geração e na remoção de parcelas.
- Adicionar ação acessível “Remover parcela”; preservar vencimento e forma das parcelas restantes, atualizar `installmentCount` e impedir remoção da última.
- Ajustar a grade para acomodar a ação em desktop e mobile.

**Verificação**

- Produto selecionado mostra o custo cadastrado e ainda aceita alteração manual.
- Duas ou mais parcelas podem ser reduzidas sem quebrar a soma.
- Casos com resto de centavos preservam exatamente o total.
- O POST mantém o contrato atual.

### Fase 2 — histórico de compras de estoque

**Implementar**

- Em `frontend/app/pages/stock/purchases/index.vue`, incluir KPIs: compras confirmadas no período, total comprado, saldo pendente e parcelas atrasadas.
- Adicionar busca por fornecedor e filtros por período/situação.
- Melhorar hierarquia das linhas: fornecedor/data, quantidade/produtos, progresso de parcelas, total e status.
- Trocar o ícone isolado por ação textual ou menu identificado, mantendo confirmação e regras de cancelamento.
- Renderizar tabela em desktop e cartões em telas pequenas, com estados de carregamento, vazio e erro.

**Verificação**

- Filtros não alteram os dados persistidos.
- Progresso pago/total e saldos batem com as parcelas retornadas.
- Nenhum overflow horizontal em tablet/celular.

### Fase 3 — drawer de notificações

**Implementar**

- Extrair o conteúdo de notificações de `AppSidebar` para um componente dedicado ou uma região independente do layout.
- Abrir drawer fixo pela direita com backdrop, cabeçalho, botão fechar, rolagem e largura responsiva.
- Exibir severidade, fornecedor, número da parcela, vencimento, valor e link para Contas a pagar.
- Atualizar alertas ao abrir, fechar ao navegar, clicar no backdrop ou pressionar ESC; controlar foco e atributos `aria-expanded`/`aria-controls`.
- Preservar o contador no menu e no cabeçalho móvel.

**Referências**

- `frontend/app/components/organisms/AppSidebar/Index.vue`.
- `frontend/app/components/organisms/AppSidebar/styles.module.scss`.
- `IPayableAlert` e `usePurchasesStore.loadAlerts(forceRefresh)`.

**Verificação**

- Drawer não cobre nem desloca ações da página de forma imprevisível.
- Navegação por teclado, ESC, foco e leitores de tela funcionam.
- Estados sem alerta e com muitos alertas têm layout estável.

### Fase 4 — contas a pagar

**Implementar**

- Em `frontend/app/pages/payables.vue`, criar KPIs de total aberto, vencido, vence em breve e antecipável.
- Unificar alertas repetidos em uma faixa compacta e filtrável.
- Adicionar filtros por situação, fornecedor e período.
- Reorganizar cada compromisso com prioridade visual para valor e vencimento, deixando forma prevista e forma efetiva claramente distintas.
- Tornar a confirmação de pagamento explícita quanto a valor, fornecedor e meio escolhido.
- Usar tabela no desktop e cartões no celular.

**Verificação**

- “Marcar paga” atualiza lista, KPIs, alertas, Caixa e o futuro feed de Gastos.
- Pagamento em dinheiro preserva a exigência de caixa aberto; outros meios funcionam com caixa fechado.
- Status atrasada/hoje/amanhã/antecipável mantêm semântica e contraste.

### Fase 5 — Gastos como painel de saídas realizadas

**Implementar**

- Criar no domínio financeiro um feed somente leitura que combine:
  - `expenses` como origem `operational`;
  - `cash_entries` de `kind = stock_payment` como origem `stock`.
- Filtrar por período e origem `all|operational|stock`, paginar e ordenar pela data efetiva.
- Expor endpoint dedicado, preferencialmente `GET /financial/expenses`, sem sobrecarregar o CRUD `/expenses`.
- Criar contrato `IRealizedExpense` com identificação, origem, descrição, categoria, valor, data, meio, fornecedor/parcela opcionais e capacidade de edição.
- Ampliar o resumo com totais operacionais, estoque pago e total realizado, sem contabilização duplicada.
- Atualizar `useExpensesStore`, `ExpensesTemplate`, `ExpensesTable` e `FinancialSummaryGrid` com abas Todos/Operacionais/Estoque, KPIs, distribuição e coluna de origem/meio.
- Permitir editar/remover somente registros operacionais; para estoque, oferecer navegação à compra/conta relacionada.
- Invalidar feed e resumo após quitar uma parcela.

**Verificação**

- Compra pendente não aparece em Gastos.
- Após quitar uma de duas parcelas, apenas o valor quitado aparece em Estoque e em Todos.
- Operacionais aparecem em Todos e Operacionais, nunca em Estoque.
- O total Todos é a soma das origens sem duplicidade.
- Projeções futuras permanecem em Contas a pagar.

### Fase 6 — testes e validação visual full-stack

**Implementar**

- Testes Go para origem/período, pendência, quitação única, não duplicidade e totais.
- Testes frontend para preenchimento do custo, redistribuição de parcelas, filtros e estados do drawer.
- Atualizar `frontend/e2e/critical.spec.ts` com a jornada completa: produto → compra parcelada → ausência em Gastos → quitação → presença apenas do valor pago.
- Cobrir as rotas Compras, Nova compra, Contas a pagar e Gastos em desktop, tablet e celular, nos temas claro e escuro.

**Comandos de verificação**

- `go test ./...`
- typecheck e testes unitários definidos no `frontend/package.json`.
- suíte Playwright configurada no projeto.
- verificação visual real, incluindo overflow, contraste, foco, ESC e ausência de sobreposição.

## Critérios de aceite finais

- O custo unitário aparece automaticamente ao selecionar ou preselecionar produto.
- Parcelas podem ser removidas com redistribuição exata e nunca ficam com soma inválida.
- Compras e Contas a pagar são legíveis e acionáveis em desktop e celular.
- Notificações abrem em drawer direito acessível e não cobrem o menu.
- Gastos mostra claramente o que já saiu, separado em Operacionais e Estoque.
- Nenhuma conta futura é apresentada como gasto realizado.
- Testes backend, frontend, E2E, acessibilidade e inspeção visual passam.
