package main

import (
	"bufio"
	"context"
	"fmt"
	"iter"
	"log"
	"os"
	"strings"

	"agentic-layer-custom/pkg/agents"
	"agentic-layer-custom/pkg/model/kimi"
	"github.com/joho/godotenv"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

// MockLLM provides a robust stateless mock for testing the gateway routing.
type MockLLM struct{}

func (m *MockLLM) Name() string { return "mock-llm" }
func (m *MockLLM) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		var prompt string
		if req.Config != nil && req.Config.SystemInstruction != nil {
			for _, p := range req.Config.SystemInstruction.Parts {
				prompt += p.Text
			}
		}
		for _, c := range req.Contents {
			for _, p := range c.Parts {
				prompt += p.Text
			}
		}
		prompt = strings.ToLower(prompt)

		var response string
		// Simplify: If user says PDU or create, System Agent should say CONNECTION_AGENT.
		// If prompt contains SystemAgent instructions, it's the System Agent turn.
		
		if strings.Contains(prompt, "systemagent") || strings.Contains(prompt, "route") {
			if strings.Contains(prompt, "pdu") || strings.Contains(prompt, "create") || strings.Contains(prompt, "token") || strings.Contains(prompt, "registration") || strings.Contains(prompt, "connect") {
				response = "ROUTING_TO: CONNECTION_AGENT"
			} else if strings.Contains(prompt, "hi") || strings.Contains(prompt, "hello") {
				response = "I'm the 6G Core System Agent. How can I help you?"
			} else {
				response = "I am ready to route your intent."
			}
		} else {
			// This is the worker agent's turn
			response = "I have successfully processed your connection intent and performed the signaling."
		}

		resp := &model.LLMResponse{
			Content: &genai.Content{
				Role:  "model",
				Parts: []*genai.Part{{Text: response}},
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

	fmt.Println("\n========================================")
	fmt.Println("   6G AI CORE AGENT GATEWAY (PoC)       ")
	fmt.Println("========================================")
	fmt.Println("Available Workers: [CONNECTION_AGENT]")
	fmt.Println("Type 'exit' to quit.")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\nIntent > ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}
		if input == "" {
			continue
		}

		response, err := agents.RouteIntent(ctx, systemAgent, connectionAgent, input)
		if err != nil {
			fmt.Printf("Error processing intent: %v\n", err)
			continue
		}

		fmt.Printf("\n[Final Gateway Output]\n%s\n", response)
	}
	fmt.Println("\nExiting 6G Agent Gateway...")
}
