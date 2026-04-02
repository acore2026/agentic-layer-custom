## Why

The current Connection Agent is limited to a hardcoded set of tools and a single, manually defined workflow. To support the flexible, multi-step procedures defined in `SKILL.md` files (like initial registration, PDU session establishment, and ACN), we need a generic orchestration layer. This allows the agent to dynamically load these skills and execute their defined tool sequences using LLM-driven reasoning, eliminating the need for manual state management or hardcoded logic for every new procedure.

## What Changes

- **Dynamic Skill Loading**: The Connection Agent will now automatically read all `SKILL.md` files from the `skill/` directory and inject their workflows into its system instructions.
- **Universal Tool Mocking**: A generic tool implementation will be created that can be registered under any name. It logs calls for verification and returns flexible, SUCCESS-oriented responses that the LLM can use for state extraction.
- **Automated Tool Discovery**: The system will parse `SKILL.md` files at startup to identify all required tools and register them dynamically with the agent.
- **LLM-Driven State Management**: Hardcoded state passing (like manually extracting a token) will be replaced with explicit instructions for the LLM to manage state between sequential tool calls.

## Capabilities

### New Capabilities
- `skill-orchestration`: Dynamic loading and execution of multi-step skills from Markdown definitions.
- `universal-mocking`: Generic tool execution and logging for verification of agent sequences.

### Modified Capabilities
- `connection-agent`: Transition from a hardcoded tool set to a dynamic, skill-based tool set.

## Impact

- `pkg/agents/connection.go`: Agent initialization and instruction building logic.
- `pkg/tools/signaling.go`: Addition of the universal mock tool and discovery utility.
- `cmd/agent-gateway/main.go`: Updated initialization flow for the Connection Agent.
