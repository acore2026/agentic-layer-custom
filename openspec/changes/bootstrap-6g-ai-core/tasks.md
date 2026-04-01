## 1. Project Initialization

- [x] 1.1 Initialize Go module: `go mod init agentic-layer-custom`
- [x] 1.2 Add `google.golang.org/adk` dependency: `go get google.golang.org/adk`
- [x] 1.3 Create the main package structure and entry point in `cmd/agent-gateway/main.go`

## 2. Mock Signaling Tools

- [x] 2.1 Implement `Issue_Access_Token_tool` in `pkg/tools/signaling.go`
- [x] 2.2 Implement `Create_Subnet_PDUSession_tool` in `pkg/tools/signaling.go`
- [x] 2.3 Ensure tools are blocking and return success/failure strings

## 3. Connection Agent Implementation

- [x] 3.1 Initialize the Connection Agent using `adk-go`
- [x] 3.2 Configure the Connection Agent with a ReAct prompt and Gemini provider
- [x] 3.3 Register signaling tools with the Connection Agent
- [x] 3.4 Implement state extraction and management for tools

## 4. System Agent Implementation

- [x] 4.1 Initialize the System Agent as a router using `adk-go`
- [x] 4.2 Configure the System Agent to categorize intents and route to the Connection Agent
- [x] 4.3 Implement clarification request logic for ambiguous intents

## 5. Integration and Testing

- [x] 5.1 Connect the System Agent to the Connection Agent worker function
- [x] 5.2 Create a sample main loop to test intent processing from a CLI
- [x] 5.3 Verify successful routing, tool execution, and clarification scenarios
