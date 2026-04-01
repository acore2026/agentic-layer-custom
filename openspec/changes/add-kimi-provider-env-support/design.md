## Context

The current system relies on hardcoded LLM configurations and Gemini. There is no standard way to load environment variables from a file, making deployment and local development less secure and flexible.

## Goals / Non-Goals

**Goals:**
- Provide a Moonshot AI (Kimi) model provider implementation.
- Standardize configuration via `.env` files and environment variables.
- Ensure the Kimi provider is compatible with the existing `adk-go` agent logic.

**Non-Goals:**
- Porting all possible LLM providers (only Kimi for now).
- Implementing complex secret management services.

## Decisions

- **Kimi Implementation**: Use an OpenAI-compatible client structure to communicate with Moonshot AI. Since `adk-go` doesn't provide a native OpenAI model, we'll implement the `model.LLM` interface directly in a new package `pkg/model/kimi`.
- **Configuration Library**: Use `github.com/joho/godotenv`. It is the industry standard for Go and simple to integrate.
- **Provider Switching**: Modify `cmd/agent-gateway/main.go` to select the LLM provider based on an environment variable (`LLM_PROVIDER`).

## Risks / Trade-offs

- **[Risk] API Incompatibility** → **Mitigation**: Moonshot AI is OpenAI-compatible, so we will use established patterns for the API request/response format.
- **[Risk] Missing .env file** → **Mitigation**: The system will gracefully continue if `.env` is missing but required variables are present in the environment.
- **[Risk] Duplicate model logic** → **Mitigation**: We will try to make the provider generic if possible, but for this PoC, a dedicated `kimi` package is clearer.
