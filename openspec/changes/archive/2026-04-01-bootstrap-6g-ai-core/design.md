## Context

Building a proof-of-concept for a 6G core network that uses an AI layer to manage control-plane intents. This requires a reliable and extensible framework for agent orchestration and tool execution.

## Goals / Non-Goals

**Goals:**
- Implement a **System Agent** to route natural language intents to worker agents.
- Implement a **Connection Agent** using a ReAct loop to process intents via specialized tools.
- Create mock Go API tools for issuing tokens and creating PDU sessions.
- Use the `google/adk-go` framework for agent and tool management.

**Non-Goals:**
- Handling LLM context window exhaustion.
- Implementing complex rollback or error recovery logic for signaling.
- Real-time performance optimizations (seconds of latency are acceptable).
- Full integration with a real 5G/6G core (using mock signaling for now).

## Decisions

- **Framework Choice**: `google/adk-go`. Rationale: It provides a structured way to define agents, tools, and prompts for LLM-driven applications in Go.
- **LLM Provider**: Gemini. Rationale: High-quality reasoning and native integration with many AI tools.
- **Connection Agent Pattern**: ReAct (Reason + Act). Rationale: It allows the agent to reason about the user's intent and execute tools sequentially, which is ideal for multi-step network signaling.
- **Tool Design**: Strictly blocking. Rationale: Simplifies state management and ensures that the LLM receives the outcome of signaling before moving to the next step.

## Risks / Trade-offs

- **[Risk] LLM Latency** → **Mitigation**: Latency is acceptable for this PoC. We will focus on functional correctness.
- **[Risk] State Management Error** → **Mitigation**: The LLM will be explicitly instructed to extract and pass state between tool calls in the system prompt.
- **[Risk] Incorrect Intent Routing** → **Mitigation**: The System Agent will return a clarification request if it's unsure, preventing incorrect worker agent invocation.
