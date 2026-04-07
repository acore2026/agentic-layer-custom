package main

import (
	"context"
	"iter"
	"strings"

	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

// MockLLM provides a robust stateless mock for testing the gateway routing.
type MockLLM struct{}

func (m *MockLLM) Name() string { return "mock-llm" }
func (m *MockLLM) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		var systemText string
		var prompt string
		var hasFunctionResponse bool
		if req.Config != nil && req.Config.SystemInstruction != nil {
			for _, p := range req.Config.SystemInstruction.Parts {
				systemText += p.Text
			}
		}
		for _, c := range req.Contents {
			for _, p := range c.Parts {
				prompt += p.Text
				if p.FunctionResponse != nil {
					hasFunctionResponse = true
				}
			}
		}
		fullPrompt := strings.ToLower(systemText + prompt)

		var parts []*genai.Part
		// If systemText contains SystemAgent instructions, it's the System Agent turn.
		if strings.Contains(strings.ToLower(systemText), "systemagent") {
			if strings.Contains(fullPrompt, "pdu") || strings.Contains(fullPrompt, "create") || strings.Contains(fullPrompt, "token") || strings.Contains(fullPrompt, "registration") || strings.Contains(fullPrompt, "connect") {
				parts = append(parts, &genai.Part{Text: "ROUTING_TO: CONNECTION_AGENT"})
			} else if strings.Contains(fullPrompt, "hi") || strings.Contains(fullPrompt, "hello") {
				parts = append(parts, &genai.Part{Text: "I'm the 6G Core System Agent. How can I help you?"})
			} else {
				parts = append(parts, &genai.Part{Text: "I am ready to route your intent."})
			}
		} else {
			// This is the worker agent's turn (ConnectionAgent)
			if !hasFunctionResponse {
				parts = append(parts, &genai.Part{
					Thought: true,
					Text:    "To process this connection request, I first need to check the UE subscription status in UDM.",
				})
				parts = append(parts, &genai.Part{
					FunctionCall: &genai.FunctionCall{
						Name: "Subscription_tool",
						Args: map[string]any{"ue_id": "SUCI_12345"},
					},
				})
			} else {
				parts = append(parts, &genai.Part{
					Thought: true,
					Text:    "Subscription is valid. Completing the connection procedure.",
				})
				parts = append(parts, &genai.Part{
					Text: "I have successfully processed your connection intent and performed the signaling. UE SUCI_12345 is now connected.",
				})
			}
		}

		resp := &model.LLMResponse{
			Content: &genai.Content{
				Role:  "model",
				Parts: parts,
			},
		}
		yield(resp, nil)
	}
}
