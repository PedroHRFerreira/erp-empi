# Feedback global e alertas temáticos

## Objetivo

Substituir os diálogos nativos do navegador e unificar os diálogos personalizados em um sistema global de feedback coerente com a identidade visual do ERP, nos temas claro e escuro.

## Escopo aprovado

- Criar `useSystemFeedback` como interface global para confirmações, alertas, erros, informações e sucessos.
- Usar modal central para confirmações e ações de risco.
- Usar toast global para sucesso, informação e erro.
- Fechar toasts automaticamente após cinco segundos, pausando no hover e oferecendo fechamento manual.
- Preservar acessibilidade: foco inicial e restaurado, armadilha de foco, ESC, semântica ARIA e navegação por teclado.
- Substituir todos os usos de `window.alert` e `window.confirm` no frontend.
- Migrar os diálogos próprios de confirmação de pagamento e revogação para o mesmo padrão central.
- Respeitar tokens visuais existentes e funcionar nos temas light e dark.

## Arquitetura

- Um composable global mantém a fila de toasts e uma confirmação ativa.
- Um host único, montado no layout autenticado, renderiza os feedbacks via `Teleport`.
- Chamadores recebem `Promise<boolean>` para confirmações e não conhecem detalhes de apresentação.
- Ações assíncronas continuam pertencendo às páginas/stores; o host apenas coleta a decisão do usuário.

## Critérios de aceite

- Nenhum `window.alert`, `window.confirm` ou `window.prompt` permanece no frontend.
- Confirmações perigosas exigem ação explícita e nunca fecham por temporizador.
- Toasts são legíveis, não bloqueiam a interface e funcionam em desktop e celular.
- Modal e toast mantêm contraste e hierarquia em light e dark.
- Fluxos de clientes, recibos, recuperação, gastos, estoque, compras, pagamento e revogação preservam o comportamento atual.
- Testes frontend, typecheck/build e inspeção visual real passam ou têm qualquer bloqueio externo documentado.
