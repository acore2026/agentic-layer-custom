## ADDED Requirements

### Requirement: Real-Time Event Tracing
The system SHALL capture and stream agent events (e.g., tool calls, responses, reasoning steps) to the web dashboard for real-time visualization.

#### Scenario: Visualize Tool Call
- **WHEN** an agent executes a tool (e.g., `Auth_tool`)
- **THEN** the web dashboard SHALL display the tool name, input arguments, and execution status in the trace log

### Requirement: Chain of Thought Visualization
The system SHALL stream "thinking" or reasoning parts of LLM responses to the web dashboard to visualize the agent's internal decision-making process.

#### Scenario: Display Reasoning
- **WHEN** an agent produces a response part with `Thought: true`
- **THEN** the web dashboard SHALL display this content as a distinct reasoning step in the UI
