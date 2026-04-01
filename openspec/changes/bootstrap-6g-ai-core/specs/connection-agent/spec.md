## ADDED Requirements

### Requirement: ReAct Loop Execution
The Connection Agent SHALL implement a ReAct (Reason + Act) loop using `google/adk-go` and the Gemini LLM provider.

#### Scenario: Successful intent execution
- **WHEN** the Connection Agent receives a raw intent like "I need an access token and then a PDU session for my subnet"
- **THEN** the agent reasons, calls the `Issue_Access_Token_tool`, receives the result, and then calls `Create_Subnet_PDUSession_tool` with the extracted state

### Requirement: Blocking Tool Invocation
Registered Go API tools MUST be strictly blocking and return a success or failure string back to the LLM.

#### Scenario: Tool call returns result to LLM
- **WHEN** the `Issue_Access_Token_tool` is called
- **THEN** it waits for the mock signaling to complete and returns the token string as the tool's output to the LLM

### Requirement: LLM State Extraction and Management
The Connection Agent SHALL delegate state extraction (e.g., UE IDs or Access Tokens) from tool outputs to the LLM for use in subsequent tool calls.

#### Scenario: Token passed from one tool to the next
- **WHEN** `Issue_Access_Token_tool` returns a token string
- **THEN** the LLM extracts that token and passes it as an argument to `Create_Subnet_PDUSession_tool`
