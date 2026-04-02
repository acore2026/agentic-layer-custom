package agents

import (
	"fmt"
	"iter"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// NewGatewayAgent creates a unified orchestrator for the Web UI.
func NewGatewayAgent(system, connection agent.Agent) (agent.Agent, error) {
	return agent.New(agent.Config{
		Name:        "6G-AI-Gateway",
		Description: "Unified gateway for 6G Core signaling. Routes intents and executes connection procedures.",
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				// 1. Run System Agent to get routing decision
				fmt.Printf("[Gateway] Running System Agent for routing...\n")
				var routingDecision string
				for event, err := range system.Run(ctx) {
					if err != nil {
						yield(nil, err)
						return
					}
					if event.Content != nil {
						for _, p := range event.Content.Parts {
							if p.Text != "" {
								routingDecision = p.Text
							}
						}
					}
					// Forward system agent events (like reasoning) to the UI
					if !yield(event, nil) {
						return
					}
				}

				// 2. Decide if we should delegate to Connection Agent
				if strings.Contains(strings.ToUpper(routingDecision), "CONNECTION_AGENT") {
					fmt.Printf("[Gateway] Routing to Connection Agent...\n")
					// Inject a "routing" event so the user sees the transition
					yield(&session.Event{
						LLMResponse: model.LLMResponse{
							Content: &genai.Content{
								Role:  "model",
								Parts: []*genai.Part{{Text: "\n--- Routing to Connection Agent ---\n"}},
							},
						},
					}, nil)

					// Run the Connection Agent with the SAME context
					for event, err := range connection.Run(ctx) {
						if err != nil {
							yield(nil, err)
							return
						}
						if !yield(event, nil) {
							return
						}
					}
				}
			}
		},
	})
}
