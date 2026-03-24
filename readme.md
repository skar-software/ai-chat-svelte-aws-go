# AI Chat Widget

A monorepo for a reusable AI chat widget with a Go Fiber streaming API and a Svelte 5 UI (Vercel AI).

## Demo video

<video src="./demo-min.mp4" controls muted playsinline width="100%"></video>

Direct link: [./demo.mp4](./demo.mp4)

## What's inside

- aichat_go — Go API (Fiber v3) with OpenAI streaming support
- aichat_ui — SvelteKit UI component library and demo app

See the detailed READMEs:

- [aichat_go/readme.md](aichat_go/readme.md)
- [aichat_ui/readme.md](aichat_ui/readme.md)

## Requirements

- Go 1.21+
- Node.js 18+
- pnpm 9+

## Quick start

### 1) API

From aichat_go:

```bash
cp .env.example .env   # if present; set OPENAI_API_KEY
go run ./cmd/api
```

### 2) UI

From aichat_ui:

```bash
pnpm install
pnpm dev
```

## Scripts

### API

```bash
go test ./...
```

### UI

```bash
pnpm lint
pnpm check
pnpm test
```

## License

See [LICENSE](LICENSE).
