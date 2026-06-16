# server

BFF Nitro do Nuxt.

O browser chama apenas `/api/*` neste frontend. O BFF:

- possui rotas explicitas por recurso, sempre com arquivos `index.get.ts`, `index.post.ts`, `index.put.ts` ou `index.delete.ts`;
- usa `$fetch` direto para chamar a API Go;
- repassa o header `Authorization` recebido do composable `useApiFetch`;
- nao guarda regra de token nem proxy generico em `server/utils`.
