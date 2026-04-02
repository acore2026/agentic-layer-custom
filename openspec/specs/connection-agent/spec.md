## ADDED Requirements

### Requirement: ReAct Loop Execution
The Connection Agent SHALL implement a ReAct (Reason + Act) loop using `google/adk-go` and use dynamically loaded skill instructions for its orchestration logic. It SHALL be compatible with the ADK Launcher to ensure multi-step signaling sequences are captured and visualized in the web UI.

#### Scenario: Dynamic Skill Execution
- **WHEN** the agent receives an intent like "initial registration for UE-01"
- **THEN** the agent SHALL reason based on the loaded `init-registration` skill and call the corresponding tool sequence, and each step SHALL be visible in the web UI trace log

### Requirement: Blocking Tool Invocation
Registered Go API tools MUST be strictly blocking and return a success or failure string back to the LLM.

#### Scenario: Tool call returns result to LLM
- **WHEN** the `Issue_Access_Token_tool` is called
- **THEN** it waits for the mock signaling to complete and returns the token string as the tool's output to the LLM

### Requirement: LLM State Extraction and Management
The Connection Agent SHALL delegate all state extraction and passing between tools to the LLM, following the instructions provided in the skill definitions.

#### Scenario: State passing from skill instructions
- **WHEN** a skill instruction says to pass a token from Tool A to Tool B
- **THEN** the LLM SHALL extract the token from Tool A's output and provide it as an argument to Tool B
