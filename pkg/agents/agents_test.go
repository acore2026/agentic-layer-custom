package agents_test

import (
	"context"
	"fmt"
	"iter"
	"testing"

	"agentic-layer-custom/pkg/agents"
	"agentic-layer-custom/pkg/tools"

	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

// MockLLM is a simple implementation of model.LLM for testing without real API calls.
type MockLLM struct {
	Response string
}

func (m *MockLLM) Name() string { return "mock-llm" }

func (m *MockLLM) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		resp := &model.LLMResponse{
			Content: &genai.Content{
				Parts: []*genai.Part{{Text: m.Response}},
			},
		}
		yield(resp, nil)
	}
}

func TestSystemAgentRouting(t *testing.T) {
	ctx := context.Background()
	
	// Mock System Agent that always routes to CONNECTION_AGENT
	systemModel := &MockLLM{Response: "ROUTING_TO: CONNECTION_AGENT"}
	systemAgent, err := agents.NewSystemAgent(systemModel)
	if err != nil {
		t.Fatalf("Failed to create system agent: %v", err)
	}

	// Mock Connection Agent that returns a simple success
	connectionModel := &MockLLM{Response: "Connection established successfully."}
	connectionAgent, err := agents.NewConnectionAgent(connectionModel)
	if err != nil {
		t.Fatalf("Failed to create connection agent: %v", err)
	}

	t.Run("Should route to Connection Agent", func(t *testing.T) {
		response, err := agents.RouteIntent(ctx, systemAgent, connectionAgent, "create a pdu session")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		expected := "Connection established successfully."
		if response != expected {
			t.Errorf("Expected response %q, got %q", expected, response)
		}
	})

	t.Run("Should request clarification on unknown intent", func(t *testing.T) {
		clarificationModel := &MockLLM{Response: "Can you please clarify what you want to do?"}
		clarificationAgent, err := agents.NewSystemAgent(clarificationModel)
		if err != nil {
			t.Fatalf("Failed to create clarification agent: %v", err)
		}
		
		response, err := agents.RouteIntent(ctx, clarificationAgent, connectionAgent, "unknown")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		expected := "Can you please clarify what you want to do?"
		if response != expected {
			t.Errorf("Expected response %q, got %q", expected, response)
		}
	})
}

func TestSignalingTools(t *testing.T) {
	t.Run("IssueAccessTokenTool should return a token", func(t *testing.T) {
		args := &tools.IssueAccessTokenArgs{UEID: "UE-1"}
		// Pass nil for tool.Context as it's not used in our mock tool
		res, err := tools.IssueAccessTokenTool(nil, args)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if res.Token == "" {
			t.Error("Expected a token, got empty string")
		}
		fmt.Printf("Mock token: %s\n", res.Token)
	})
}
