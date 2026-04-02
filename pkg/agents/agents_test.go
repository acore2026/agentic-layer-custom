package agents_test

import (
	"context"
	"fmt"
	"iter"
	"strings"
	"testing"

	"agentic-layer-custom/pkg/agents"
	"agentic-layer-custom/pkg/tools"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
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

// runAgent is a test helper to run an agent and get the final text response.
func runAgent(ctx context.Context, a agent.Agent, input string) (string, error) {
	ss := session.InMemoryService()
	r, err := runner.New(runner.Config{
		AppName:        "test-app",
		Agent:          a,
		SessionService: ss,
	})
	if err != nil {
		return "", err
	}

	sessResp, err := ss.Create(ctx, &session.CreateRequest{
		AppName: "test-app",
		UserID:  "test-user",
	})
	if err != nil {
		return "", err
	}

	msg := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: input}},
	}
	var responseText string

	for event, err := range r.Run(ctx, "test-user", sessResp.Session.ID(), msg, agent.RunConfig{}) {
		if err != nil {
			return "", err
		}
		if event.Content != nil {
			for _, p := range event.Content.Parts {
				if p.Text != "" && !strings.Contains(p.Text, "Routing to Connection Agent") {
					responseText = p.Text
				}
			}
		}
	}
	return responseText, nil
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
	connectionAgent, err := agents.NewConnectionAgent(connectionModel, "../../skill")
	if err != nil {
		t.Fatalf("Failed to create connection agent: %v", err)
	}

	gatewayAgent, err := agents.NewGatewayAgent(systemAgent, connectionAgent)
	if err != nil {
		t.Fatalf("Failed to create gateway agent: %v", err)
	}

	t.Run("Should route to Connection Agent", func(t *testing.T) {
		response, err := runAgent(ctx, gatewayAgent, "create a pdu session")
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
		
		gatewayAgent, _ := agents.NewGatewayAgent(clarificationAgent, connectionAgent)
		response, err := runAgent(ctx, gatewayAgent, "unknown")
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
		fmt.Printf("Mock token: %s\\n", res.Token)
	})
}
