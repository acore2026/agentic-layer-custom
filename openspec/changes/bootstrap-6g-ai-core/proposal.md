## Why

Building a functional proof-of-concept for a 6G core network requires a decoupled AI layer to enable intent-driven networking. This change introduces a "Skill match and execution" architecture using Go 1.22+ and the `google/adk-go` framework to demonstrate how AI agents can manage network intents and execute signaling tools.

## What Changes

- Introduction of a **System Agent** as the intent gateway for routing natural language requests.
- Introduction of a **Connection Agent** utilizing the ReAct (Reason + Act) prompting approach to handle connection-related intents.
- Implementation of mock blocking Go API tools (`Issue_Access_Token_tool` and `Create_Subnet_PDUSession_tool`) for the Connection Agent.
- Scaffolding of the agentic layer using the `google/adk-go` framework.

## Capabilities

### New Capabilities
- `system-agent`: Acts as an LLM router to categorize and hand off raw natural language text to worker agents or request clarification.
- `connection-agent`: Executes a ReAct loop to process intents by calling registered Go API tools and managing state between them.

### Modified Capabilities
- (None)

## Impact

- New Go package for the 6G AI core.
- Dependency on `google/adk-go` and Gemini LLM provider.
- Foundation for future integration with `free5gc` control-plane signaling.
