## Context

The 6G AI Core PoC requires a real-time visual interface. The React frontend expects a WebSocket stream of structured JSON events to update its internal state and UI panels (Chain of Thought, PCAP, etc.). The Go backend currently uses a CLI-focused runner. We need to bridge the ADK agent lifecycle with a persistent WebSocket connection.

## Goals / Non-Goals

**Goals:**
- Implement a WebSocket server using `github.com/gorilla/websocket`.
- Define standard JSON structures for `ai_payload`, `llm_thought`, `network_pcap`, and `workflow_complete`.
- Integrate telemetry emission into `GatewayAgent`, `SystemAgent`, `ConnectionAgent`, and `UniversalMockTool`.
- Support concurrent execution of intents over WebSocket.

**Non-Goals:**
- Implementing real network signaling (continuing with mock signaling).
- Persistent storage of conversation history (in-memory only for the stream).
- Authentication/Authorization for the WebSocket (out of scope for PoC).

## Decisions

- **WebSocket Framework**: Use `github.com/gorilla/websocket` for its robustness and wide adoption in the Go ecosystem.
- **Event Bus Pattern**: Implement a simple internal event channel or "Telemetry Hub" that agents can write to without knowing about the WebSocket connection directly. This keeps the agent logic decoupled from the transport layer.
- **Global Intent Handler**: The WebSocket handler will instantiate the `GatewayAgent` and run it for each incoming `execute_intent` message.
- **PCAP Mapping**: `UniversalMockTool` will be updated to take a context that allows emitting events to the hub. It will map tool names to simulated network entities (e.g., `Subscription_tool` -> `UDM`).

## Risks / Trade-offs

- **[Risk] WebSocket Connection Stability** → Network issues could drop the stream. 
- **[Mitigation]** → Frontend should implement a simple reconnect logic. PoC will focus on happy-path stability.
- **[Risk] Concurrency Race Conditions** → Multiple events emitted simultaneously might be interleaved.
- **[Mitigation]** → Use a thread-safe channel for the Telemetry Hub to serialize outgoing messages to the WebSocket client.
- **[Trade-off] Performance overhead** → JSON marshaling and emission for every part of the LLM stream.
- **[Rationale]** → Necessary for the real-time "thinking" visualization required by the frontend.
