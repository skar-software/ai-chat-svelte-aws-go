# aichat_go

Go API for the AI chat widget (Fiber v3, OpenAI streaming).

## Run

```bash
cp .env.example .env   # if present; set OPENAI_API_KEY
go run ./cmd/api
```

## Tests

```bash
go test ./...
```

With coverage:

```bash
go test ./... -cover
```

Packages under test:

| Package | Focus |
|--------|--------|
| `internal/chat` | `StreamParser`, `Service` (mock provider) |
| `internal/store` | In-memory conversations and messages |
| `internal/config` | `Load()` and env defaults |
| `internal/api` | HTTP handlers, SSE stream, tenant middleware, error handler |
| `internal/providers/openai` | Intent routing, structured parts, deterministic stream |
| `internal/observability` | Request logger + request ID |

The OpenAI provider’s live HTTP calls are **not** exercised in unit tests; deterministic modes (artifact/plan/etc.) are covered without the network.
