## ADDED Requirements

### Requirement: Intent Categorization and Routing
The System Agent SHALL act as an LLM router that receives raw natural language text and categorizes it to route to the appropriate Worker Agent.

#### Scenario: Successful routing to Connection Agent
- **WHEN** the System Agent receives a natural language string like "I want to create a PDU session for my UE"
- **THEN** the LLM categorizes the intent and calls the mock Connection Agent function with the raw text

### Requirement: Clarification Request on Ambiguity
The System Agent SHALL NOT fallback programmatically if the intent is highly ambiguous or the LLM fails to output a valid routing target. It MUST return a message asking the user for clarification.

#### Scenario: Ambiguous intent received
- **WHEN** the System Agent receives an ambiguous string like "do something"
- **THEN** the system returns a message asking the UE/App for clarification instead of routing to a worker
