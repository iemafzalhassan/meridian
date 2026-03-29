# Open questions

Items that are **intentionally undecided** or need product/engineering judgement. Resolve by PR + update this file.

| ID | Topic | Question | Candidates / notes |
|----|-------|----------|---------------------|
| OQ1 | Local API security | How should the **:7474** daemon authenticate the browser extension? | Shared secret file; OS keychain; short-lived token |
| OQ2 | Index versioning | Do we need **schema_version** in payload for migrations? | Bump collection name vs in-payload version |
| OQ3 | Good first issues | Is GitHub **search** total accurate enough vs GraphQL `label:` queries? | Rate limits vs accuracy |
| OQ4 | Hybrid search | When to add **sparse / BM25** in Qdrant vs dense-only v0.1? | Product priority vs complexity |
| OQ5 | Profile PII | Store GitHub handle in plaintext only, or hash for telemetry? | Privacy policy |
| OQ6 | Distribution | **Homebrew**, **scoop**, or raw GitHub releases first? | Maintainer capacity |
| OQ7 | Evaluation | What is the **golden set** of (skills → expected repos) for regression? | Curated CSV + CI job |

**Process:** when closing a row, link the decision doc or PR and **delete or archive** the row to avoid stale planning debt.
