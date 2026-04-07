## Why

Consolidate the "Skill Workshop" capability—which enables dynamic generation and validation of 6G signaling procedures—into the primary 6G AI Core Agent Gateway. This migration eliminates the overhead of managing two separate processes/ports and provides a unified agentic platform for both network orchestration and procedure authoring.

## What Changes

- **Core Logic Migration**: Port the ADK-based orchestrator and agent definitions (Intent Analysis, Skill Writer, Format Checker) from the standalone backend to the Gateway project.
- **Unified API Surface**: Integrate the Skill Workshop WebSocket (`/ws/agent-run`) and REST (`/api/tools`) endpoints into the existing Gateway server.
- **Single Process Execution**: Configure the system to run both the signaling gateway and the skill generation engine within a single process on port 8080.
- **Tool Catalog Shared Knowledge**: Ensure the dynamic generation agents have access to the same tool catalog used by the signaling tools for consistent validation.

## Capabilities

### New Capabilities
- `skill-generation`: Capability to generate, validate, and iteratively repair 6G signaling procedures (Markdown skills) using a multi-agent workflow.

### Modified Capabilities
- `websocket-api`: Extend the existing WebSocket infrastructure to support real-time telemetry for the skill generation process.
- `skills-api`: Update the skills management API to potentially trigger or interface with the generation engine.

## Impact

- **Infrastructure**: `cmd/agent-gateway/main.go` will be modified to register new routes and initialize the migration-related agents.
- **API**: New endpoints at `/ws/agent-run` and `/api/tools`.
- **Agents**: Addition of three specialized agents in `pkg/agents/` (Intent Analysis, Skill Writer, Format Checker).
- **Dependencies**: Integration of any unique dependencies from the `skill_workshop` (e.g., additional ADK runner logic).
