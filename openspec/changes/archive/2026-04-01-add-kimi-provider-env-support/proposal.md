## Why

Adding a Kimi provider expands the system's LLM choices to include Moonshot AI, which is beneficial for users in specific regions. Additionally, .env support is needed for better configuration management and to avoid hardcoding API keys.

## What Changes

- Introduction of a **Kimi model provider** that implements the `model.LLM` interface.
- Support for **.env file loading** using `github.com/joho/godotenv`.
- Creation of a **.env.example** file to guide configuration.
- Integration of environment variables for API key management across agents.

## Capabilities

### New Capabilities
- `kimi-provider`: Implementation of the Kimi (Moonshot AI) model provider as an alternative to Gemini.
- `env-config`: Mechanism to load and manage configuration via environment variables and .env files.

### Modified Capabilities
- (None)

## Impact

- New package in `pkg/model/kimi`.
- Changes to `cmd/agent-gateway/main.go` for .env loading and provider selection.
- Dependency on `github.com/joho/godotenv`.
