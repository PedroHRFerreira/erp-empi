# Deploy

## Recomendacao atual

Use tudo no Render, com API no plano gratuito, frontend estatico gratuito, cron de keepalive e Postgres pago pequeno.

Motivo: o Render e o caminho mais simples para operar. O banco gratuito do Render nao atende producao de um ano porque expira apos 30 dias. O plano escolhido evita copiar connection string manualmente, mantem deploy automatico pelo GitHub e deixa tudo em um unico painel.

Plano inicial:

- `erp-empi-web`: Static Site, gratuito.
- `erp-empi-api`: Web Service Free.
- `erp-empi-keepalive`: Cron Job Starter, executa `/health` a cada 10 minutos.
- `erp-empi-db`: Render Postgres `basic-256mb`, 5 GB de storage.

Custo estimado inicial: cerca de USD 8.50/mes, considerando USD 1 do cron, USD 6 do Postgres `basic-256mb` e USD 1.50 de 5 GB de storage. Conferir o valor final no checkout do Render antes de criar.

Links uteis:

- Render Free: https://render.com/docs/free
- Render Blueprint: https://render.com/docs/blueprint-spec
- Render Pricing: https://render.com/pricing
- Render Cron Jobs: https://render.com/docs/cronjobs
- Render Postgres Flexible Plans: https://render.com/docs/postgresql-refresh

## Arquitetura de deploy

- `erp-empi-web`: Static Site no Render, gerado com Nuxt.
- `erp-empi-api`: Web Service Docker no Render.
- `erp-empi-keepalive`: Cron Job no Render, chama `https://<api>/health`.
- `erp-empi-db`: Postgres gerenciado no Render.

O frontend estatico chama a API diretamente usando `NUXT_PUBLIC_API_HOST`. A API usa `FRONTEND_URL` para liberar CORS somente para o dominio do frontend. A API recebe `DB_WRITE_DSN` automaticamente do banco criado pelo Blueprint.

## Variaveis obrigatorias no Render

O `render.yaml` ja cria ou solicita as principais variaveis:

- `ADMIN_CPF`: CPF do usuario admin inicial, somente numeros.
- `ADMIN_PASSWORD`: senha inicial do admin.
- `ADMIN_EMAIL`: email do usuario admin inicial.
- `ADMIN_PHONE`: telefone do usuario admin inicial, somente numeros.
- `JWT_ACCESS_SECRET`: gerado automaticamente pelo Render.

## Passos no Render

1. Entre no Render e conecte o GitHub.
2. Crie um novo Blueprint apontando para este repositorio.
3. Confirme o arquivo `render.yaml`.
4. Confira no checkout se o Postgres e o cron estao com o custo esperado.
5. Preencha `ADMIN_CPF`, `ADMIN_PASSWORD`, `ADMIN_EMAIL` e `ADMIN_PHONE` quando o Render solicitar.
6. Crie o Blueprint.
7. Aguarde `erp-empi-api`, `erp-empi-web`, `erp-empi-keepalive` e `erp-empi-db` ficarem ativos.
8. Abra o Static Site e faca login com o admin inicial.

## Deploy automatico

O Blueprint usa `autoDeployTrigger: commit`. Cada push na branch conectada no Render dispara novo deploy automaticamente para API e frontend.

Fluxo normal:

1. Fazer alteracao no codigo.
2. Commitar e enviar para o GitHub.
3. Acompanhar o deploy no painel do Render.
4. Validar o site quando os servicos ficarem verdes.

## Acompanhamento mensal

- Conferir se os ultimos deploys estao verdes.
- Verificar se o cron `erp-empi-keepalive` executou sem falhas recentes.
- Conferir uso e cobranca no billing.
- Conferir tamanho usado pelo banco.
- Conferir logs recentes da API.

## Observacoes

- O banco e migrado pelo `GORM AutoMigrate` quando a API sobe.
- A API continua no plano gratuito. O cron reduz o efeito de sleep por inatividade, mas nao transforma o servico free em servico com garantia paga.
- Se o uso crescer ou houver instabilidade, o primeiro upgrade recomendado e trocar apenas `erp-empi-api` de `free` para `starter`.
- O Postgres foi configurado com 5 GB porque o sistema salva dados estruturados e deve consumir pouco no primeiro ano. Aumente o storage no Render se o uso crescer.
