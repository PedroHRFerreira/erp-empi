# Integração Cloudflare R2 para backups

## Escopo aprovado

- Instalar globalmente as skills oficiais da Cloudflare para o Codex.
- Registrar os MCPs oficiais `cloudflare`, `cloudflare-docs`, `cloudflare-bindings`, `cloudflare-builds` e `cloudflare-observability`.
- Autenticar via OAuth na conta Cloudflare do proprietário do sistema.
- Criar o bucket privado Standard `erp-empi-backups`.
- Manter backups diários por 30 dias e cópias mensais por 365 dias.
- Usar o Cron Job separado definido em `render.yaml`, com execução diária às 03:00 UTC.

## Segurança

- O token de backup deve ter acesso apenas ao bucket necessário.
- Credenciais não serão gravadas no repositório.
- A produção não será alterada durante a instalação da integração.
- A primeira execução deve validar o tamanho remoto e o SHA-256; a restauração será testada separadamente.

## Custos aceitos

- Skills e MCPs não possuem cobrança própria.
- R2 Standard deverá permanecer na faixa gratuita enquanto estiver abaixo dos limites oficiais.
- O novo Cron Job do Render possui cobrança mínima adicional de US$ 1 por mês.
- Qualquer uso futuro acima da faixa gratuita do R2 será cobrado pela Cloudflare conforme a tabela vigente.

## Aceite

A integração estará pronta quando os MCPs estiverem registrados, a autenticação estiver concluída, o bucket e sua retenção estiverem configurados e as credenciais necessárias puderem ser vinculadas ao serviço de backup sem exposição no código.
