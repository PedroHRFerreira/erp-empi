# Rods SDK no ERP Empi

O `@pedrohrferreira/rods-sdk` é gerenciado na raiz e governa todo o
repositório. `.ai/` é a fonte versionada da verdade e somente o target Codex
está habilitado.

## Atualização e diagnóstico

```bash
corepack pnpm run rods:upgrade:dry-run
corepack pnpm run rods:upgrade
corepack pnpm run rods:sync
corepack pnpm run rods:doctor
```

O pacote usa o alias `rods-sdk -> @pedrohrferreira/rods-sdk` porque o comando
`rods upgrade` da versão `0.1.5` ainda procura o nome legado ao identificar a
dependência do projeto.

## Context Engine

O banco fica isolado em `.rods/context-engine` e não é versionado.

```bash
CONTEXT_ENGINE_HOME=.rods/context-engine corepack pnpm exec context project add erp-empi .
corepack pnpm run context:ingest
corepack pnpm run context:ingest:review
corepack pnpm run context:stats
CONTEXT_ENGINE_HOME=.rods/context-engine corepack pnpm exec context search "termo" --scope general --limit 8
CONTEXT_ENGINE_HOME=.rods/context-engine corepack pnpm exec context read <chunkId>
```

## Cache Q&A

Escolha sempre uma política de validade explícita:

```bash
CONTEXT_ENGINE_HOME=.rods/context-engine corepack pnpm exec rods qa store --question "pergunta" --answer - --policy conceptual
CONTEXT_ENGINE_HOME=.rods/context-engine corepack pnpm exec rods qa search "pergunta"
corepack pnpm run rods:qa:stats
CONTEXT_ENGINE_HOME=.rods/context-engine corepack pnpm exec rods qa prune --stale --dry-run
```

Use `files` com `--files <caminhos>` para respostas ligadas a arquivos e
`repository` para respostas que dependem do estado completo do Git.

## Escalonamento Codex

O escalonamento está em modo `advisory`: ele classifica o trabalho sem trocar o
modelo automaticamente. Isso evita aliases instáveis e mantém a escolha do
modelo sob controle do operador.

```bash
corepack pnpm exec rods escalation classify "descrição da tarefa" --files <arquivos> --root . --json
```

Claude, `claude-mem`, `caveman` e workflows Codex/Claude permanecem
desabilitados. O RTK é o adapter padrão para compactar saídas de shell; confira
o ganho acumulado com `rtk gain`.
