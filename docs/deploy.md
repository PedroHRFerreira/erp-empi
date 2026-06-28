# Deploy

## Recomendacao atual

Use Render para a API e o frontend, mas use um Postgres externo no plano gratuito.

Motivo: o Render tem Web Services e Static Sites gratuitos, mas o Postgres gratuito do Render expira apos 30 dias. Para um banco que nao some, prefira Neon ou Supabase no gratuito. Se a prioridade for muito espaco gratuito e voce aceitar administrar servidor, Oracle Cloud Always Free costuma ser a melhor opcao porque permite VM e volume persistente maiores, mas exige SSH, firewall, backup e manutencao.

Links uteis:

- Render Free: https://render.com/docs/free
- Render Blueprint: https://render.com/docs/blueprint-spec
- Neon pricing: https://neon.com/pricing
- Supabase pricing: https://supabase.com/pricing
- Oracle Always Free: https://docs.oracle.com/en-us/iaas/Content/FreeTier/freetier_topic-Always_Free_Resources.htm

## Arquitetura de deploy

- `erp-empi-web`: Static Site no Render, gerado com Nuxt.
- `erp-empi-api`: Web Service Docker no Render.
- Postgres: banco externo, informado em `DB_WRITE_DSN`.

O frontend estatico chama a API diretamente usando `NUXT_PUBLIC_API_HOST`. A API usa `FRONTEND_URL` para liberar CORS somente para o dominio do frontend.

## Variaveis obrigatorias no Render

O `render.yaml` ja cria ou solicita as principais variaveis:

- `DB_WRITE_DSN`: string de conexao Postgres com SSL.
- `ADMIN_CPF`: CPF do usuario admin inicial, somente numeros.
- `ADMIN_PASSWORD`: senha inicial do admin.
- `JWT_ACCESS_SECRET`: gerado automaticamente pelo Render.

Exemplo de `DB_WRITE_DSN` em formato URL:

```text
postgresql://USER:PASSWORD@HOST:5432/DATABASE?sslmode=require
```

Exemplo em formato chave/valor:

```text
host=HOST user=USER password=PASSWORD dbname=DATABASE port=5432 sslmode=require TimeZone=America/Sao_Paulo
```

## Passos no Render

1. Crie o banco no Neon ou Supabase e copie a connection string com SSL.
2. No Render, crie um novo Blueprint apontando para este repositorio.
3. Confirme o arquivo `render.yaml`.
4. Preencha `DB_WRITE_DSN`, `ADMIN_CPF`, `ADMIN_PASSWORD`, `ADMIN_EMAIL` e `ADMIN_PHONE` quando o Render solicitar.
5. Aguarde o deploy do `erp-empi-api` e do `erp-empi-web`.
6. Abra o Static Site e faca login com o admin inicial.

## Observacoes

- O banco e migrado pelo `GORM AutoMigrate` quando a API sobe.
- No plano gratuito do Render, a API pode dormir por inatividade e a primeira chamada pode demorar.
- Para producao com uso diario, configure backup no provedor do banco.
