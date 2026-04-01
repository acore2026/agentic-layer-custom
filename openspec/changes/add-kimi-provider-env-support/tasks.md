## 1. Environment Configuration

- [x] 1.1 Add `github.com/joho/godotenv` dependency: `go get github.com/joho/godotenv`
- [x] 1.2 Create `.env.example` in the project root
- [x] 1.3 Update `cmd/agent-gateway/main.go` to load `.env` at startup

## 2. Kimi Provider Implementation

- [x] 2.1 Create package `pkg/model/kimi`
- [x] 2.2 Implement `model.LLM` interface for Kimi (Moonshot AI)
- [x] 2.3 Add support for OpenAI-compatible request/response mapping in the Kimi provider

## 3. Gateway Integration

- [x] 3.1 Update `cmd/agent-gateway/main.go` to select provider based on `LLM_PROVIDER` env var
- [x] 3.2 Update agent initialization logic to use keys from environment variables
- [x] 3.3 Verify system with Kimi provider selected
