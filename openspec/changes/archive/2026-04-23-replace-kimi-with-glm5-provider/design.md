## Context

The repository currently splits provider logic across two paths: `cmd/agent-gateway/main.go` instantiates a dedicated Kimi model, while `pkg/workshop/orchestrator.go` already uses a generic OpenAI-compatible client for non-Kimi providers. The requested migration is cross-cutting because it changes runtime defaults, environment-variable names, documentation, and provider-specific tests. The supplied GLM-5 configuration also introduces a base-URL normalization requirement so both a raw API base and a full `/chat/completions` endpoint can be accepted safely.

## Goals / Non-Goals

**Goals:**
- Make `glm5` the supported configured backend for gateway and workshop execution paths.
- Replace Kimi-specific environment and documentation defaults with `GLM_*` settings.
- Normalize GLM base URLs consistently so the runtime accepts either a base path or a full chat-completions URL.
- Preserve the current OpenAI-compatible integration pattern, including configurable reasoning/thinking behavior.

**Non-Goals:**
- Building a multi-provider abstraction layer beyond what is needed for GLM-5.
- Adding new UI behavior or changing agent orchestration semantics unrelated to model selection.
- Keeping Kimi as a first-class documented runtime option in this repository.

## Decisions

### Use GLM-5 as the default OpenAI-compatible backend
The runtime will default to `LLM_PROVIDER=glm5` and use GLM-specific settings (`GLM_API_KEY`, `GLM_BASE_URL`, `GLM_MODEL`, and optional thinking configuration) for initialization. This matches the deployment target and removes ambiguity in local startup.

Alternative considered: keep the current `kimi` default and add GLM-5 as another option. Rejected because it would preserve the wrong operational default and keep the existing config/documentation mismatch.

### Reuse or adapt the existing OpenAI-compatible client instead of building a new bespoke transport
GLM-5 is provided through an OpenAI-compatible endpoint, and the workshop path already uses `pkg/model/openai_compatible.go`. The implementation should converge on one normalization and request-shaping path where practical, rather than maintaining separate Kimi-only logic for the gateway.

Alternative considered: fork the Kimi package into a separate GLM-only provider. Rejected because it duplicates nearly identical HTTP and tool-call handling.

### Normalize configured base URLs before request dispatch
Provider construction will strip a trailing `/chat/completions` suffix when necessary and then build requests against a single normalized base. This allows operators to supply either `https://dashscope.aliyuncs.com/compatible-mode/v1` or the full chat-completions URL without breaking calls.

Alternative considered: require one exact URL format. Rejected because the supplied deployment configuration already uses the full endpoint, while other code paths may expect a base URL.

## Risks / Trade-offs

- [Risk] GLM-5 may differ from Kimi in reasoning or tool-call response shape. → Mitigation: keep provider tests focused on content parsing, tool-call decoding, and base-URL normalization.
- [Risk] Renaming environment variables can break existing local setups. → Mitigation: update `.env.example`, docs, and migration guidance in the removed Kimi spec requirements.
- [Risk] Gateway and workshop flows could drift if they continue using different provider constructors. → Mitigation: centralize GLM-specific normalization and defaults in `pkg/model/`.

## Migration Plan

1. Add the GLM-5 provider configuration surface and default it in the gateway and workshop startup paths.
2. Update docs and `.env.example` to advertise `glm5` and `GLM_*` variables.
3. Replace or retire Kimi-specific tests with GLM-focused coverage for initialization and URL normalization.
4. Validate `go test ./...` and a local `go run ./cmd/agent-gateway` startup using GLM configuration.

Rollback: restore the previous Kimi-specific provider wiring and `.env.example` entries if GLM-5 compatibility fails during validation.

## Open Questions

- Should the repository keep dormant Kimi code behind an unsupported path, or remove it entirely in the implementation phase?
- Should thinking be controlled by a dedicated `GLM_ENABLE_THINKING` variable or by existing request-level reasoning flags only?
