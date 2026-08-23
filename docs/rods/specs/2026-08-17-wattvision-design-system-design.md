# WattVision — reformulação integral do design system

## Objetivo

Substituir o QuestUI pelo WattVision em toda a interface do ERP, transmitindo confiabilidade técnica e controle por meio de um dashboard analítico simples, adequado a laptops, tablets e celulares.

## Escopo aprovado

- Aplicar o novo sistema visual em todas as páginas, layouts, formulários, tabelas e estados da aplicação.
- Reorganizar a estrutura visual das telas sem alterar textos, dados, regras de negócio, APIs ou funcionalidades.
- Preservar e persistir a alternância entre temas escuro e claro.
- Validar autenticação, páginas principais, carregamento, erros, estados vazios e responsividade.

## Sistema visual

- Tema escuro: fundo técnico escuro, superfícies `#1E1E1E`, bordas `#2C2C2E` e texto de alto contraste sem aparência lúdica.
- Tema claro: cinzas claros de baixo brilho, sem branco puro, com os mesmos acentos e contraste equivalente.
- Dados e foco: ciano `#00E5FF`; sucesso e gradientes: `#30D158`/`#32D74B`; alertas e picos: `#FF453A`.
- Cards: raio de 16 px, padding de 20 px e borda sutil de 1 px.
- Tabelas: divisores sutis e hover de linha por mudança de superfície.
- Alertas: fundo vermelho escuro, texto vermelho e borda esquerda de 4 px.
- Tipografia: Inter 14 px no corpo e semibold 24 px nos títulos; Fira Code bold 32 px em KPIs e valores relevantes.
- Espaçamento em múltiplos de 8 px e estrutura baseada em grid de 12 colunas.
- Dashboard: três KPIs superiores; conteúdo principal em relação 8/4 quando houver gráfico e alertas; empilhamento progressivo em viewports menores.
- Iconografia: raios, tomadas e medidores; evitar lâmpadas genéricas.
- Movimento sutil e compatível com `prefers-reduced-motion`.

## Responsividade e acessibilidade

- Desktop, tablet e celular devem manter hierarquia, legibilidade e ações acessíveis.
- Tabelas devem continuar utilizáveis horizontalmente em telas estreitas.
- Controles interativos devem ter foco visível e alvos adequados ao toque.
- Manter conformidade WCAG 2.2 AA.

## Critérios de aceitação

1. Nenhum resquício visual do QuestUI permanece na interface exibida.
2. Tema escuro e claro funcionam, persistem e não usam branco puro como fundo.
3. Todos os fluxos existentes mantêm o comportamento e os contratos atuais.
4. Estados normal, vazio, carregando e erro usam a linguagem visual WattVision.
5. Typecheck, testes unitários, build, testes Go/API e Playwright full-stack passam.
6. As telas são inspecionadas em Chromium nos viewports desktop (1440×900), tablet (1024×768) e celular (390×844).
