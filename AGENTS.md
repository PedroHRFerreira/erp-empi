# AGENTS.md

## Rods SDK Defaults

Use rods-sdk defaults before reading large files, running noisy commands, or scanning the repository manually.

Detected stack: Go + Node/TypeScript (Nuxt)

1. Run `context_engine.search` with task-specific terms.
2. Read only relevant chunks with `context_engine.read`.
3. Use RTK for shell commands when it is available, especially for `git`, tests, logs, diffs, and broad searches.
4. Operate through the local harness/CLI, MCP tools, skills, and adapters. Do not call AI provider APIs directly from this framework.
5. Fall back to local file reads only when the index is missing or stale.
6. If fallback local reads solved the task, run `context_engine.ingest` on the relevant file or directory before finishing.
7. If a card or link implies external dependencies, ask whether to proceed, only plan, or take another action before running Context Engine search.

## Reading Map

| Case                             | Skill                                      |
| -------------------------------- | ------------------------------------------ |
| repository context / file lookup | `.ai/skills/context-search-first/SKILL.md` |
| architecture / boundaries        | `.ai/skills/architecture/SKILL.md`         |

| quality / readiness | `.ai/skills/quality/SKILL.md` |
| review / PR / commit | `.ai/skills/review/SKILL.md` |

## Running The Framework

Update and synchronize the consumer project:

```bash
corepack pnpm run rods:upgrade:dry-run
corepack pnpm run rods:upgrade
corepack pnpm run rods:sync
```

By default, skills stay in `.ai/skills`. If Codex needs a physical projection in another directory, pass it explicitly:

```bash
pnpm exec rods adapter sync --target codex --codex-skills-dir .codex/skills
```

Register and index the project in Context Engine:

```bash
CONTEXT_ENGINE_HOME=.rods/context-engine corepack pnpm exec context project add erp-empi .
corepack pnpm run context:ingest
corepack pnpm run context:stats
```

Search indexed context:

```bash
CONTEXT_ENGINE_HOME=.rods/context-engine corepack pnpm exec context search "search term"
CONTEXT_ENGINE_HOME=.rods/context-engine corepack pnpm exec context read <chunkId>
```

## Governance

Project governance lives in `.ai/`.

- `.ai/constitution.md` contains stable rules.
- `.ai/skills/*/SKILL.md` contains skills used as the project source of truth.
- `.ai/adapters/` contains optional adapter notes for external tools.
