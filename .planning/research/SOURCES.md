# Canonical sources & versions

Values below reflect the **Meridian repo** as of the last documentation pass. **Verify** in [go.mod](../../go.mod) before quoting exact versions in external reports.

## Upstream data

| Resource | URL | Notes |
|----------|-----|--------|
| CNCF landscape (raw) | [https://raw.githubusercontent.com/cncf/landscape/master/landscape.yml](https://raw.githubusercontent.com/cncf/landscape/master/landscape.yml) | Ingested by `internal/ingestion/cncf.go` |
| CNCF landscape (human) | [https://landscape.cncf.io/](https://landscape.cncf.io/) | Context only |

## Runtime & infrastructure

| Component | Reference | Notes |
|-----------|-----------|--------|
| Qdrant (Docker) | [https://qdrant.tech/documentation/guides/installation/](https://qdrant.tech/documentation/guides/installation/) | Local: REST `6333`, gRPC `6334` (see [docker-compose.yml](../../docker-compose.yml)) |
| Qdrant Go client | Module `github.com/qdrant/go-client` | gRPC client used in `internal/vectorstore` |

## APIs

| API | Base / docs | Meridian usage |
|-----|-------------|----------------|
| GitHub REST | [https://docs.github.com/en/rest](https://docs.github.com/en/rest) | `google/go-github/v67`: repo metadata, languages, search (good first issues), contributors |
| Ollama | [https://github.com/ollama/ollama/blob/main/docs/api.md](https://github.com/ollama/ollama/blob/main/docs/api.md) | Via **langchaingo** Ollama LLM client |
| OpenRouter | [https://openrouter.ai/docs](https://openrouter.ai/docs) | OpenAI-compatible **`/v1`** base ** `https://openrouter.ai/api/v1`** |

## Go stack (direct dependencies — verify in go.mod)

Representative direct requires (not exhaustive):

- `github.com/spf13/cobra`
- `github.com/spf13/viper`
- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/lipgloss`
- `github.com/charmbracelet/glamour`
- `github.com/tmc/langchaingo`
- `github.com/qdrant/go-client`
- `github.com/google/go-github/v67`
- `github.com/go-resty/resty/v2`
- `go.uber.org/zap`
- `gopkg.in/yaml.v3`
- `github.com/stretchr/testify` (tests)

## Extension (planned v0.2)

| Piece | Reference |
|-------|-----------|
| Chrome MV3 | [https://developer.chrome.com/docs/extensions/mv3/intro/](https://developer.chrome.com/docs/extensions/mv3/intro/) |
| CRXJS + Vite | [https://crxjs.dev/vite-plugin/getting-started/react](https://crxjs.dev/vite-plugin/getting-started/react) | Not yet integrated in repo |
