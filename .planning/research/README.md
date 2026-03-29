# Meridian — Research & reference index

This folder holds **curated, stable pointers** for planning and implementation. Prefer these docs over stale chat context when onboarding an engineer or LLM.

| File | Purpose |
|------|---------|
| [SOURCES.md](SOURCES.md) | Canonical URLs, APIs, version pins |
| [integrations.md](integrations.md) | How Meridian talks to Qdrant, GitHub, Ollama, OpenRouter |
| [data-contracts.md](data-contracts.md) | Landscape parsing, `ProjectDoc`, prompts, Qdrant payload |
| [open-questions.md](open-questions.md) | Unresolved product/tech choices |

**Upstream narrative docs (repo root):** [project.md](../../project.md), [requirement.md](../../requirement.md), [roadmap.md](../../roadmap.md), [state.md](../../state.md).

---

## How to use this folder

1. **Before a feature:** read [state.md](../../state.md) + [requirement.md](../../requirement.md) for scope.  
2. **Before touching ingest/search:** read [data-contracts.md](data-contracts.md) + [integrations.md](integrations.md).  
3. **Before changing deps:** check [SOURCES.md](SOURCES.md) and run `go test ./...`.

---

## Maintenance

- Update **SOURCES.md** when bumping `go.mod` majors (Qdrant client, go-github, langchaingo).
- Update **data-contracts.md** when `ProjectDoc` or `meridian_projects` schema changes.
- Keep **open-questions.md** short; resolve items by linking PRs or ADRs when they exist.
