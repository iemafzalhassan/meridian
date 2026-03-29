# Integration notes (Qdrant, GitHub, LLMs)

How Meridian connects to external systems and what to watch for when changing code.

---

## 1. Qdrant

### Client & transport

- Go package: **`github.com/qdrant/go-client/qdrant`**
- **gRPC** address derived from `qdrant.url` in config (see `internal/qdrantutil` and `internal/vectorstore/qdrant.go`).
- Default local URL in example config: `http://localhost:6334` (gRPC port).

### Collection

- Name: **`meridian_projects`** (config key `qdrant.collection`).
- Vectors: **single dense** vector per point; **Cosine** distance.
- **Size** must match embedding model output (`qdrant.embedding_dims`).

### Writes

- **Upsert** with `Wait: true` after batching points.
- Payload is built from `ProjectDoc` via JSON → `map[string]any` → Qdrant `TryValueMap` (see `internal/vectorstore/payload.go`).

### Reads

- **Search:** `QueryPoints` with `NewQueryDense(embedding)`; optional post-filter (implementation may filter in memory).
- **Explore:** `ScrollPoints` with pagination via `next_page_offset` (see `ScrollCategory`).

### Operational tips

- If ingest fails with dimension mismatch, fix **`embeddings.model`** vs **`qdrant.embedding_dims`** before re-ingesting.
- Deleting the collection resets all points; keep **backup** or versioned collection names only with a migration plan.

---

## 2. GitHub API

### Client

- **`github.com/google/go-github/v67/github`**
- Auth: optional **`MERIDIAN_GITHUB_TOKEN`** → `github.token` in config.
- HTTP client timeout from `github.timeout`.

### Endpoints used (conceptually)

| Concern | Approach |
|---------|----------|
| Repo summary | `Repositories.Get` |
| Languages | `Repositories.ListLanguages` |
| Good first issues | `Search.Issues` with `repo:owner/name is:open label:"good first issue"` |
| Contributors | `Repositories.ListContributors` (paginated, capped) |

### Rate limiting

- **Search API** is stricter than core REST.
- Code should **retry** and optionally **sleep** on rate-limit / abuse errors (see `internal/ingestion/github.go`).

### Engineering note

Contributor count is **approximate** (pagination cap). Product copy should not claim exact org-wide contributor totals.

---

## 3. Embeddings providers

### Interface

- `internal/embeddings/embedder.go`: **`Embedder`** with `Embed(ctx, text) ([]float32, error)`.

### Ollama

- Config: `embeddings.provider: ollama`, `embeddings.model` (e.g. `nomic-embed-text`).
- Uses **langchaingo** `llms/ollama` and `CreateEmbedding`.
- Server URL: **`llm.ollama_url`** (same host as chat Ollama in current config design).

### OpenRouter

- Config: `embeddings.provider: openrouter`.
- Uses **langchaingo** `llms/openai` with:
  - `WithBaseURL("https://openrouter.ai/api/v1")`
  - `WithToken` from **`llm.openrouter_key`** / `MERIDIAN_OPENROUTER_KEY`
  - `WithModel` / `WithEmbeddingModel` from **`embeddings.model`**

### Dimension policy

| Model (typical) | Dims |
|-----------------|------|
| `nomic-embed-text` | **768** |
| `text-embedding-3-small` | **1536** |

`qdrant.embedding_dims` **must** match.

---

## 4. Chat / completion providers

### Interface

- `internal/llm/provider.go`: **`Provider`** with `Complete` and `Embed` (OpenRouter/Ollama implement both for interface symmetry; `find` uses **embeddings** package for query vectors).

### Ollama

- `llms/ollama` + `Call` for completion.

### OpenRouter

- `llms/openai` + `Call` with OpenRouter base URL.

### Rerank prompt

- System + user text concatenated for **`find`** (see `internal/agent/find.go` + `internal/agent/prompt.go`).
- Expected model output: **JSON array** of objects with `name`, `reason`, `good_first_issues`, `repo_url`, `score`.

---

## 5. HTTP for landscape only

- **`github.com/go-resty/resty/v2`** in `internal/ingestion/cncf.go` for downloading `landscape.yml`.
- Retries delegated to `internal/retry` where used.

---

## 6. Future: extension ↔ local API (v0.2)

| Concern | Planned |
|---------|---------|
| Port | **7474** (convention from product brief) |
| Auth | Loopback token or OS keychain (TBD — see [open-questions.md](open-questions.md)) |
| Payloads | Reuse `ProjectDoc` + search scores; add issue snippets |

No server implementation exists in-repo yet; extension is scaffold-only.
