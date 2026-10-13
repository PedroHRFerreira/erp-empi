# Marketing Studio

Administradores acessam **Marketing** no menu. Configure a identidade visual uma vez, escolha uma referência em **Modelos** ou use **Criar** para divulgar produtos do estoque, serviços ou ofertas combinadas. Consulte [o guia dos modelos da EMPI](marketing-models.md) para detalhes do catálogo importado.

## Uso

1. Escolha o tipo de conteúdo, objetivo e composição. Produtos preenchem título, descrição e preço pelo estoque; serviços e ofertas são preenchidos manualmente.
2. Ajuste textos, foto, cores e formato. Escolha uma das três sugestões de legenda ou escreva a sua.
3. Salve para reutilizar, baixe o PNG e copie a legenda. A publicação nas redes sociais é manual.

**Minhas criações** permite abrir, duplicar e excluir artes. Nome, preço e identidade visual ficam preservados na criação: alterações posteriores no estoque ou na marca só entram por ação explícita no editor.

## Processamento e persistência

- Prévia e exportação usam Canvas no navegador, com os mesmos desenhos e dimensões: Quadrado 1080×1080; Feed 1080×1350; Story 1080×1920. Ajustes visuais e download não chamam a API.
- Upload, leitura de imagens e salvamento usam APIs autenticadas, restritas a administradores. PNG/JPEG até 5 MiB e 20 milhões de pixels são validados antes da persistência.
- PostgreSQL armazena identidade visual compartilhada do negócio, criações e imagens em tabelas próprias. A migração segue o mecanismo de schema já existente no backend.
- O módulo não altera preços do estoque, recibos ou descontos. Não gera métricas de receita atribuída e não integra IA, Instagram ou WhatsApp nesta versão.
- O desempenho de exportação depende do navegador e aparelho. A execução local não exige nem garante uso exclusivo da GPU.

## Verificação

Backend: `rtk go test ./internal/domain/marketing/... ./internal/api/http/... ./internal/infra/database/... ./internal/app/dig/...`

Frontend: `rtk proxy corepack pnpm --dir frontend test`

Fluxos em Chromium: `rtk proxy corepack pnpm --dir frontend exec playwright test e2e/marketing.spec.ts`. Requer backend e frontend ativos e conta administradora; configure `E2E_ADMIN_CPF`, `E2E_ADMIN_PASSWORD` e `PLAYWRIGHT_BASE_URL` quando necessário.

Os testes de navegador verificam persistência, reabertura, duplicação, exclusão, PNG local sem chamadas à API, preservação do rascunho em falha de salvamento e rejeição de imagem inválida nos três tamanhos configurados.

## Personalização compacta

A etapa Personalizar mostra texto principal, foto e Mais opções. Nos modelos EMPI, tocar em outro texto na prévia seleciona esse texto para edição. Formato, produto, preço, outros elementos e ajustes visuais continuam acessíveis em Mais opções. A prévia acompanha a edição e permanece visível durante a rolagem em telas pequenas.
