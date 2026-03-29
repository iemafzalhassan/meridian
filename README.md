# Meridian

Find your place in open source.

Meridian is an AI-powered open source contribution intelligence toolkit with two surfaces:
- `v0.1`: terminal-first CLI agent (single cross-platform Go binary)
- `v0.2`: GitHub-tab-only browser extension that calls a local Meridian API

## Quickstart

**Repository:** [github.com/iemafzalhassan/meridian](https://github.com/iemafzalhassan/meridian)

### Prerequisites
- Go 1.22+
- Docker (for local Qdrant)

### 1) Clone and build
```bash
git clone https://github.com/iemafzalhassan/meridian.git
cd meridian
make build
```

### 2) Start Qdrant locally
```bash
make qdrant
```

### 3) Configure Meridian
```bash
mkdir -p ~/.meridian
cp .meridian.yaml.example ~/.meridian/config.yaml
```

### 4) Ingest CNCF landscape into Qdrant (required before `find` / `explore`)
```bash
make ingest
# or: ./bin/meridian ingest --limit 50
```

Ollama (or OpenRouter for embeddings) must be running and match `qdrant.embedding_dims` (768 for `nomic-embed-text`, 1536 for `text-embedding-3-small`).

### 5) Run CLI
```bash
./bin/meridian --help
./bin/meridian find --skills "Go, Kubernetes"
./bin/meridian explore --category observability --limit 15
```

## Troubleshooting

| Symptom | Cause | What to do |
|--------|--------|------------|
| `find` / `explore` used to print `not implemented` | Phase 1 stubs | Rebuild: `make build`; commands are implemented end-to-end now. |
| `command not found: meridian` | Binary not on `PATH` | Run `./bin/meridian` from the repo, or `export PATH="$PWD/bin:$PATH"`, or install the binary where your shell looks. |
| `No projects in Qdrant yet` | Empty index | Start Qdrant (`make qdrant`), configure LLM/embeddings, then `meridian ingest --limit 50`. |
| Connection errors to Qdrant | Wrong URL or daemon down | Check `qdrant.url` (gRPC, usually port **6334**) and `docker compose ps`. |

## Configuration Reference

Meridian reads configuration from `~/.meridian/config.yaml` with environment variable overrides.

### Core settings
- `llm.provider`: `ollama` or `openrouter`
- `llm.model`: chat/completion model
- `llm.ollama_url`: local Ollama endpoint
- `llm.openrouter_key`: API key for OpenRouter
- `embeddings.provider`: `ollama` or `openrouter`
- `embeddings.model`: embedding model (`nomic-embed-text` by default)
- `qdrant.url`: Qdrant gRPC endpoint
- `qdrant.collection`: must be `meridian_projects`

### Environment overrides
- `MERIDIAN_OPENROUTER_KEY`
- `MERIDIAN_OLLAMA_URL`
- `MERIDIAN_GITHUB_TOKEN`

### Security best practices
- Never commit real API keys in tracked files.
- Keep credentials in environment variables or untracked user config only.
- Treat developer profile/form data as personal data; avoid logging raw secrets or sensitive identifiers.

## CLI Commands

- `meridian init`
- `meridian find --skills "Go,gRPC,Kubernetes"`
- `meridian explore --category observability`
- `meridian recommend`
- `meridian ingest --limit 100 --dry-run`

## Contributing Guide

1. Fork and clone the repository.
2. Build and run tests:
   ```bash
   make build
   make test
   ```
3. Start local dependencies:
   ```bash
   make qdrant
   ```
4. Keep interfaces provider-agnostic and configuration-driven.
5. Wrap errors with context (`fmt.Errorf("domain: %w", err)`).
6. Include tests for behavior changes and avoid introducing global mutable state.

## License

Apache License 2.0. See `LICENSE`.
