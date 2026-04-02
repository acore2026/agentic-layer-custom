## Why

The current backend implementation lacks a structured, real-time streaming interface for the frontend to consume. To support the React frontend's requirement for live visualization of agent "thinking", tool calls (PCAP), and multi-agent orchestration, we need a WebSocket-based API that emits discrete, typed events. This will replace the simple request-response model with a rich, asynchronous stream.

## What Changes

- **WebSocket Endpoint**: Introduce a new `/v1/intents/stream` WebSocket endpoint.
- **Typed Event Streaming**: The backend will emit structured JSON events of types: `ai_payload`, `llm_thought`, `network_pcap`, and `workflow_complete`.
- **Agent Telemetry Integration**: Hook into the ADK agent lifecycle to emit `ai_payload` and `llm_thought` events.
- **Tool PCAP Emission**: Update `UniversalMockTool` to emit `network_pcap` events for tool requests and responses.
- **Workflow State Management**: Emit `workflow_complete` when the agent orchestration finishes.

## Capabilities

### New Capabilities
- `websocket-api`: Asynchronous streaming interface for frontend collaboration.
- `telemetry-streamer`: System-wide event emitter for agent and tool activities.

### Modified Capabilities
- `connection-agent`: Enhanced to emit granular execution events for the web stream.
- `universal-mocking`: Updated to provide PCAP-formatted telemetry for simulated signaling.

## Impact

- `cmd/agent-gateway/main.go`: Addition of WebSocket server and route handling.
- `pkg/agents/`: Integration of telemetry emitters in the agent runner/loop.
- `pkg/tools/`: Extension of `UniversalMockTool` to support structured event emission.
- Dependencies: Addition of a WebSocket library (e.g., `github.com/gorilla/websocket`).
