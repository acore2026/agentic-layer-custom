## ADDED Requirements

### Requirement: Skills Retrieval Endpoint
The system SHALL provide a `/v1/skills` endpoint that returns a list of all available network signaling skills.

#### Scenario: Successful skills retrieval
- **WHEN** a client sends a GET request to `/v1/skills`
- **THEN** the system SHALL return a 200 OK response with a JSON body containing a list of skills, each including `id`, `name`, `description`, and `definition` (raw Markdown).

### Requirement: Skill Definition Loading
The system SHALL dynamically load skill definitions from the `skill/` directory by parsing existing Markdown files.

#### Scenario: Skill discovery from filesystem
- **WHEN** the endpoint is called
- **THEN** it SHALL scan the `skill/` directory for subdirectories containing `SKILL.md` files and populate the response with their metadata and raw content.
