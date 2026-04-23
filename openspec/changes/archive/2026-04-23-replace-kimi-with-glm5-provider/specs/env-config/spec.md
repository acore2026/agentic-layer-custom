## ADDED Requirements

### Requirement: Default GLM-5 Backend Selection
The system SHALL default provider selection to `glm5` when the repository is configured for its standard runtime path and no alternative backend is explicitly chosen.

#### Scenario: Default backend is glm5
- **WHEN** the gateway starts with the repository's default provider configuration
- **THEN** it SHALL select `glm5` as the backend used for model initialization

## MODIFIED Requirements

### Requirement: Configuration via Environment Variables
The system SHALL prioritize environment variables for configuring LLM providers, including API keys, model names, base URLs, and provider-specific runtime flags.

#### Scenario: Using GLM configuration from environment
- **WHEN** the `GLM_API_KEY` environment variable is set
- **THEN** the GLM-5 provider SHALL use `GLM_API_KEY`, `GLM_BASE_URL`, and `GLM_MODEL` to authenticate and configure OpenAI-compatible requests
