## Context

The `agentic-layer-custom` project is a Go-based 6G AI Core Agent Gateway that orchestrates signaling procedures via three primary agents (`Gateway`, `System`, `Connection`). Separately, the `skill_workshop/backend` project provides a multi-agent system (also ADK-based) for generating and validating the "skills" (Markdown files) used by these agents. This design outlines the migration of the generation logic into the main Gateway project for a unified, single-process platform.

## Goals / Non-Goals

**Goals:**
- Port the specialized Skill Workshop agents (`Intent Analysis`, `Skill Writer`, `Format Checker`) into the Gateway project.
- Integrate the Skill Workshop's WebSocket-based agent runner and REST-based tool catalog endpoints.
- Ensure all services (Signaling Gateway + Skill Generation) run on port 8080 within a single process.
- Synchronize the tool catalog knowledge between the signaling simulation and the skill generation engine.

**Non-Goals:**
- Migration or modification of any frontend (UI) components.
- Refactoring the existing signaling agent (`GatewayAgent`, `SystemAgent`) logic.
- Redesigning the ADK's core `runner` or `session` implementations.

## Decisions

- **Architecture: Integrated Agent Definitions**: Move Skill Workshop agent prompts and instruction providers into `pkg/agents/`. These will be treated as sub-agents in a specialized "Generation Workflow."
- **Routing: Extension of WebsocketSublauncher**: Register the `/ws/agent-run` and `/api/tools` endpoints in `pkg/api/websocket.go` (the current sublauncher for custom APIs).
- **Orchestration: Dedicated Workshop Orchestrator**: Create a `WorkshopOrchestrator` in `pkg/api/` that manages the skill generation lifecycle (analysis -> drafting -> iterative correction).
- **Shared Assets: Unified Tool Catalog**: Modify `pkg/tools/signaling.go` (or create a bridge) so that the dynamic generation agents use the same tool definitions used for signaling execution.

## Risks / Trade-offs

- **[Risk] Resource Contention** → [Mitigation] Monitor memory usage when running multiple LLM-backed agents (signaling + generation) concurrently. Ensure context windows are managed efficiently.
- **[Risk] Naming Collisions in ADK State** → [Mitigation] Use unique, namespaced state keys (e.g., `workshop.markdown_draft`) for the skill generation session to prevent interference with signaling session state.
- **[Risk] Configuration Complexity** → [Mitigation] Centralize configuration (API keys, model selection) in the Gateway's `.env` and `launcher.Config`.
