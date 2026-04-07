## ADDED Requirements

### Requirement: Skill Workshop WebSocket Endpoint
The system SHALL provide a WebSocket endpoint at `/ws/agent-run` to handle the asynchronous generation and validation of 6G signaling skills (procedures).

#### Scenario: Successful skill workshop connection
- **WHEN** a client connects to `/ws/agent-run` and sends a `start_run` message
- **THEN** the system SHALL initialize the skill generation orchestrator and stream telemetry events.
