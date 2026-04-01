## ADDED Requirements

### Requirement: Kimi Model Provider
The system SHALL provide a Kimi (Moonshot AI) model provider that implements the `model.LLM` interface. This allows the system to use Kimi as a reasoning engine for its agents.

#### Scenario: Successful Kimi model initialization
- **WHEN** the Kimi provider is initialized with a valid API key and model name
- **THEN** it SHALL return a `model.LLM` instance ready to generate content

### Requirement: Content Generation via Kimi API
The Kimi provider SHALL support generating content by communicating with the Moonshot AI API using the OpenAI-compatible endpoint.

#### Scenario: Agent runs with Kimi provider
- **WHEN** an agent is configured with the Kimi provider and receives a prompt
- **THEN** the Kimi provider SHALL call the Moonshot AI API and return the generated text response
