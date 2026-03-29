# Meridian — Project handbook

This document gives engineers and LLM agents a **single canonical overview**: identity, architecture, repository layout, runtime behavior, and where to change code.

---

## 1. Identity

| Field | Value |
|--------|--------|
| **Name** | Meridian |
| **Tagline** | Find your place in open source. |
| **Target repo** | `github.com/iemafzalhassan/meridian` |
| **License** | Apache-2.0 |
| **Language (CLI)** | Go (`go.mod` module: `github.com/iemafzalhassan/meridian`) |
| **Language (extension)** | TypeScript (v0.2, scaffold only today) |

---

## 2. Problem the product solves

Developers want to contribute to open source but get lost in keyword search and huge landscapes (e.g. CNCF). **Meridian** combines:

- **Semantic search** over an enriched project index (dense vectors in **Qdrant**).
- **LLM reasoning** to rerank and explain fit (skills → projects, “why you belong”).

Sources of truth for the narrative: [requirement.md](requirement.md), [roadmap.md](roadmap.md), [state.md](state.md).

---

## 3. Product surfaces (non‑negotiable scope)

| Surface | Version | Status |
|---------|---------|--------|
| **CLI** | v0.1 | Primary; **single static binary**, terminal-first |
| **Browser extension** | v0.2 | Scaffold; **GitHub.com tabs only** (no global tab analysis) |

Both must eventually share:

- Same **Qdrant** collection and schema.
- Same **LLM / embedding provider abstraction** (Ollama + OpenRouter).
- Same **ingestion / enrichment pipeline** (CNCF landscape → GitHub → embed → upsert).

---

## 4. High-level architecture

```text
                    ┌─────────────────────┐
                    │  CNCF landscape.yml  │
                    └──────────┬──────────┘
                               │ HTTP (resty) + YAML parse
                               ▼
                    ┌─────────────────────┐
                    │ GitHub API enrich    │ stars, languages, topics, GFI count, …
                    └──────────┬──────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │ Build EmbedText      │ per ProjectDoc contract
                    └──────────┬──────────┘
                               │ embeddings.Embedder (Ollama / OpenRouter)
                               ▼
                    ┌─────────────────────┐
                    │ Qdrant (meridian_     │ dense vectors + payload
                    │   projects)         │
                    └──────────┬──────────┘
                               │
         ┌─────────────────────┴─────────────────────┐
         │                                           │
         ▼                                           ▼
  meridian find / recommend                  meridian explore
  (embed query → top-20 → LLM rerank)       (scroll + category substring)
```

---

## 5. Repository layout (authoritative)

```text
meridian/
├── cmd/meridian/main.go          # Entry: config load, zap, cobra Execute
├── internal/
│   ├── agent/                    # find/explore orchestration, rerank prompt helpers
│   ├── cli/                      # Cobra commands + wiring
│   ├── config/                   # Viper: ~/.meridian/config.yaml + env
│   ├── embeddings/               # Embedder interface + Ollama / OpenRouter
│   ├── ingestion/                # CNCF fetch/parse, GitHub enrich, pipeline
│   ├── llm/                      # Provider interface + Ollama / OpenRouter
│   ├── profile/                  # User profile types + paths
│   ├── qdrantutil/               # Qdrant URL → host:port
│   ├── retry/                    # Shared retry helper (+ tests)
│   └── vectorstore/              # ProjectDoc, Qdrant client wrapper, payload ↔ struct
├── extension/                    # v0.2 TS scaffold (Vite, manifest, content script)
├── docker-compose.yml            # Local Qdrant
├── Makefile
├── .meridian.yaml.example        # Copy → ~/.meridian/config.yaml
├── README.md
├── project.md                    # This file
├── requirement.md
├── roadmap.md
├── state.md
└── .planning/research/           # Research index + integration notes
```

---

## 6. Configuration & local paths

| Purpose | Path |
|---------|------|
| **Config** | `~/.meridian/config.yaml` (optional; defaults apply if missing) |
| **User profile** | `~/.meridian/profile.yaml` (`recommend` expects this; `init` is still minimal) |
| **Example config** | Repo root: [`.meridian.yaml.example`](.meridian.yaml.example) |

**Environment overrides** (see `internal/config/config.go`):

- `MERIDIAN_OPENROUTER_KEY` → `llm.openrouter_key`
- `MERIDIAN_OLLAMA_URL` → `llm.ollama_url`
- `MERIDIAN_GITHUB_TOKEN` → `github.token`

**Secrets:** never commit real keys; use env or untracked user config.

---

## 7. Data model (vector payload)

Defined in `internal/vectorstore/types.go` as **`ProjectDoc`**. Minimum fields engineers must preserve when changing ingestion or Qdrant:

- Identity: `id`, `name`, `repo_url`, `homepage_url`
- Taxonomy: `category`, `sub_category` (from CNCF landscape walk)
- Enrichment: `languages`, `topics`, `stars`, `good_first_issues`, `contributor_count`, `open_issues_count`, `last_commit_date`
- **Embed source:** `embed_text` — built by `ingestion.BuildEmbedText` using the product contract (name, description, category, languages, topics)

**Qdrant:**

- Collection name: **`meridian_projects`** (config default; do not rename without migration plan).
- Vector distance: **Cosine** (see `internal/vectorstore/qdrant.go`).
- Point ID: deterministic hash of normalized `repo_url` (FNV-1a 64-bit → numeric id).

---

## 8. CLI commands (current behavior)

| Command | Role |
|---------|------|
| `meridian init` | Minimal Bubbletea flow; Glamour-rendered welcome (full profile wizard deferred). |
| `meridian find --skills "..." [--level ...]` | Embed skills → Qdrant top-20 → LLM rerank JSON → table; fallback if LLM/parse fails. |
| `meridian explore --category "..." [--limit ...]` | Scroll Qdrant, substring match on category path, sort by stars. |
| `meridian recommend` | Load profile YAML, merge skills/interests, same pipeline as `find`. |
| `meridian ingest [--limit N] [--dry-run]` | Fetch landscape → enrich → embed → upsert. |

**Operational prerequisite:** Qdrant up, embedding provider available, `qdrant.embedding_dims` aligned with embedding model (768 vs 1536).

---

## 9. Key dependencies (Go)

Pinned in [go.mod](go.mod). Notable:

- **CLI:** `spf13/cobra`, `spf13/viper`, `uber-go/zap`
- **TUI / output:** `charmbracelet/bubbletea`, `lipgloss`/`table`, `glamour`
- **LLM:** `tmc/langchaingo` (Ollama + OpenAI-compatible OpenRouter)
- **Vector DB:** `qdrant/go-client`
- **HTTP / APIs:** `go-resty/resty`, `google/go-github/v67`
- **YAML:** `gopkg.in/yaml.v3`

**Note:** Application JSON for LLM payloads uses `encoding/json` for compatibility with recent Go toolchains; `sonic` may appear transitively but is not required in app packages.

---

## 10. How to work safely in this codebase

1. **No global mutable service singletons** — pass `config`, `logger`, clients via structs / Cobra `Dependencies`.
2. **Wrap errors** with context: `fmt.Errorf("domain: %w", err)`.
3. **External calls:** use timeouts from config + `internal/retry` where appropriate.
4. **Provider swappability:** new backends implement `llm.Provider` and `embeddings.Embedder`; select via config only.
5. **Apache-2.0 headers** on new `.go` files.

---

## 11. Cross-links

| Doc | Use when |
|-----|----------|
| [requirement.md](requirement.md) | What “done” means, constraints, acceptance criteria |
| [roadmap.md](roadmap.md) | Phased delivery and milestones |
| [state.md](state.md) | What is implemented vs stubbed today |
| [.planning/research/README.md](.planning/research/README.md) | External sources, APIs, risks for planning |

---

## 12. Extension (v0.2) — current repo reality

Under `extension/`: Vite config, MV3 `manifest.json`, `background.ts`, GitHub-only `content.ts`, React `sidebar/App.tsx`. **Not yet** wired to a local Meridian API (port 7474) or CRXJS; treat as **scaffold** until roadmap items land.
