## 1. Provider Runtime Migration

- [x]  1.1 Add GLM-5 provider construction in `pkg/model/`, including base URL normalization for both base-path and `/chat/completions` inputs.
- [x]  1.2 Update `cmd/agent-gateway/main.go` to default to `glm5` and initialize the runtime from `GLM_API_KEY`, `GLM_BASE_URL`, and `GLM_MODEL`.
- [x]  1.3 Update `pkg/workshop/orchestrator.go` to use the GLM-5 configuration path consistently with the gateway flow.
- [x]  1.4 Remove or retire Kimi-specific runtime wiring that is no longer part of the supported repository configuration.

## 2. Configuration and Documentation

- [x]  2.1 Update `.env.example` to document `LLM_PROVIDER=glm5` and the required `GLM_*` variables.
- [x]  2.2 Refresh repository docs that currently describe Kimi as the active provider so they describe GLM-5 and its default endpoint/model behavior.
- [x]  2.3 Document any operator-facing migration notes for replacing `KIMI_*` settings with `GLM_*` settings.

## 3. Validation

- [x]  3.1 Replace or update provider tests to cover GLM-5 initialization and base URL normalization behavior.
- [x]  3.2 Run `go test ./...` and fix any regressions caused by the provider migration.
- [x]  3.3 Start the gateway locally with GLM-5 configuration and confirm the service boots with the new default provider path.
