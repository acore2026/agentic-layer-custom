## ADDED Requirements

### Requirement: Skill Generation Endpoint
The system SHALL provide a WebSocket endpoint at `/ws/agent-run` that accepts a `start_run` message to initiate the dynamic generation of a network signaling skill (procedure).

#### Scenario: Successful skill generation run
- **WHEN** a client establishes a WebSocket connection and sends a `start_run` JSON message with a valid prompt
- **THEN** the system SHALL stream a `run_started` event followed by telemetry from the `intent_analysis_agent`, `skill_writer_agent`, and `markdown_format_checker_agent`.

### Requirement: Multi-Agent Generation Workflow
The generation process SHALL utilize a coordinated sequence of specialized agents:
1.  **Intent Analysis Agent**: Categorizes the request and summarizes the workflow.
2.  **Skill Writer Agent**: Generates a Markdown draft of the signaling procedure.
3.  **Format Checker Agent**: Validates and repairs the draft based on the tool catalog.

#### Scenario: Orchestrated agent transitions
- **WHEN** the `intent_analysis_agent` completes its task
- **THEN** the system SHALL automatically transition to the `skill_writer_agent` and stream its progress to the client.

### Requirement: Iterative Format Correction
The system SHALL automatically invoke the `markdown_format_checker_agent` to validate and repair any Markdown draft that fails schema validation (e.g., invalid tool calls, missing YAML frontmatter). The system MUST support up to 3 repair attempts before failing the run.

#### Scenario: Automatic repair of invalid tool call
- **WHEN** the `skill_writer_agent` generates a draft containing a tool call not present in the catalog
- **THEN** the `markdown_format_checker_agent` SHALL be triggered with the specific validation error to provide a corrected version.

### Requirement: Tool Catalog API
The system SHALL provide a `GET /api/tools` endpoint that returns a normalized JSON list of all available network signaling tools and their parameter schemas, derived from the project's source of truth for tools.

#### Scenario: Retrieval of tool catalog
- **WHEN** a client performs a GET request to `/api/tools`
- **THEN** the system SHALL return a 200 OK response with a list of tools including their names, descriptions, and required/optional arguments.
