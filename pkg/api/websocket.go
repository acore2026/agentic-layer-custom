package api

import (
	"agentic-layer-custom/pkg/observability"
	"agentic-layer-custom/pkg/telemetry"
	"agentic-layer-custom/pkg/workshop"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/cmd/launcher"
	weblauncher "google.golang.org/adk/cmd/launcher/web"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WebsocketSublauncher implements weblauncher.Sublauncher.
type WebsocketSublauncher struct {
	gateway      agent.Agent
	serviceAgent *workshop.ServiceAgent
	langfuse     *observability.Langfuse
}

func NewLauncher(gateway agent.Agent, sa *workshop.ServiceAgent, langfuse *observability.Langfuse) weblauncher.Sublauncher {
	return &WebsocketSublauncher{
		gateway:      gateway,
		serviceAgent: sa,
		langfuse:     langfuse,
	}
}

func (l *WebsocketSublauncher) Keyword() string                       { return "ws" }
func (l *WebsocketSublauncher) Parse(args []string) ([]string, error) { return args, nil }
func (l *WebsocketSublauncher) CommandLineSyntax() string             { return "" }
func (l *WebsocketSublauncher) SimpleDescription() string {
	return "Adds WebSocket intent streaming and Service Agent API"
}

func (l *WebsocketSublauncher) SetupSubrouters(router *mux.Router, config *launcher.Config) error {
	// Unified handler for all paths to ensure maximum compatibility with frontend expectations
	handler := func(w http.ResponseWriter, r *http.Request) {
		HandleUnifiedWebSocket(w, r, l.gateway, l.serviceAgent, l.langfuse)
	}

	router.HandleFunc("/v1/intents/stream", handler)
	router.HandleFunc("/ws/agent-run", handler)
	router.HandleFunc("/ws", handler)

	router.HandleFunc("/api/health", HandleHealth).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/skills", HandleSkills).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/tools", workshop.HandleToolsCatalog).Methods("GET", "OPTIONS")
	return nil
}

func (l *WebsocketSublauncher) UserMessage(webURL string, printer func(v ...any)) {
	printer(fmt.Sprintf("       ws:     unified agent api active at %s/ws/agent-run", webURL))
	printer(fmt.Sprintf("       ws:     signaling stream active at %s/v1/intents/stream", webURL))
}

// IntentRequest represents the message from the React frontend.
type IntentRequest struct {
	Type string `json:"type"`
	Data struct {
		Intent     string `json:"intent"`
		ScenarioID string `json:"scenarioId"`
	} `json:"data"`
}

// HandleUnifiedWebSocket manages a WebSocket connection and dispatches to the appropriate engine
// based on the message type (signaling vs workshop).
func HandleUnifiedWebSocket(w http.ResponseWriter, r *http.Request, gateway agent.Agent, sa *workshop.ServiceAgent, langfuse *observability.Langfuse) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[WS] Upgrade failed: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Printf("[WS] Client connected to %s\n", r.URL.Path)

	var writeMu sync.Mutex

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("[WS] Connection closed: %v\n", err)
			return
		}

		var generic struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(p, &generic); err != nil {
			continue
		}

		switch generic.Type {
		case "execute_intent":
			var req IntentRequest
			if err := json.Unmarshal(p, &req); err != nil {
				continue
			}
			go handleSignaling(conn, &writeMu, req, gateway, langfuse)

		case "start_run":
			var req workshop.StartRunRequest
			if err := json.Unmarshal(p, &req); err != nil {
				continue
			}
			go handleWorkshop(conn, &writeMu, req, sa)

		default:
			fmt.Printf("[WS] Unknown message type: %s\n", generic.Type)
		}
	}
}

func handleSignaling(conn *websocket.Conn, mu *sync.Mutex, req IntentRequest, gateway agent.Agent, langfuse *observability.Langfuse) {
	fmt.Printf("[WS] Executing signaling intent: %s\n", req.Data.Intent)

	// Subscribe to telemetry hub
	telemetryChan := telemetry.GetHub().Subscribe()
	defer telemetry.GetHub().Unsubscribe(telemetryChan)

	// Goroutine to pump telemetry to WebSocket
	done := make(chan struct{})
	go func() {
		defer close(done)
		for event := range telemetryChan {
			msg, _ := json.Marshal(event)
			mu.Lock()
			err := conn.WriteMessage(websocket.TextMessage, msg)
			mu.Unlock()
			if err != nil {
				return
			}
		}
	}()

	// Run the gateway agent
	ctx := context.Background()
	ss := session.InMemoryService()
	run, err := runner.New(runner.Config{
		AppName:        "6G-AI-Core",
		Agent:          gateway,
		SessionService: ss,
		PluginConfig:   pluginConfig(langfuse),
	})
	if err != nil {
		fmt.Printf("[WS] Runner creation failed: %v\n", err)
		return
	}

	sessResp, _ := ss.Create(ctx, &session.CreateRequest{
		AppName: "6G-AI-Core",
		UserID:  "web-user",
	})
	ctx = decorateSignalingContext(ctx, langfuse, req, sessResp.Session.ID())

	msg := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: req.Data.Intent}},
	}

	for _, err := range run.Run(ctx, "web-user", sessResp.Session.ID(), msg, agent.RunConfig{}) {
		if err != nil {
			fmt.Printf("[WS] Agent run error: %v\n", err)
			break
		}
	}
}

func pluginConfig(langfuse *observability.Langfuse) runner.PluginConfig {
	if langfuse == nil {
		return runner.PluginConfig{}
	}
	return langfuse.PluginConfig
}

func decorateSignalingContext(ctx context.Context, langfuse *observability.Langfuse, req IntentRequest, sessionID string) context.Context {
	if langfuse == nil {
		return ctx
	}

	metadata := map[string]string{
		"app_name":     "6G-AI-Core",
		"route":        "signaling",
		"session_id":   sessionID,
		"message_type": req.Type,
	}
	if req.Data.ScenarioID != "" {
		metadata["scenario_id"] = req.Data.ScenarioID
	}

	return langfuse.DecorateContext(ctx, observability.TraceOptions{
		TraceName: "agent-gateway.execute_intent",
		UserID:    "web-user",
		Tags:      []string{"adk-go", "agent-gateway", "signaling"},
		Metadata:  metadata,
	})
}

func handleWorkshop(conn *websocket.Conn, mu *sync.Mutex, req workshop.StartRunRequest, sa *workshop.ServiceAgent) {
	fmt.Printf("[WS] Executing workshop run: %s\n", req.RunID)

	emit := func(event workshop.StreamEvent) error {
		mu.Lock()
		defer mu.Unlock()
		return conn.WriteJSON(event)
	}

	if err := sa.Run(context.Background(), req, emit); err != nil {
		fmt.Printf("[WS] Workshop run failed: %v\n", err)
		_ = emit(workshop.StreamEvent{
			RunID: req.RunID,
			Type:  "run_error",
			Data: map[string]any{
				"message": "Agent run failed.",
				"detail":  err.Error(),
			},
		})
	}
}
