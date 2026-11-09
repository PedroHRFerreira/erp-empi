# Apresentação pública do ERP EMPI

Desenho aprovado pelo usuário em 09/11/2026.

## Escopo

- Substituir somente o destaque `stickman-fight` pelo Gist técnico sobre baixa duplicada de estoque, preservando os outros cinco destaques.
- Adicionar uma seção curta de notas técnicas ao README do perfil, com título descritivo e link para o Gist.
- Atualizar a descrição do repositório ERP EMPI e reorganizar seu README com apresentação visual, fluxos, arquitetura, execução, testes e documentação.
- Incluir até três capturas das telas reais de métricas, recibos e estoque, com dados fictícios em ambiente isolado.

## Apresentação

Manter a identidade atual do projeto, títulos diretos, texto em português e documentação técnica após a visão geral visual. As imagens são ativos de documentação solicitados pelo usuário. Não alterar telas, regras de negócio ou funcionalidades para compor as capturas.

Descrição aprovada: ERP para oficinas: clientes, recibos, estoque e gestão financeira com Go, Nuxt 4 e PostgreSQL.

## Evidências e validação

- Documentar somente recursos encontrados no código, nas rotas, nos testes e nos scripts do projeto.
- Capturar a interface real do frontend com serviços existentes e dados de demonstração, sem consultar a base operacional.
- Distinguir o PostgreSQL da aplicação e eventual SQLite usado exclusivamente no ambiente descartável de capturas.
- Conferir o README renderizado, os caminhos das imagens e o link para o Gist.
- Preservar a cópia de trabalho original do usuário.

## Aceitação

O perfil exibe o Gist entre os seis destaques e uma seção de notas técnicas. O repositório exibe a descrição aprovada e um README com imagens legíveis, dados de demonstração identificados, arquitetura e comandos reais de execução e teste.

## Conferência realizada

- Capturas atualizadas sobre a revisão pública `69fde53`, após integrar os commits remotos.
- Ambiente descartável: frontend Nuxt e API Go reais, com SQLite em memória e dados fictícios.
- Fluxos inspecionados: login, métricas, recibos e produtos (`/stock/products`).
- Capturas finais em 1440 × 900; overrides de viewport removidos após a inspeção.
- Tabelas mantêm a rolagem horizontal original do produto, sem alterações de layout para as capturas.
- Build do frontend e `TestMarkPaidDecreasesStockOnce` aprovados na revisão atual.
