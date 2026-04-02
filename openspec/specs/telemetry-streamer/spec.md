## ADDED Requirements

### Requirement: Agent Activity Capture
The system SHALL capture all prompt requests and LLM responses from both `SystemAgent` and `ConnectionAgent` and emit them as `ai_payload` events.

#### Scenario: System Agent routing decision
- **WHEN** the `SystemAgent` decides to route to `ConnectionAgent`
- **THEN** the system SHALL emit an `ai_payload` event containing the routing message

### Requirement: LLM Thought Chunking
The system SHALL detect `Thought` parts in the LLM stream and emit them as `llm_thought` events for real-time visualization of the reasoning process.

#### Scenario: Reasoning chunk emitted
- **WHEN** the LLM provider yields a part with `Thought: true`
- **THEN** the system SHALL emit an `llm_thought` event with the reasoning text
