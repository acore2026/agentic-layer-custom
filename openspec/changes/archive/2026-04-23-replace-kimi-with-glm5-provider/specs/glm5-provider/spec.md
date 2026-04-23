## ADDED Requirements

### Requirement: GLM-5 Model Provider
The system SHALL provide a GLM-5 model provider for agent execution using an OpenAI-compatible API configuration. This provider SHALL be usable by the gateway and workshop flows when `LLM_PROVIDER` is set to `glm5`.

#### Scenario: Successful GLM-5 model initialization
- **WHEN** the GLM-5 provider is initialized with a valid API key and model name
- **THEN** it SHALL return a model instance ready to generate content for agent runs

### Requirement: GLM-5 Base URL Normalization
The system SHALL normalize configured GLM-5 base URLs so operators can supply either an API base path or a full `/chat/completions` endpoint without breaking requests.

#### Scenario: Full chat completions URL is normalized
- **WHEN** the GLM-5 provider is configured with a URL ending in `/chat/completions`
- **THEN** it SHALL normalize the value before issuing OpenAI-compatible chat completion requests

#### Scenario: Base path URL is accepted
- **WHEN** the GLM-5 provider is configured with a base URL that does not end in `/chat/completions`
- **THEN** it SHALL issue chat completion requests using the normalized OpenAI-compatible endpoint

### Requirement: GLM-5 Default Endpoint and Model
The system SHALL support GLM-5 defaults aligned with the repository deployment target, including the configured DashScope-compatible endpoint and `glm-5` model name when explicit overrides are not provided.

#### Scenario: Default GLM-5 configuration is used
- **WHEN** the runtime selects the `glm5` provider without explicit base URL or model overrides
- **THEN** it SHALL target the repository's default GLM-5 endpoint and model configuration
