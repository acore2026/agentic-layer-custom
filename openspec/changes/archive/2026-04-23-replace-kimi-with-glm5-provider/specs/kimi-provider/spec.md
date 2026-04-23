## REMOVED Requirements

### Requirement: Kimi Model Provider
**Reason**: The repository is switching its supported OpenAI-compatible backend from Kimi to GLM-5.
**Migration**: Configure `LLM_PROVIDER=glm5` and replace `KIMI_API_KEY`, `KIMI_BASE_URL`, and `KIMI_MODEL` with `GLM_API_KEY`, `GLM_BASE_URL`, and `GLM_MODEL`.

### Requirement: Content Generation via Kimi API
**Reason**: Runtime requests will no longer target the Moonshot-specific Kimi endpoint in the supported repository configuration.
**Migration**: Route agent traffic through the GLM-5 provider configuration and normalized DashScope-compatible endpoint instead.
