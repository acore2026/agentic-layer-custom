# 6G AI Core - Frontend/Backend Interface Design

## Overview

Because LLM inference and multi-agent orchestration take time, the interface between the React frontend and the Go backend MUST be asynchronous and streaming.

We will use a **WebSocket (WS)** connection (or Server-Sent Events). The frontend sends a single user intent, and the backend streams back a series of structured JSON events as the ADK agents "think" and execute tools.

## Connection Endpoint

`ws://localhost:8080/v1/intents/stream`

## 1. Client-to-Server (The Request)

When the user clicks "Execute Intent", the React frontend sends this JSON payload over the WebSocket:

```json
{
  "type": "execute_intent",
  "data": {
    "intent": "Connect my new embodied agent to a high-reliability subnet.",
    "scenarioId": "ACN" 
  }
}
```

## 2. Server-to-Client (The Event Stream)

The Go backend will stream discrete events back to the frontend. Every message must have a `type` so the React app knows which UI panel to update.

### Event Type: `ai_payload`

- **Target UI:** Left Panel (LLM API Payloads)
- **Trigger:** Whenever the System Agent or Connection Agent sends a prompt to Kimi/Gemini, or receives a structured response.

```json
{
  "type": "ai_payload",
  "data": {
    "agent": "SystemAgent",
    "role": "user",
    "content": "Categorize intent: 'Connect my new embodied agent...'\\nAvailable Agents: Connection, Compute, Sense."
  }
}
```

### Event Type: `llm_thought`

- **Target UI:** Left Panel (Can be rendered inside the assistant payload as dimmed text)
- **Trigger:** When `pkg/model/kimi/kimi.go` yields a `genai.Part.Thought` chunk.

```json
{
  "type": "llm_thought",
  "data": {
    "agent": "ConnectionAgent",
    "chunk": "First, I need to check the subscription status using Subscription_tool..."
  }
}
```

### Event Type: `network_pcap`

- **Target UI:** Right Panel (Simulated Network Traffic PCAP & Details)
- **Trigger:** Emitted by the `UniversalMockTool` in `pkg/tools/` when a tool is called (Request) and when it finishes (Response).

**Example: Outgoing Tool Request**
```json
{
  "type": "network_pcap",
  "data": {
    "direction": "request",
    "source": "ConnectionAgent",
    "destination": "UDM",
    "protocol": "HTTP/2",
    "info": "GET /nudm-sdm/v2/sm-data",
    "details": {
      "method": "GET",
      "path": "/nudm-sdm/v2/sm-data",
      "toolName": "Subscription_tool",
      "payload": { "ueId": "SUCI_12345" }
    }
  }
}
```

**Example: Incoming Tool Response**
```json
{
  "type": "network_pcap",
  "data": {
    "direction": "response",
    "source": "UDM",
    "destination": "ConnectionAgent",
    "protocol": "HTTP/2",
    "info": "200 OK (Application/JSON)",
    "details": {
      "status": 200,
      "payload": { "is_subscribed": true, "slice": "eMBB" }
    }
  }
}
```

### Event Type: `workflow_complete`

- **Target UI:** Global (Stops the progress bar and thinking animations)
- **Trigger:** When the ADK agent loop finishes and outputs "DONE".

```json
{
  "type": "workflow_complete",
  "data": {
    "status": "success",
    "finalMessage": "Embodied agent successfully connected to the subnet."
  }
}
```

## 3. Frontend Integration Plan

To integrate this into `App.jsx`, we will replace the `runSimulation()` mock loop with a real WebSocket client:

1. Create a new `WebSocket('ws://localhost:8080/v1/intents/stream')`.
2. Send the `execute_intent` JSON.
3. Listen on `ws.onmessage`.
4. Use a `switch(message.type)` statement.
   - If `ai_payload`: Append to `aiPayloads` state.
   - If `network_pcap`: Call `addPacket()` to update the PCAP table and detail view.
   - If `workflow_complete`: Set `isProcessing(false)`.