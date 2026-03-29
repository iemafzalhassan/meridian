# Data contracts (landscape → ProjectDoc → Qdrant)

Precise field and text rules for ingest, search, and LLM prompts.

---

## 1. CNCF landscape.yml walk

Implemented in `internal/ingestion/cncf.go`:

- Root key **`landscape`**: array of entries.
- Each entry has **`category`** map with:
  - `name` → Meridian **`category`**
  - `subcategories` → array of maps with **`subcategory`**:
    - `name` → Meridian **`sub_category`**
    - `items` → array of maps with **`item`**:
      - `name`, `description`, `homepage_url`, `repo_url`

**Inclusion rule:** `repo_url` must exist and contain **`github.com`** (case-insensitive). Others skipped.

---

## 2. `ProjectDoc` (application schema)

Go struct: `internal/vectorstore/types.go`

| Field | Source |
|-------|--------|
| `id` | Typically repo URL (set in pipeline from landscape) |
| `name` | Item `name` |
| `description` | Item `description` |
| `category` | Category `name` |
| `sub_category` | Subcategory `name` |
| `repo_url` | Item `repo_url` (`.git` stripped in ingest paths where applicable) |
| `homepage_url` | Item `homepage_url` |
| `languages` | GitHub `ListLanguages` keys, sorted by size descending |
| `topics` | GitHub repo topics |
| `stars` | `stargazers_count` |
| `good_first_issues` | Search API total for label `good first issue` |
| `contributor_count` | Approximation from contributors list pagination |
| `open_issues_count` | GitHub repo field |
| `last_commit_date` | Pushed-at timestamp (RFC3339 string) |
| `embed_text` | See §3 |

**Serialization:** payload stored in Qdrant is JSON-compatible map derived from struct (see `ProjectDocPayload`).

---

## 3. `embed_text` construction (required format)

Implemented in `internal/ingestion/embedtext.go`:

```text
{Name}. {Description}. Category: {Category}.
Technologies: {Languages joined with ", "}. Topics: {Topics joined with ", "}.
```

Whitespace: outer `strings.TrimSpace` on name/description/category; joins use `", "`.

Any change here affects **vector geometry** — require **re-embed** or **versioned** collection when experimenting.

---

## 4. Qdrant payload & vector

| Aspect | Contract |
|--------|----------|
| Collection | `meridian_projects` |
| Vector | Dense `[]float32`, length = `qdrant.embedding_dims` |
| Distance | Cosine |
| Point ID | Deterministic from repo URL (FNV-1a 64-bit) |

Payload must remain **JSON-serializable** fields only (no funcs, no cyclic structs).

---

## 5. `find` — candidate DTO to LLM

`internal/agent/find.go` builds a **reduced** JSON list for the model:

- `name`, `description` (truncated), `category`, `sub_category`, `repo_url`, `stars`, `good_first_issues`, `languages`, `topics`, `contributor_count`

**Rerank output contract (model):**

JSON array of objects:

```json
{
  "name": "string",
  "reason": "string",
  "good_first_issues": 0,
  "repo_url": "string",
  "score": 0.0
}
```

Parser: `encoding/json` after stripping optional ``` fences (`vectorstore.StripJSONFences`).

---

## 6. `explore` — category matching

- **Input:** user string compared to **`category + " " + sub_category`**, case-folded, **substring** match.
- **Ranking:** by **`stars`** descending among matches (current behavior).

---

## 7. Profile file (for `recommend`)

Path: `~/.meridian/profile.yaml`

```yaml
name: ""
github: ""
skills: []
level: ""        # beginner | intermediate | advanced (conventional)
interests: []
last_update: ""
```

Loading: `internal/profile/profile.go`.  
**Gap:** `meridian init` does not yet persist this file automatically — see [state.md](../../state.md).
