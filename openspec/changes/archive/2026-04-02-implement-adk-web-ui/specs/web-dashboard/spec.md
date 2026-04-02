## ADDED Requirements

### Requirement: Browser-Based Interface
The system SHALL provide a web-based dashboard accessible via a local browser (e.g., `http://localhost:8080`) to interact with and monitor agents.

#### Scenario: Launch Web UI
- **WHEN** the user starts the gateway with the web UI enabled
- **THEN** the system SHALL start a local web server and log the access URL

### Requirement: Multi-Agent Selection
The web dashboard SHALL allow users to select from available agents (SystemAgent, ConnectionAgent) to initiate interactions.

#### Scenario: Select Agent in UI
- **WHEN** the dashboard is loaded in the browser
- **THEN** it SHALL display a dropdown or list of all registered workers for selection
