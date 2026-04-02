package main

import (
	"context"
	"fmt"
	"iter"
	"log"
	"os"
	"strings"

	"agentic-layer-custom/pkg/agents"
	"agentic-layer-custom/pkg/api"
	"agentic-layer-custom/pkg/model/kimi"
	"github.com/joho/godotenv"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/console"
	"google.golang.org/adk/cmd/launcher/universal"
	"google.golang.org/adk/cmd/launcher/web"
	webapi "google.golang.org/adk/cmd/launcher/web/api"
	"google.golang.org/adk/cmd/launcher/web/webui"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/session"
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

		// fmt.Printf("[MOCK] systemText contains systemagent: %v, route: %v\n", strings.Contains(strings.ToLower(systemText), "systemagent"), strings.Contains(strings.ToLower(systemText), "route"))

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

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	ctx := context.Background()

	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		provider = "mock"
	}

	var m model.LLM
	var err error

	switch provider {
	case "mock":
		m = &MockLLM{}
		fmt.Println("Using Mock provider for testing")
	case "kimi":
		apiKey := os.Getenv("KIMI_API_KEY")
		modelName := os.Getenv("KIMI_MODEL")
		if modelName == "" {
			modelName = "moonshot-v1-8k"
		}
		if apiKey == "" {
			log.Fatal("KIMI_API_KEY must be set when using kimi provider")
		}
		m = kimi.NewModel(apiKey, modelName)
		fmt.Printf("Using Kimi provider (model: %s)\n", modelName)
	case "gemini":
		modelName := os.Getenv("GEMINI_MODEL")
		if modelName == "" {
			modelName = "gemini-1.5-flash"
		}
		m, err = gemini.NewModel(ctx, modelName, nil)
		if err != nil {
			log.Printf("Warning: Failed to initialize Gemini: %v. Switching to Mock provider.", err)
			m = &MockLLM{}
		} else {
			fmt.Printf("Using Gemini provider (model: %s)\n", modelName)
		}
	default:
		log.Fatalf("Unknown LLM provider: %s", provider)
	}

	connectionAgent, err := agents.NewConnectionAgent(m, "skill")
	if err != nil {
		log.Fatalf("Failed to initialize Connection Agent: %v", err)
	}

	systemAgent, err := agents.NewSystemAgent(m)
	if err != nil {
		log.Fatalf("Failed to initialize System Agent: %v", err)
	}

	gatewayAgent, err := agents.NewGatewayAgent(systemAgent, connectionAgent)
	if err != nil {
		log.Fatalf("Failed to initialize Gateway Agent: %v", err)
	}

	// Use GatewayAgent as the root, others as sub-workers
	loader, err := agent.NewMultiLoader(gatewayAgent, systemAgent, connectionAgent)
	if err != nil {
		log.Fatalf("Failed to create agent loader: %v", err)
	}

	cfg := &launcher.Config{
		AgentLoader:    loader,
		SessionService: session.InMemoryService(),
	}

	l := universal.NewLauncher(
		console.NewLauncher(),
		web.NewLauncher(
			webapi.NewLauncher(),
			webui.NewLauncher(),
			api.NewLauncher(gatewayAgent),
		),
	)

	fmt.Println("Launching ADK Web UI on http://localhost:8080/ui/ ...")
	// The universal launcher needs "web" to activate the web server, 
	// then "api" for REST, "webui" for dashboard, and "ws" for our custom stream.
	if err := l.Execute(ctx, cfg, []string{"web", "api", "webui", "ws"}); err != nil {
		log.Fatalf("Launcher execution failed: %v", err)
	}
}
