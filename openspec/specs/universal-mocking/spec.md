## ADDED Requirements

### Requirement: Universal Mock Tool
The system SHALL provide a generic tool implementation that can be registered under any name and log all calls for sequence verification.

#### Scenario: Log generic tool call
- **WHEN** the universal mock tool is registered as `Auth_tool` and called by the agent
- **THEN** it SHALL log "[MOCK] Calling Auth_tool with args: ..." to the standard output

### Requirement: Flexible Mock Results
The universal mock tool SHALL return success responses with common fields that the LLM can use for subsequent state extraction.

#### Scenario: Return generic success result
- **WHEN** a tool is called by the agent
- **THEN** it SHALL return a map containing `status: SUCCESS` and a generic `token` string
