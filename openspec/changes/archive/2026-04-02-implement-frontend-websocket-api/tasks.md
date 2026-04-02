## 1. Setup and Dependencies

- [x] 1.1 Add `github.com/gorilla/websocket` dependency: `go get github.com/gorilla/websocket`
- [x] 1.2 Define shared event structures in `pkg/agents/telemetry.go` (ai_payload, llm_thought, network_pcap, workflow_complete).

## 2. Telemetry Hub Implementation

- [x] 2.1 Implement a `TelemetryHub` in `pkg/agents/telemetry.go` to handle thread-safe event broadcasting via Go channels.
- [x] 2.2 Provide a global or context-based way for agents and tools to access the hub.

## 3. Agent and Tool Integration

- [x] 3.1 Update `GatewayAgent` to emit `ai_payload` events for requests and `workflow_complete` upon finish.
- [x] 3.2 Update `ConnectionAgent` loop to emit `llm_thought` events when `genai.Part.Thought` is detected.
- [x] 3.3 Update `UniversalMockTool` in `pkg/tools/signaling.go` to emit `network_pcap` events for requests and responses.

## 4. WebSocket API Implementation

- [x] 4.1 Create `pkg/api/websocket.go` to handle the `/v1/intents/stream` endpoint.
- [x] 4.2 Implement the `execute_intent` handler that triggers the `GatewayAgent` and pipes telemetry to the WS connection.
- [x] 4.3 Update `cmd/agent-gateway/main.go` to start the HTTP/WS server alongside the existing ADK Web UI.

## 5. Verification

- [x] 5.1 Use a WebSocket test tool (or the React frontend) to connect to `ws://localhost:8080/v1/intents/stream`.
- [x] 5.2 Send an `execute_intent` and verify the sequence of JSON events matches the API Design document.
- [x] 5.3 Confirm that tool calls generate valid PCAP-formatted events.
