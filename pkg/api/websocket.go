package api

import (
	"agentic-layer-custom/pkg/telemetry"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
	gateway agent.Agent
}

func NewLauncher(gateway agent.Agent) weblauncher.Sublauncher {
	return &WebsocketSublauncher{gateway: gateway}
}

func (l *WebsocketSublauncher) Keyword() string { return "ws" }
func (l *WebsocketSublauncher) Parse(args []string) ([]string, error) { return args, nil }
func (l *WebsocketSublauncher) CommandLineSyntax() string { return "" }
func (l *WebsocketSublauncher) SimpleDescription() string { return "Adds WebSocket intent streaming API" }

func (l *WebsocketSublauncher) SetupSubrouters(router *mux.Router, config *launcher.Config) error {
	router.HandleFunc("/v1/intents/stream", func(w http.ResponseWriter, r *http.Request) {
		HandleIntentsStream(w, r, l.gateway)
	})
	return nil
}

func (l *WebsocketSublauncher) UserMessage(webURL string, printer func(v ...any)) {
	printer(fmt.Sprintf("       ws:     intent stream active at %s/v1/intents/stream", webURL))
}

// IntentRequest represents the message from the React frontend.
type IntentRequest struct {
	Type string `json:"type"`
	Data struct {
		Intent     string `json:"intent"`
		ScenarioID string `json:"scenarioId"`
	} `json:"data"`
}

// HandleIntentsStream manages the WebSocket connection and orchestrates the agent run.
func HandleIntentsStream(w http.ResponseWriter, r *http.Request, gateway agent.Agent) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[WS] Upgrade failed: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("[WS] Client connected")

	// Subscribe to telemetry hub
	telemetryChan := telemetry.GetHub().Subscribe()
	defer telemetry.GetHub().Unsubscribe(telemetryChan)

	// Goroutine to pump telemetry to WebSocket
	go func() {
		for event := range telemetryChan {
			msg, _ := json.Marshal(event)
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("[WS] Read failed: %v\n", err)
			return
		}

		var req IntentRequest
		if err := json.Unmarshal(p, &req); err != nil {
			continue
		}

		if req.Type == "execute_intent" {
			fmt.Printf("[WS] Executing intent: %s\n", req.Data.Intent)
			
			// Run the gateway agent
			go func() {
				ctx := context.Background()
				ss := session.InMemoryService()
				run, err := runner.New(runner.Config{
					AppName:        "6G-AI-Core",
					Agent:          gateway,
					SessionService: ss,
				})
				if err != nil {
					fmt.Printf("[WS] Runner creation failed: %v\n", err)
					return
				}

				sessResp, _ := ss.Create(ctx, &session.CreateRequest{
					AppName: "6G-AI-Core",
					UserID:  "web-user",
				})

				msg := &genai.Content{
					Role:  "user",
					Parts: []*genai.Part{{Text: req.Data.Intent}},
				}

				// The GatewayAgent internally handles its own telemetry emission
				// We just need to drive the execution
				for _, err := range run.Run(ctx, "web-user", sessResp.Session.ID(), msg, agent.RunConfig{}) {
					if err != nil {
						fmt.Printf("[WS] Agent run error: %v\n", err)
						break
					}
				}
			}()
		}
	}
}
