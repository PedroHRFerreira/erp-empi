# ERP EMPI Autocenter

Monolito para operacao administrativa da EMPI Autocenter.

Stack:

- Go + Echo + GORM + PostgreSQL
- Nuxt 4 + TypeScript + Pinia + SCSS
- Nitro BFF em `frontend/server/api`
- Docker Compose para execucao local/deploy simples

## Setup local

```sh
make setup
docker compose up --build
```

Servicos:

- Frontend: `http://localhost:3000`
- API: `http://localhost:8080/health`
- PostgreSQL: `localhost:5432`

O admin inicial e criado no bootstrap usando as variaveis `ADMIN_*`.

## Desenvolvimento sem Docker

Backend:

```sh
go run ./cmd/api
```

Frontend:

```sh
pnpm --dir frontend install
pnpm --dir frontend dev
```

## Fluxos do MVP

- Login com CPF e senha do admin.
- Cadastro automatico de cliente ao criar recibo.
- Recibos com produtos usados, status pendente/pago, WhatsApp, texto para Instagram e impressao/PDF.
- Estoque paginado, com adicionar, editar, inativar, CSV/Excel e impressao/PDF.
- Ao marcar recibo como pago, o estoque e baixado uma unica vez.
- Metricas calculadas a partir de users, receipts e stock_items.
- Perfil do admin com margem de revenda e juros de cartão.

## Validacao

```sh
make test
go test ./...
pnpm --dir frontend lint
pnpm --dir frontend test
```

## Documentacao

- `docs/architecture.md`
- `docs/security.md`
- `docs/deploy.md`
- READMEs nas pastas principais
