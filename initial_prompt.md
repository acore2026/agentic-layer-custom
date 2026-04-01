You are an expert Go backend developer and telecom network architect specializing in 5G/6G core networks (like `free5gc`) and AI agent integrations. 

I am building a functional proof-of-concept for a 6G core network. It features a decoupled AI layer built on top of a 5G core. The goal is to enable intent-driven networking using a "Skill match and execution" architecture.

Please help me write the initial Go code for this architecture using Go 1.22+ and the `google/adk-go` framework. We are using Gemini as our LLM provider.

# Architecture Context
There are 4 types of agents in this network. We will focus on two for this task:
1. **System Agent:** The intent gateway/entry point. 
2. **Connection Agent:** The worker agent for connection-related intents.

## 1. System Agent Specifications
- Acts as an LLM router. It receives raw natural language text from the Device Domain (UE -> RAN -> System Agent) or App Domain (APP -> NEF/GW -> System Agent).
- **Goal:** Read the text, categorize it, and route it to the appropriate Worker Agent (e.g., Connection Agent).
- **Handoff:** Passes the raw natural language text directly to the worker.
- **Error Handling:** If the intent is highly ambiguous or the LLM fails to output a valid routing target, the System Agent must NOT fallback programmatically. Instead, it must return a message asking the UE/App for clarification.

## 2. Connection Agent Specifications
- Acts as a worker utilizing the ReAct (Reason + Act) prompting approach via the LLM.
- **Skills:** The system prompt will be injected with specific "Skills" (YAML/Markdown format) that outline tool inventories and critical rules. 
- **Execution:** It receives the raw intent, thinks, and calls independent Go API tools. 
- **State Management:** The LLM is responsible for extracting state variables (like UE IDs or Access Tokens) from the output of one tool and passing them as arguments into the next tool.
- **Tool Behavior:** The underlying Go API tools (refactored from free5gc control-plane signaling) will be strictly blocking. The tool will wait for the signaling to complete and return a success/fail string back to the LLM.

## Constraints & Allowances
- **Latency:** Ignore standard 3GPP timeouts for control plane procedures. This is a PoC; LLM inference latency (seconds) is acceptable.
- **Rollbacks & Context Window:** Ignore rollback logic and LLM context window exhaustion for now.

# Task
Please generate the foundational Go code (using `google/adk-go`) to bootstrap this system. Specifically, I need:

1. **The System Agent Logic:** A function that takes a natural language string, uses the LLM to categorize it, and either routes it to a mock Connection Agent function or returns a clarification request.
2. **The Connection Agent ReAct Setup:** The scaffolding for the Connection Agent using `adk-go` that initializes a ReAct loop.
3. **Mock Tool Setup:** Define two simple, blocking dummy tools in Go (e.g., `Issue_Access_Token_tool` and `Create_Subnet_PDUSession_tool`) and show how they are registered to the Connection Agent so the LLM can call them and pass state between them.

Please provide clean, well-commented Go code that I can use as the backbone for this prototype.