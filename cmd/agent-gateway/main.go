package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"agentic-layer-custom/pkg/agents"
	"agentic-layer-custom/pkg/api"
	"agentic-layer-custom/pkg/model/kimi"
	"agentic-layer-custom/pkg/workshop"
	"strconv"

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
)

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
		m = kimi.NewKimiModel(apiKey, modelName)
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

	serviceAgent := workshop.NewServiceAgent()

	// Use GatewayAgent as the root, others as sub-workers
	loader, err := agent.NewMultiLoader(gatewayAgent, systemAgent, connectionAgent)
	if err != nil {
		log.Fatalf("Failed to create agent loader: %v", err)
	}

	port := 8080
	if p := os.Getenv("API_PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	}

	cfg := &launcher.Config{
		AgentLoader:    loader,
		SessionService: session.InMemoryService(),
	}

	l := universal.NewLauncher(
		console.NewLauncher(),
		web.NewLauncher(
			api.NewLauncher(gatewayAgent, serviceAgent),
			webapi.NewLauncher(),
			webui.NewLauncher(),
		),
	)

	fmt.Printf("Launching ADK Web UI on http://localhost:%d/ui/ ...\n", port)
	// The universal launcher needs "web" to activate the web server,
	// then "api" for REST, "webui" for dashboard, and "ws" for our custom stream.
	executeArgs := []string{"web", "--port", strconv.Itoa(port), "api", "webui", "ws"}
	if err := l.Execute(ctx, cfg, executeArgs); err != nil {
		log.Fatalf("Launcher execution failed: %v", err)
	}
}
