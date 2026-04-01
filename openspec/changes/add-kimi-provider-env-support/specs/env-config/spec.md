## ADDED Requirements

### Requirement: .env Support
The system SHALL support loading environment variables from a `.env` file using the `joho/godotenv` library. This allows for sensitive configuration like API keys to be managed externally from the code.

#### Scenario: .env file loading at startup
- **WHEN** the `agent-gateway` starts and a `.env` file is present in the root directory
- **THEN** it SHALL load the environment variables defined in the `.env` file

### Requirement: Configuration via Environment Variables
The system SHALL prioritize environment variables for configuring LLM providers, including API keys and model names.

#### Scenario: Using API keys from environment
- **WHEN** the `KIMI_API_KEY` environment variable is set
- **THEN** the Kimi provider SHALL use this key to authenticate with the Moonshot AI API
