package agents

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// NewSystemAgent creates the intent gateway agent.
func NewSystemAgent(m model.LLM) (agent.Agent, error) {
	config := llmagent.Config{
		Name:        "SystemAgent",
		Description: "The intent gateway that categorizes and routes natural language requests.",
		Instruction: `You are the System Agent of a 6G Core Network.
Your task is to categorize user intents and route them to the appropriate worker.

ROUTING TARGETS:
1. 'CONNECTION_AGENT': For any intents related to UE connections, PDU sessions, access tokens, or signaling.

CLARIFICATION RULE:
If the user's intent is highly ambiguous (e.g., "do something", "hi") or you cannot determine the correct target, you MUST return a response asking the user for clarification.

RESPONSE FORMAT:
If you know the target, respond EXACTLY as: 'ROUTING_TO: [TARGET_NAME]'.
Example: 'ROUTING_TO: CONNECTION_AGENT'.
If you need clarification, respond with the question for the user.`,
		Model:           m,
		IncludeContents: llmagent.IncludeContentsDefault,
	}

	return llmagent.New(config)
}

// RouteIntent processes a raw natural language string and routes it to the correct worker or asks for clarification.
func RouteIntent(ctx context.Context, systemAgent agent.Agent, connectionAgent agent.Agent, intent string) (string, error) {
	fmt.Printf("[System Agent] Categorizing intent: %s\n", intent)

	// Run the System Agent to get routing decision
	responseText, err := RunAgent(ctx, systemAgent, intent)
	if err != nil {
		return "", err
	}

	fmt.Printf("[System Agent] LLM Decision: %s\n", responseText)

	if strings.Contains(strings.ToUpper(responseText), "CONNECTION_AGENT") {
		fmt.Println("[System Agent] Routing intent to Connection Agent...")
		return RunAgent(ctx, connectionAgent, intent)
	}

	// If no specific routing target was identified, return the LLM's response (which should be a clarification request).
	fmt.Println("[System Agent] No routing target identified, returning clarification.")
	return responseText, nil
}

// RunAgent is a helper to run an agent and get the final text response.
func RunAgent(ctx context.Context, a agent.Agent, input string) (string, error) {
	ss := session.InMemoryService()
	r, err := runner.New(runner.Config{
		AppName:        "6G-AI-Core",
		Agent:          a,
		SessionService: ss,
	})
	if err != nil {
		return "", err
	}

	sessResp, err := ss.Create(ctx, &session.CreateRequest{
		AppName: "6G-AI-Core",
		UserID:  "poc-user",
	})
	if err != nil {
		return "", err
	}

	// FIX: Added Role: "user" to ensure the message is processed correctly by adk-go and the LLM
	msg := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: input}},
	}
	var responseText string

	for event, err := range r.Run(ctx, "poc-user", sessResp.Session.ID(), msg, agent.RunConfig{}) {
		if err != nil {
			return "", err
		}
		if event.Content != nil {
			for _, p := range event.Content.Parts {
				if p.Text != "" {
					responseText = p.Text
				}
			}
		}
	}
	return responseText, nil
}
