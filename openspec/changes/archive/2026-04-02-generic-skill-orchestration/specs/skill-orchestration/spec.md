## ADDED Requirements

### Requirement: Dynamic Skill Loading
The system SHALL load all `SKILL.md` files from the `skill/` directory and include their contents in the Connection Agent's system prompt.

#### Scenario: Multiple skills loaded
- **WHEN** `skill/init-registration/SKILL.md` and `skill/acn/SKILL.md` exist
- **THEN** the Connection Agent's system prompt SHALL include the workflow instructions from both files

### Requirement: Automated Tool Discovery
The system SHALL parse all loaded `SKILL.md` files to identify unique tool names defined by the `CALL` pattern and register them with the agent.

#### Scenario: Discovery of unique tools
- **WHEN** `SKILL.md` files mention `CALL "Auth_tool"` and `CALL "Subscription_tool"`
- **THEN** the Connection Agent SHALL register both `Auth_tool` and `Subscription_tool` during initialization
