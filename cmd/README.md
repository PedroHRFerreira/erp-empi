# cmd

Entrypoints da aplicacao.

- `api/main.go` deve permanecer pequeno.
- Bootstrap, dependencias externas e roteamento ficam em `api/components`.
- Regra de negocio pertence a `internal/domain`.
