# Meridian — Roadmap

Phased delivery aligned with the original build order. Status snapshots belong in [state.md](state.md); **this file is the forward plan**.

---

## Phase 0 — Bootstrap (done)

- Go module, repo layout, Apache-2.0, Makefile, Compose for Qdrant, example config, README.

## Phase 1 — Foundation (done)

- Cobra root + version, Viper config (`~/.meridian/config.yaml`), Zap, dependency injection pattern.
- Scaffold all internal packages and extension tree.

## Phase 2 — Data pipeline (largely done; iterate)

**Goals**

- Reliable CNCF ingest, GitHub enrichment with **rate-limit awareness**, EmbedText builder, Qdrant **create + upsert**.
- `meridian ingest --limit` / `--dry-run`.

**Enhancements (next)**

- [ ] Incremental ingest / idempotent updates by `repo_url` + metadata hash.
- [ ] Optional **full-text index** on `category` for server-side `explore` filters (vs scroll + substring).
- [ ] Metrics: ingest duration, skip reasons, GitHub 403/429 counters.

## Phase 3 — LLM + embeddings (done; harden)

**Goals**

- `llm.Provider` + `embeddings.E Embedder` factories; Ollama + OpenRouter via langchaingo.
- Config-only provider selection.

**Enhancements (next)**

- [ ] Central HTTP transport with shared retry/backoff policy per provider.
- [ ] **Embedding validation** on startup (`probe` dimension vs `qdrant.embedding_dims`).
- [ ] Document model pairs (chat + embed) per provider in docs.

## Phase 4 — Agent core (partial → complete)

**Done today**

- `find`: embed query → Qdrant search → LLM JSON rerank → table; robust fallback.

**Planned**

- [ ] Dedicated **ReAct** loop module (`internal/agent`): tool boundaries (search, fetch_issue, summarize).
- [ ] Structured output validation (JSON schema / repair pass).
- [ ] Optional **JSON mode** flags per provider for lower parse failure rate.

## Phase 5 — CLI polish (partial)

**Done today**

- Lipgloss tables, Glamour in `init`, `explore` limits.

**Planned**

- [ ] **Full `init` TUI**: name, GitHub handle, skills, level, interests → **`~/.meridian/profile.yaml`**.
- [ ] **README preview** for a repo (Glamour) in CLI (`meridian show <repo>`?) — optional.
- [ ] Config doctor: `meridian doctor` checks Qdrant, Ollama, GH token, dims.

## Phase 6 — Browser extension + local API (not started)

**Requirements**

- TypeScript extension, **Vite + CRXJS**, **React + Tailwind**.
- **github.com** host permission only; sidebar panel.
- Local **Meridian API** (Go) on **port 7474**: profile, fit score, issues snippet, LLM “why you fit”.

**Deliverables**

- [ ] OpenAPI or minimal route spec for extension ↔ daemon.
- [ ] Authn for local server (even if loopback-only token).
- [ ] Packaged build instructions (Chrome Web Store later; dev load unpacked first).

---

## Dependency graph (high level)

```text
Phase2 (ingest/index) ──► Phase4 (agent/find quality)
        │
        └──────────────► Phase6 (extension needs stable API + index)

Phase3 (LLM/embed) ───► Phase4/5 (better explanations + UX)
```

---

## Milestone suggestions (calendar-agnostic)

| Milestone | Outcome |
|-----------|---------|
| **M1 — Usable CLI** | Ingest + find + explore + recommend stable on fresh laptop. |
| **M2 — Profile** | Full init + profile schema versioned. |
| **M3 — Agent v2** | ReAct + tools + evaluation fixture set. |
| **M4 — Extension α** | Sidebar + local API + single happy-path demo repo. |

---

## Cross-references

- [requirement.md](requirement.md) — normative requirements  
- [state.md](state.md) — what exists in the repo **today**  
- [.planning/research/README.md](.planning/research/README.md) — ecosystem notes  
