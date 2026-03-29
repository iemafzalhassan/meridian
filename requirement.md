# Meridian — Requirements

Audience: **engineers, PMs, and LLM agents** planning or changing Meridian. This is the normative list of what the product must do and what it must never do.

---

## 1. Problem statement (fixed)

**Developers** want to contribute to **open source** but:

- General search is **keyword‑centric** and noisy.
- Ecosystem maps (e.g. **CNCF landscape**) are **large and static**.
- Mapping **personal skills** to **concrete repos** is hard.

**Meridian** addresses this with **semantic search + LLM reasoning** over a **curated, enriched index** of CNCF (and extensible OSS) projects.

---

## 2. Goals

| ID | Goal |
|----|------|
| G1 | Help a developer **discover repos** where their skills transfer. |
| G2 | Surface **good-first-issue** signals where available. |
| G3 | Explain **why** a project fits (LLM-generated, grounded on candidates). |
| G4 | Run **locally** with OSS-friendly infra (Qdrant self-hosted, Ollama optional). |
| G5 | Support **cloud LLMs** via OpenRouter without hard-coding a single vendor. |

---

## 3. Non-goals (explicit)

| ID | Non-goal |
|----|----------|
| NG1 | **General web** or **all-browser-tabs** analysis in the extension. |
| NG2 | **Python** in the CLI/runtime path. |
| NG3 | Replacing GitHub’s native issue tracker UI as the source of truth for issues. |
| NG4 | Guaranteed completeness of every CNCF row (some items lack `repo_url`). |

---

## 4. User stories (v0.1 CLI)

| ID | As a… | I want… | So that… |
|----|--------|---------|----------|
| US1 | Developer | to **index** CNCF + GitHub metadata into a vector DB | I can search semantically offline/local. |
| US2 | Developer | **`find --skills`** | I get a **ranked, explained** short list. |
| US3 | Developer | **`explore --category`** | I browse a **slice** of the landscape by theme. |
| US4 | Developer | **`recommend`** using a **saved profile** | I don’t re-type skills every time. |
| US5 | Operator | **`ingest --dry-run`** | I can validate pipeline without writes. |

---

## 5. Functional requirements

### 5.1 Ingestion & index

| Req ID | Requirement |
|--------|-------------|
| F1 | Ingest **CNCF** `landscape.yml` from the official raw URL (see `.planning/research/SOURCES.md`). |
| F2 | Parse **category**, **subcategory**, **item** fields: `name`, `description`, `homepage_url`, `repo_url`. |
| F3 | **Skip** items without a **github.com** `repo_url`. |
| F4 | Enrich each repo via **GitHub API**: topics, languages, stars, approximate contributor count, open issues count, last push time, **good first issue** count (search-based). |
| F5 | Build **`embed_text`** exactly per product contract (see [project.md](project.md) §7). |
| F6 | Upsert into Qdrant collection **`meridian_projects`** with **dense** vectors; dimension **must match** configured `qdrant.embedding_dims`. |

### 5.2 Search & agent

| Req ID | Requirement |
|--------|-------------|
| F7 | **`find`**: embed user skill string → **vector search ~top 20** → LLM **rerank top 5** with JSON schema `{name, reason, good_first_issues, repo_url, score}`. |
| F8 | **`explore`**: filter by **category substring** (case-insensitive) on stored category path; present a bounded list sorted by **stars** (current implementation). |
| F9 | **`recommend`**: load **`~/.meridian/profile.yaml`**; use skills (+ interests as supplemental text) and level with same core pipeline as `find`. |

### 5.3 Configuration & providers

| Req ID | Requirement |
|--------|-------------|
| F10 | **LLM** and **embeddings** providers selectable via config: **`ollama`** | **`openrouter`**. |
| F11 | **OpenRouter** uses OpenAI-compatible **`https://openrouter.ai/api/v1`** with user-supplied key. |
| F12 | **Ollama** URL configurable; used for chat and/or embeddings per config. |
| F13 | Config file **`~/.meridian/config.yaml`** optional; sensible defaults when missing. |

---

## 6. Non-functional requirements

| Req ID | Requirement |
|--------|-------------|
| NF1 | **Single static binary** for CLI distribution (Go build). |
| NF2 | **Structured logging** (zap); no logging of secrets or raw tokens. |
| NF3 | **Timeouts** on HTTP/gRPC operations (configured per subsystem). |
| NF4 | **Retries** for flaky external calls (implemented for key paths; extend consistently). |
| NF5 | Errors **wrapped** with domain context (`fmt.Errorf("…: %w", err)`). |
| NF6 | **Dependency injection**; avoid new global mutable singletons. |
| NF7 | **Apache-2.0** license headers on Go sources. |

---

## 7. Hard constraints (must not violate)

| C1 | **Qdrant** is the vector database; collection name **`meridian_projects`**. |
| C2 | Embedding width **768** (e.g. nomic-embed-text) or **1536** (e.g. text-embedding-3-small) must stay **consistent** between ingest and collection schema. |
| C3 | Browser extension (**v0.2**): **github.com only**; no “monitor all tabs”. |
| C4 | **No Python** in the CLI codebase. |

---

## 8. Acceptance criteria (smoke)

Given a fresh machine with Docker, Go, and (for local embeddings) Ollama:

1. `make qdrant` → Qdrant answers on `6333`/`6334`.
2. Config + `meridian ingest --limit 30` (non-dry) → points exist in `meridian_projects`.
3. `meridian find --skills "Go"` → tabular output; if LLM available, JSON rerank path used; else similarity fallback.
4. `meridian explore --category observability` → non-empty when ingest included matching categories.
5. `MERIDIAN_GITHUB_TOKEN` optional but **strongly recommended** to avoid rate limits.

---

## 9. Open requirements (tracked for future PRs)

| ID | Gap |
|----|-----|
| O1 | **`init`**: full Bubbletea **multi-step** profile capture + **Save** to `~/.meridian/profile.yaml`. |
| O2 | **Agent**: explicit ReAct loop module vs one-shot rerank (see [roadmap.md](roadmap.md)). |
| O3 | **Extension**: Vite + **CRXJS**, sidebar UI, **local API** on **7474**. |
| O4 | **Hybrid search** / sparse BM25 in Qdrant (specified in vision; not required for v0.1 smoke). |
| O5 | **golangci-lint** in CI; `make lint` assumes local install. |

---

## 10. Document map

- [project.md](project.md) — architecture & file map  
- [roadmap.md](roadmap.md) — phased plan  
- [state.md](state.md) — implementation snapshot  
- [.planning/research/](.planning/research/) — external references & decisions context  
