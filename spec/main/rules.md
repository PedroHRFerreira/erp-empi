# Regras duráveis da branch main

- Quando um design Figma existir, sempre seguir seu padrão responsivo.
- Quando um design Figma existir ou o projeto for frontend, validar o resultado diretamente em navegador real com a skill `visual-check` antes de concluir.
- Se `visual-check` não completar uma interação, um browser MCP como Playwright MCP pode ser usado como último recurso, mantendo o fluxo nas skills e as ferramentas MCP primitivas.
- Usar WattVision como design system de todo o ERP: aparência técnica de dashboard analítico, sem elementos lúdicos, cores pastel ou fundos brancos puros.
- Oferecer temas escuro e claro persistidos; no tema claro, usar cinzas de baixo brilho e preservar os mesmos acentos e contraste técnico.
- Usar Inter no corpo e nos títulos e Fira Code em KPIs e valores relevantes.
- Usar ciano `#00E5FF` para dados e foco, verde `#30D158`/`#32D74B` para sucesso e gradientes e vermelho `#FF453A` para alertas e picos.
- Usar grid de 12 colunas, espaçamento em múltiplos de 8 px, cards com raio de 16 px e tabelas com divisores sutis e hover por mudança de superfície.
- Preferir ícones de raio, tomada ou medidor e evitar ícones genéricos de lâmpada.
- Validar todos os fluxos afetados em desktop, tablet e celular e atender WCAG 2.2 AA.
- Preferir transições sutis e respeitar `prefers-reduced-motion`; reservar cores de alerta para estados que realmente exigem atenção.
