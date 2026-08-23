# Validação completa da aplicação — 16/08/2026

## Escopo e ambiente

A aplicação foi compilada e executada em uma stack Docker descartável, isolada da stack de desenvolvimento, com frontend em `127.0.0.1:13000`, API em `127.0.0.1:18080` e PostgreSQL em `127.0.0.1:55432`. O navegador foi controlado por Playwright com Chromium em ambiente isolado, pois o MCP Playwright não estava disponível nesta sessão.

Foram validados desktop (1440 × 900), tablet (1024 × 768) e mobile (390 × 844). A análise automática de acessibilidade usou axe-core com regras WCAG 2 A/AA, 2.1 A/AA e 2.2 AA, bloqueando violações sérias ou críticas.

## Jornadas executadas

| Jornada | Resultado | Evidência principal |
| --- | --- | --- |
| Login inválido | Aprovado | Mensagem de credenciais inválidas anunciada com `role="alert"` |
| Login administrativo | Aprovado | Redirecionamento para Métricas e sessão autenticada |
| Métricas | Aprovado | Resumo e tabelas carregados sem erro de console/API |
| Metas | Aprovado | Tela, progresso e controles renderizados nos três viewports |
| Recibos | Aprovado | Listagem e criação pelo wizard de Recibo rápido, valor de R$ 120,00 |
| Recuperação | Aprovado | Lista carregada e acessível |
| Clientes | Aprovado | Lista carregada, tabela navegável por teclado |
| Gastos | Aprovado | Gasto criado, persistido e refletido no resumo financeiro |
| Caixa | Aprovado | Abertura diária, resumo operacional e lançamento de saída do gasto |
| Estoque | Aprovado | Produto criado e localizado na listagem em desktop/tablet/mobile |
| Perfil | Aprovado | Conteúdo autenticado carregado |
| Logout e proteção de rota | Aprovado | Sessão encerrada e `/profile` redirecionada para `/login` |
| Responsividade | Aprovado | Sem overflow horizontal; menu móvel funcional e acessível |
| Console e API | Aprovado | Nenhum erro de console, page error ou resposta HTTP 5xx nas rotas percorridas |
| Acessibilidade automática | Aprovado | Nenhuma violação séria/crítica nas páginas percorridas |

As jornadas mutáveis de recibo, gasto e caixa foram executadas uma vez no projeto desktop para evitar duplicação de efeitos na base compartilhada. Navegação, acessibilidade, estoque e autenticação foram repetidos nos três projetos.

## Bugs encontrados e retestados

1. Documento sem idioma e título: corrigidos com metadados globais em português.
2. Estrutura inválida de listas de métricas (`dl`): corrigida para semântica válida.
3. Tabelas largas sem região focável: receberam rótulo e navegação por teclado.
4. Feedback de login não anunciado: recebeu região de alerta.
5. Sidebar sem fluxo móvel robusto: substituída por drawer com botão nomeado, estados aberto/fechado e layout tablet compacto.
6. Divergência de hidratação em datas: datas e horários agora usam explicitamente `America/Sao_Paulo` no servidor e cliente.
7. Cadastro de gasto falhava sem contexto operacional: confirmada a exigência de caixa aberto; a jornada agora informa a pré-condição e mantém o lançamento transacional.
8. Inconsistências visuais e tipográficas: unificadas pelos tokens QuestUI.

## Gates técnicos

- `go test ./...`: aprovado.
- `pnpm lint` (`nuxt typecheck`): aprovado.
- `pnpm test`: 4 arquivos e 19 testes aprovados.
- Build de produção Nuxt/Docker: aprovado.
- Playwright: 15 testes aprovados e 6 ignorados intencionalmente; duração final de 44,6 s.
- `git diff --check`: aprovado.

## Limites da validação

Não houve envio real de e-mail, WhatsApp, impressão física ou integração com provedores externos. Os geradores locais de recibo/PDF permanecem cobertos por testes unitários e foram compilados, mas destinos externos exigem credenciais e ambientes próprios. A auditoria axe automatizada complementa, mas não substitui, uma avaliação manual completa com leitores de tela reais.
