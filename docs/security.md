# Seguranca

Controles do MVP:

- Login apenas para `users.type = admin`.
- Senha com bcrypt.
- Access token guardado pelo composable `useAuthToken` e expiracao definida por `JWT_ACCESS_TTL_MINUTES`.
- `useApiFetch` centraliza envio de `Authorization` para o BFF.
- BFF apenas repassa `Authorization` nas chamadas para a API Go.
- Rotas privadas protegidas por JWT no Go e middleware global no Nuxt.
- CORS restrito ao `FRONTEND_URL`.
- Headers seguros no Nuxt e no Echo.
- Rate limit basico no backend.
- CPF, placa, status, valores e quantidades validados no servidor.

Regras operacionais:

- Nao commitar `.env`.
- Trocar `JWT_ACCESS_SECRET` e `ADMIN_PASSWORD` antes de deploy.
- Criar secrets por ambiente no provedor de deploy.
