package agents

import (
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
)

// NewSystemAgent creates the intent gateway agent.
func NewSystemAgent(m model.LLM) (agent.Agent, error) {
	config := llmagent.Config{
		Name:        "SystemAgent",
		Description: "The intent gateway that categorizes and routes natural language requests.",
		Instruction: `You are the System Agent of a 6G Core Network.
Your ONLY task is to categorize user intents and route them to the appropriate worker.

ROUTING TARGETS:
1. 'CONNECTION_AGENT': For any intents related to UE connections, PDU sessions, access tokens, registration, or signaling.

STRICT RULE:
DO NOT provide any technical explanations, descriptions, or help yourself.
Your ONLY valid responses are routing decisions or clarification questions.

RESPONSE FORMAT:
If you identified the target, respond ONLY with: 'ROUTING_TO: [TARGET_NAME]'.
Example: 'ROUTING_TO: CONNECTION_AGENT'.
If you are absolutely unsure and need more info, respond ONLY with a short question for the user.`,
		Model:           m,
		IncludeContents: llmagent.IncludeContentsDefault,
	}

	return llmagent.New(config)
}
