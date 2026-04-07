## 1. Core Logic Migration

- [x] 1.1 Port `server/adk_agents.go` agent definitions (prompts, instruction providers) to `pkg/agents/workshop_agents.go`.
- [x] 1.2 Port `server/adk_runtime.go` orchestrator logic to `pkg/api/workshop_orchestrator.go`.
- [x] 1.3 Port `server/validators.go` markdown validation logic to `pkg/api/workshop_validators.go`.
- [x] 1.4 Port `server/openai_compatible.go` LLM provider wrapper to `pkg/model/openai_compatible.go`.
- [x] 1.5 Port `server/adk_state.go` and `server/adk_events.go` to `pkg/api/` as needed for workshop-specific types.
- [x] 1.6 Update `pkg/agents/workshop_agents.go` to use the project's internal `model.LLM` interface if compatible, otherwise use the ported OpenAI wrapper.

## 2. API Integration

- [x] 2.1 Implement the `HandleAgentRun` WebSocket handler in `pkg/api/workshop_ws.go`.
- [x] 2.2 Register the `/ws/agent-run` route in `pkg/api/websocket.go` (the `WebsocketSublauncher`).
- [x] 2.3 Implement and register the `/api/tools` GET handler in `pkg/api/websocket.go`.
- [x] 2.4 Update `cmd/agent-gateway/main.go` to initialize the `WorkshopOrchestrator` and integrate it into the server launcher.

## 3. Tool Catalog Synchronization

- [x] 3.1 Create `pkg/tools/catalog.go` to extract and normalize tool metadata from `pkg/tools/signaling.go`.
- [x] 3.2 Ensure the `Format Checker` agent and the `/api/tools` endpoint share this centralized catalog.

## 4. Verification & Testing

- [x] 4.1 Verify the `/api/tools` endpoint using `curl` to ensure valid JSON output.
- [x] 4.2 Create a `test_ws_workshop.go` script to verify end-to-end skill generation via the `/ws/agent-run` WebSocket.
- [x] 4.3 Confirm that signaling procedures and skill generation can run concurrently without interference on port 8080.
