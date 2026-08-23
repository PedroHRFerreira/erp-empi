# Backups automáticos no Cloudflare R2

O PostgreSQL pago do Render continua fornecendo PITR para recuperação recente. O Cron Job `erp-empi-database-backup` cria uma segunda camada externa diariamente às 03:00 UTC.

## Custo esperado

- Render Cron Job: mínimo de US$ 1/mês, além do tempo ativo.
- Cloudflare R2 Standard: dentro da faixa gratuita enquanto o total ficar abaixo de 10 GB e do limite gratuito de operações.
- Não usar R2 Infrequent Access, pois a faixa gratuita não se aplica a essa classe.

## Configuração do R2

1. Criar uma conta Cloudflare e habilitar R2.
2. Criar um bucket privado em armazenamento Standard, por exemplo `erp-empi-backups`.
3. Criar um token R2 exclusivo, limitado a leitura e escrita de objetos nesse bucket. Não conceder administração da conta.
4. Configurar regras de ciclo de vida no bucket:
   - prefixo `erp-empi/daily/`: excluir após 30 dias;
   - prefixo `erp-empi/monthly/`: excluir após 365 dias.
5. No serviço `erp-empi-database-backup` do Render, preencher os segredos:
   - `R2_ENDPOINT`: `https://<ACCOUNT_ID>.r2.cloudflarestorage.com`;
   - `R2_BUCKET`: nome do bucket;
   - `AWS_ACCESS_KEY_ID`: Access Key ID do token R2;
   - `AWS_SECRET_ACCESS_KEY`: Secret Access Key do token R2.

## Funcionamento

- O job usa a conexão direta do Render Postgres, não PgBouncer.
- `pg_dump` gera um arquivo customizado.
- `pg_restore --list` valida a estrutura antes do upload.
- Um SHA-256 é enviado ao lado de cada backup.
- O tamanho remoto é comparado ao arquivo local.
- No primeiro dia do mês, o mesmo backup também é salvo no prefixo mensal.
- Arquivos temporários são removidos ao final da execução.

## Primeira ativação

1. Sincronizar o Blueprint no Render.
2. Preencher todos os segredos antes de habilitar o job.
3. Usar **Trigger Run** no painel.
4. Confirmar no log a mensagem `backup uploaded and verified`.
5. Baixar o objeto e validar localmente com `pg_restore --list` e `sha256sum -c`.
6. Somente depois confiar na agenda diária.

## Teste periódico de restauração

Uma vez por mês, baixar o backup mensal mais recente, restaurá-lo em um banco vazio e executar as verificações descritas em `docs/migrations/2026-08-23-legacy-production-rehearsal.md`. Backup sem teste de restauração não deve ser considerado comprovadamente recuperável.
