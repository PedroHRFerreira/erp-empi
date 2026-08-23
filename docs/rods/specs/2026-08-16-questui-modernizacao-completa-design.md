# Modernização completa com QuestUI

## Objetivo

Modernizar todo o ERP EMPI com o QuestUI, mantendo a linguagem profissional de oficina e corrigindo falhas funcionais, de API, responsividade, acessibilidade e usabilidade encontradas durante uma auditoria completa.

## Decisões aprovadas

- Aplicar o tema fantasy do QuestUI sem metáforas de RPG nos textos do ERP.
- Corrigir causas em qualquer camada: Nuxt, BFF, API Go, domínio e persistência.
- Preservar, integrar e concluir o trabalho não commitado de Caixa.
- Testar com uma base PostgreSQL local descartável.
- Validar em Chromium real nas resoluções 1440x900, 1024x768 e 390x844.
- Atender WCAG 2.2 AA e corrigir todo problema reproduzível dentro da matriz definida.
- Manter uma suíte Playwright crítica no repositório e usar scripts temporários para a exploração ampla.
- Validar WhatsApp, NFS-e, clipboard, impressão e downloads somente até a saída local, sem envio externo.
- Usar Playwright MCP quando disponível e Playwright isolado como fallback autorizado.

## Arquitetura e experiência

- Preservar Browser -> Nuxt BFF -> API Go -> PostgreSQL e as responsabilidades atuais das camadas.
- Centralizar tokens, tipografia, estados, movimento e componentes recorrentes nas fundações SCSS existentes.
- Usar Cinzel em títulos e rótulos de UI, Spectral no conteúdo e Fira Code apenas em dados monoespaçados.
- Substituir confirmações nativas e feedback inconsistente por padrões acessíveis e reutilizáveis.
- Garantir navegação, formulários, tabelas e ações completas em desktop, tablet e celular.

## Aceite

- Login, métricas, metas, recibos, recuperação, clientes, gastos, caixa, estoque e perfil passam nos caminhos felizes e em erros relevantes.
- Testes Go, Vitest, typecheck, build Nuxt, Playwright e auditoria de acessibilidade ficam verdes.
- Não restam erros inesperados de console, chamadas de API quebradas ou defeitos reproduzíveis na matriz.
- Dois relatórios registram os testes executados e os ajustes realizados com suas justificativas.
