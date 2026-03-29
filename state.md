# Meridian — Implementation state

**Last reviewed:** 2026-03-28 (repo snapshot).  
For intent and constraints, see [requirement.md](requirement.md). For future work, see [roadmap.md](roadmap.md).

---

## Summary

| Area | Status | Notes |
|------|--------|--------|
| Go CLI binary | **Working** | `cmd/meridian` + `internal/cli` |
| Config (`~/.meridian/config.yaml`) | **Working** | Optional file; defaults + env overrides |
| Qdrant client | **Working** | gRPC via `qdrant/go-client`; collection `meridian_projects` |
| Ingest (CNCF + GitHub + embed + upsert) | **Working** | `internal/ingestion`; workers=6; requires dims match |
| `find` | **Working** | Vector search + LLM rerank + fallback table |
| `explore` | **Working** | Scroll + substring category match + sort by stars |
| `recommend` | **Working** | Loads `~/.meridian/profile.yaml`; needs file populated |
| `init` | **Partial** | Bubbletea + Glamour welcome; **no full profile save yet** |
| Agent / ReAct | **Partial** | Rerank path only; no multi-step tool loop |
| Extension v0.2 | **Scaffold** | TS/React stub; **no CRXJS, no API :7474** |
| Tests | **Minimal** | `internal/retry` tests; broader coverage TBD |
| CI / lint | **Ad hoc** | `make lint` expects local `golangci-lint` |

---

## Package map (what lives where)

| Package | Responsibility |
|---------|------------------|
| `internal/config` | Viper load, defaults, env bind |
| `internal/ingestion` | `cncf.go` fetch/parse, `github.go` enrich, `pipeline.go` orchestrate, `embedtext.go` |
| `internal/vectorstore` | `types.go` `ProjectDoc`, `qdrant.go` store, `payload.go` map ↔ Qdrant `Value` |
| `internal/embeddings` | `Embedder` + Ollama/OpenRouter + `factory.go` |
| `internal/llm` | `Provider` + Ollama/OpenRouter + `factory.go` |
| `internal/agent` | `find.go`, `explore.go`, `prompt.go`; `agent.go` minimal |
| `internal/cli` | Cobra commands, `ingest_run.go` |
| `internal/profile` | `UserProfile`, `DefaultPath`, `Load`, `SkillsString` |
| `internal/qdrantutil` | Parse `qdrant.url` to host/port |
| `internal/retry` | Small retry helper |

---

## Behavioral details (important for debugging)

### Ingestion

- **Landscape URL:** `https://raw.githubusercontent.com/cncf/landscape/master/landscape.yml`
- **GitHub:** `go-github/v67`; token via `MERIDIAN_GITHUB_TOKEN` recommended.
- **Good first issues:** Search API query with label `"good first issue"`.
- **Contributors:** Paginated list (capped pages) for approximate count.
- **Point IDs:** FNV-1a 64-bit of normalized `repo_url`.
- **Dry run:** Parses + enriches + logs; **no** Qdrant/embed writes.

### Find

- Retrieves a larger batch from Qdrant then filters to **20** (implementation detail in `vectorstore.SearchSimilar`).
- LLM must return JSON array; code strips ``` fences; on failure, **similarity-only** table (up to 20 rows in fallback).

### Explore

- Uses **Scroll** + **in-memory** substring match on `category + sub_category` (no Qdrant text index required).
- Sorted by **stars** descending.

### JSON / sonic

- CLI JSON encoding uses **`encoding/json`** for compatibility with current Go releases.
- `bytedance/sonic` may appear **indirectly** in `go.mod`; not imported from first-party packages.

---

## Known limitations / sharp edges

1. **`init` does not write `profile.yaml`** — `recommend` fails until profile exists manually.
2. **Rate limits** — unauthenticated GitHub ingest is fragile; token strongly recommended.
3. **Embedding dimension** — must match `qdrant.embedding_dims`; mismatches fail ingest or search quality.
4. **`meridian` on PATH** — users must run `./bin/meridian` or install binary (see README troubleshooting).
5. **Extension** — not usable as product yet; manifest points at TS sources (needs bundling for real load).

---

## Verification commands

```bash
make build
make test
make qdrant
./bin/meridian ingest --limit 20
./bin/meridian find --skills "kubernetes, go" --level intermediate
./bin/meridian explore --category observability --limit 10
```

---

## Changelog pointer

This file should be updated when:

- New commands or flags ship.
- Extension or local API lands.
- Data model (`ProjectDoc`) or collection name changes.

Git history remains the source of truth for exact diffs; **state.md** is the **human/LLM orientation** layer.
