package agents

import (
	"agentic-layer-custom/pkg/tools"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// NewConnectionAgent creates a worker agent for connection-related intents.
func NewConnectionAgent(m model.LLM) (agent.Agent, error) {
	// 1. Define Tools
	issueTokenTool, err := functiontool.New(functiontool.Config{
		Name:        "Issue_Access_Token_tool",
		Description: "Issues a mock access token for a given UE ID.",
	}, tools.IssueAccessTokenTool)
	if err != nil {
		return nil, err
	}

	createPduTool, err := functiontool.New(functiontool.Config{
		Name:        "Create_Subnet_PDUSession_tool",
		Description: "Creates a mock PDU session on a specific subnet for a UE ID using an access token.",
	}, tools.CreateSubnetPDUSessionTool)
	if err != nil {
		return nil, err
	}

	// 2. Configure Agent
	config := llmagent.Config{
		Name:        "ConnectionAgent",
		Description: "A worker agent that handles connection-related intents like issuing tokens and creating PDU sessions.",
		Instruction: `You are the Connection Agent of a 6G AI Core Network.
Your task is to process connection-related intents using the provided signaling tools.
You MUST use a ReAct (Reason + Act) approach:
1. Reason: Analyze the user's intent and decide which signaling tool to call first.
2. Act: Call the chosen tool with the correct arguments (e.g., ue_id).
3. Observe: Review the tool's response. If it contains a 'token', you MUST use it for subsequent calls.
4. Continue: If more steps are needed (e.g., creating a PDU session after getting a token), repeat the reasoning and acting process.

CRITICAL STATE MANAGEMENT:
- When you call 'Issue_Access_Token_tool', it returns a 'token'.
- You MUST pass this 'token' as the 'access_token' argument to the 'Create_Subnet_PDUSession_tool'.
- If a 'ue_id' is provided in the intent, use it for all tool calls.

Once all steps are completed, summarize the actions taken for the user.`,
		Model:           m,
		Tools:           []tool.Tool{issueTokenTool, createPduTool},
		IncludeContents: llmagent.IncludeContentsDefault,
	}

	return llmagent.New(config)
}
