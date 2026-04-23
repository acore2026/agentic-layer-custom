## Why

The repository **currently** documents and defaults to Kimi-specific provider configuration, but the deployment target now needs GLM-5 instead. We need the runtime, configuration surface, and specs to reflect GLM-5 as the supported OpenAI-compatible backend so local runs and production wiring stay aligned.

## What Changes

- Replace the Kimi-first runtime path with a GLM-5 provider path for the gateway and workshop flows.
- Add GLM-5 configuration defaults for API key, base URL normalization, model name, and thinking behavior.
- Update environment-variable documentation and startup behavior to prefer `LLM_PROVIDER=glm5` and `GLM_*` settings.
- **BREAKING** Remove Kimi-specific configuration as the primary documented provider path for this repository.

## Capabilities

### New Capabilities

- `glm5-provider`: Support GLM-5 as the repository's OpenAI-compatible model provider, including default endpoint handling and model initialization.

### Modified Capabilities

- `env-config`: Change provider-related environment variables and examples from `KIMI_*` to `GLM_*`, and document GLM-5 as the default backend.
- `kimi-provider`: Retire the current Kimi-specific runtime requirement so repository behavior no longer depends on Moonshot-specific configuration.



## Impact

- Changes to `cmd/agent-gateway/main.go` and `pkg/workshop/orchestrator.go` for provider selection and defaults.
- Updates in `pkg/model/` to support GLM-5 endpoint normalization and provider construction.
- Refresh of `.env.example`, contributor docs, and provider-facing tests.
- New spec deltas for GLM-5 provider support and environment configuration changes.
