# Ensaio da migração legada de produção

Data do ensaio: 2026-08-23

## Backup usado

- Origem: PostgreSQL de produção, acesso somente leitura por `pg_dump`.
- Formato: custom (`pg_restore`).
- Arquivo temporário: `/tmp/erp-empi-production-2026-08-23.dump`.
- SHA-256: `7ec674bb771f68283a02899226261b9b17a5a7db69826775b74a9cd3cf134902`.
- O arquivo temporário deve ser copiado para armazenamento durável e criptografado antes do deploy.

## Ambiente do ensaio

- Banco isolado: `empi_migration_dryrun_20260823` no PostgreSQL local.
- O backup foi restaurado com `pg_restore --no-owner --no-acl --exit-on-error`.
- A produção não recebeu escritas.
- A migração foi executada duas vezes por `go run ./cmd/migrate`.

## Resultado antes/depois

| Invariante | Resultado |
| --- | ---: |
| Gastos totais preservados | 83 |
| Gastos ativos | 80 |
| Total dos gastos ativos | R$ 10.126,28 |
| Parcelas legadas de gastos | 80 |
| Total das parcelas legadas de gastos | R$ 10.126,28 |
| Produtos preservados | 9 |
| Quantidade disponível preservada | 25 |
| Quantidade usada preservada | 3 |
| Compras históricas criadas | 9 |
| Parcelas legadas de estoque | 9 |
| Total histórico de estoque | R$ 1.175,78 |
| Recibos preservados | 63 |
| Usuários preservados | 26 |
| Sessões de caixa retroativas | 0 |
| Movimentos de caixa retroativos | 0 |
| Parcelas com origem inválida | 0 |
| Marcadores da migração após duas execuções | 1 |

## Roteiro de publicação futura

1. Agendar janela de manutenção e bloquear escrita na aplicação antiga.
2. Gerar um novo backup customizado e armazená-lo de forma durável e criptografada.
3. Validar o arquivo com `./scripts/predeploy-migrate.sh verify <backup.dump>` e guardar o SHA-256 exibido fora do servidor.
4. Executar `BACKUP_SHA256='<sha256>' DB_WRITE_DSN='<dsn>' CONFIRM_PRODUCTION_MIGRATION=migrate-legacy-production ./scripts/predeploy-migrate.sh migrate <backup.dump>` antes de liberar a API nova.
5. Interromper o deploy se o comando retornar erro; a transação reverte esquema e dados da execução.
6. Conferir as mesmas invariantes da tabela acima usando os valores do backup feito no momento do deploy.
7. Fazer smoke test de login, recibos, estoque, gastos, históricos, Contas a pagar e caixa.
8. Liberar a aplicação somente após aprovação das contagens e dos fluxos.

## Rollback

1. Manter a aplicação em manutenção.
2. Interromper a API nova.
3. Preservar um dump do estado que falhou para diagnóstico.
4. Criar um banco limpo; não restaurar por cima do banco parcialmente migrado.
5. Restaurar o backup pré-deploy com `pg_restore --clean --if-exists` no banco de recuperação.
6. Apontar a aplicação anterior para o banco restaurado e validar contagens antes de reabrir.

## Segurança

- Nunca registrar a senha do banco em documentos, logs ou comandos versionados.
- Rotacionar a credencial de produção compartilhada durante esta preparação.
- Restringir o backup a quem administra produção e testar a restauração antes do deploy.
