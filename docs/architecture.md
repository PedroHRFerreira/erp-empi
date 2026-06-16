# Arquitetura

O ERP EMPI e um monolito com backend Go e frontend Nuxt no mesmo repositorio.

## Backend

Fluxo principal:

```text
cmd/api/main.go
  -> components.Run()
    -> config.Load()
    -> database.NewPostgresClient()
    -> dig.NewContainer()
    -> Users.SeedAdmin()
    -> http.NewServer()
```

Camadas:

- `internal/api/http`: HTTP, middleware e handlers.
- `internal/domain`: entidades, repositories e services.
- `internal/infra`: infraestrutura como banco.
- `internal/app/dig`: injecao manual de dependencias.
- `internal/shared`: seguranca, validacao e erros comuns.

## Frontend

O browser conversa apenas com o BFF em `frontend/server/api`.

```text
Browser -> useApiFetch/useAuthToken -> Nuxt BFF -> Go API -> PostgreSQL
```

Stores Pinia concentram estado, validacao de formulario e chamadas ao composable `useApiFetch`.
O BFF usa rotas explicitas por arquivo, sem catch-all e sem helper global de proxy.

## Dados

Valores monetarios sao salvos em centavos. Estoque usa soft delete por `active=false` para preservar historico de recibos.
