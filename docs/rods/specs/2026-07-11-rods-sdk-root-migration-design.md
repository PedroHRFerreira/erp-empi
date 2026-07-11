# Migração do Rods SDK para a raiz

## Objetivo

Centralizar o `@pedrohrferreira/rods-sdk` na raiz do repositório para que sua
governança cubra backend, frontend e demais arquivos do projeto.

## Escopo aprovado

- Atualizar para a versão `0.1.5`.
- Remover a dependência do pacote do `frontend`.
- Manter `.ai/` como fonte versionada da governança.
- Regenerar e sincronizar somente o target Codex.
- Habilitar RTK, Context Engine, hooks, cache Q&A e escalonamento para Codex.
- Manter Claude e o fluxo multiagente Codex/Claude desabilitados.
- Registrar e indexar o repositório inteiro no Context Engine.
- Validar a instalação com doctor, busca indexada e testes relevantes.

## Estrutura

A raiz passa a ser o projeto consumidor do Rods SDK e terá seu próprio
`package.json` e lockfile. O `frontend` continua sendo um projeto pnpm
independente para sua aplicação Nuxt, sem carregar a ferramenta de governança.

## Critérios de aceite

1. Os binários `rods`, `context` e `context-mcp` podem ser executados da raiz.
2. A configuração reconhece apenas Codex como target ativo.
3. O adapter doctor não aponta erro bloqueante na integração Codex.
4. O Context Engine indexa e encontra conteúdo de diferentes áreas do projeto.
5. O frontend permanece instalável e passa em typecheck/testes aplicáveis.
