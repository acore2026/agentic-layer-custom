## ADDED Requirements

### Requirement: Streaming WebSocket Endpoint
The system SHALL provide a WebSocket endpoint at `/v1/intents/stream` to handle asynchronous intent execution.

#### Scenario: Establish connection and send intent
- **WHEN** a client connects to `/v1/intents/stream` and sends an `execute_intent` message
- **THEN** the system SHALL start the agent orchestration and stream back events

### Requirement: Structured Event Formats
The system SHALL emit events using the specific JSON formats defined in the API Design document, including `ai_payload`, `llm_thought`, `network_pcap`, and `workflow_complete`.

#### Scenario: Workflow completion event
- **WHEN** the agent orchestration finishes
- **THEN** the system SHALL send a `workflow_complete` message with `status: success` and the final summary
