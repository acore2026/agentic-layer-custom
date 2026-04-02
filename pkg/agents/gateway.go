package agents

import (
	"agentic-layer-custom/pkg/telemetry"
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
				// Emit initial user intent
				var intent string
				if ctx.UserContent() != nil && len(ctx.UserContent().Parts) > 0 {
					intent = ctx.UserContent().Parts[0].Text
				}
				telemetry.GetHub().Emit(telemetry.TelemetryEvent{
					Type:telemetry.EventTypeAIPayload,
					Data: telemetry.AIPayloadData{
						Agent:   "SystemAgent",
						Role:    "user",
						Content: intent,
					},
				})

				// 1. Run System Agent to get routing decision
				fmt.Printf("[Gateway] Running System Agent for routing...\n")
				var routingDecision string
				var lastResponse string
				for event, err := range system.Run(ctx) {
					if err != nil {
						yield(nil, err)
						return
					}
					if event.Content != nil {
						for _, p := range event.Content.Parts {
							if p.Thought {
								telemetry.GetHub().Emit(telemetry.TelemetryEvent{
									Type:telemetry.EventTypeLLMThought,
									Data: telemetry.LLMThoughtData{Agent: "SystemAgent", Chunk: p.Text},
								})
							}
							if p.Text != "" {
								routingDecision = p.Text
								lastResponse = p.Text
								if event.Content.Role == "model" {
									telemetry.GetHub().Emit(telemetry.TelemetryEvent{
										Type:telemetry.EventTypeAIPayload,
										Data: telemetry.AIPayloadData{
											Agent:   "SystemAgent",
											Role:    "assistant",
											Content: p.Text,
										},
									})
								}
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

					// Emit ConnectionAgent request
					telemetry.GetHub().Emit(telemetry.TelemetryEvent{
						Type:telemetry.EventTypeAIPayload,
						Data: telemetry.AIPayloadData{
							Agent:   "ConnectionAgent",
							Role:    "user",
							Content: intent,
						},
					})

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
						if event.Content != nil {
							for _, p := range event.Content.Parts {
								if p.Thought {
									telemetry.GetHub().Emit(telemetry.TelemetryEvent{
										Type:telemetry.EventTypeLLMThought,
										Data: telemetry.LLMThoughtData{Agent: "ConnectionAgent", Chunk: p.Text},
									})
								}
								if p.Text != "" {
									lastResponse = p.Text
									if event.Content.Role == "model" {
										telemetry.GetHub().Emit(telemetry.TelemetryEvent{
											Type:telemetry.EventTypeAIPayload,
											Data: telemetry.AIPayloadData{
												Agent:   "ConnectionAgent",
												Role:    "assistant",
												Content: p.Text,
											},
										})
									}
								}
							}
						}
						if !yield(event, nil) {
							return
						}
					}
				}

				// Emit workflow complete
				telemetry.GetHub().Emit(telemetry.TelemetryEvent{
					Type:telemetry.EventTypeWorkflowComplete,
					Data: telemetry.WorkflowCompleteData{
						Status:       "success",
						FinalMessage: lastResponse,
					},
				})
			}
		},
	})
}
