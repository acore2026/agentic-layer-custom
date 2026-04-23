# Repository Guidelines

## Project Structure & Module Organization
`cmd/agent-gateway/` contains the main entrypoint that launches the ADK web UI, REST endpoints, and WebSocket handlers. Core application code lives under `pkg/`: `pkg/agents/` for routing agents, `pkg/api/` for HTTP/WebSocket surfaces, `pkg/workshop/` for skill-generation orchestration, `pkg/tools/` for mock signaling tools, and `pkg/model/` plus `pkg/observability/` for provider and tracing integrations. Runtime skills are stored as Markdown in `skill/*/SKILL.md`. Spec-driven work is tracked in `openspec/specs/` and archived change records under `openspec/changes/archive/`.

## Build, Test, and Development Commands
Use Go 1.26.1+.

- `go run ./cmd/agent-gateway` starts the gateway on `http://localhost:8080/ui/`.
- `API_PORT=9090 LLM_PROVIDER=glm5 GLM_API_KEY=... go run ./cmd/agent-gateway` runs the default provider on a custom port.
- `go test ./...` runs the full test suite across `pkg/...`.
- `go test ./pkg/agents ./pkg/workshop` is a useful focused pass when changing orchestration logic.
- `gofmt -w cmd pkg` formats the main Go source trees before review.

Copy `.env.example` to `.env` for local development. The code reads `API_PORT`, `LLM_PROVIDER`, `GLM_*`, `GOOGLE_API_KEY`, `GEMINI_MODEL`, `OPENAI_*`, and `LANGFUSE_*` from the environment. `LLM_PROVIDER=mock` remains useful for local routing tests without external model calls.

## Coding Style & Naming Conventions
Follow standard Go formatting and let `gofmt` decide indentation and spacing. Keep packages lowercase, exported identifiers in `CamelCase`, and tests in `*_test.go`. Prefer small, focused packages under `pkg/` and keep transport-specific code in `pkg/api/` instead of agent packages. For new skills, use a new directory such as `skill/my-flow/SKILL.md`.

## Testing Guidelines
Add unit tests next to the code they cover. Current tests use Go’s `testing` package and rely heavily on mock providers for deterministic behavior. Name tests by behavior, for example `TestSystemAgentRouting` or `TestServiceAgentDecorateContext`. Run `go test ./...` before opening a PR; changes to agent routing, APIs, or tool orchestration should include regression coverage.

## Commit & Pull Request Guidelines
Recent commits use imperative, sentence-style subjects such as `Fix port configuration by passing --port flag to ADK launcher`. Keep subjects concise, capitalized, and scoped to one change. PRs should summarize behavior changes, list required env/config updates, link the relevant issue or OpenSpec change when applicable, and include screenshots or sample payloads for `/ui`, REST, or WebSocket-facing changes.
