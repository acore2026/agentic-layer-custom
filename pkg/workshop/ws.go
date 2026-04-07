package workshop

import (
	"agentic-layer-custom/pkg/tools"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// HandleAgentRun manages the WebSocket connection for a skill generation run.
func HandleAgentRun(w http.ResponseWriter, r *http.Request, orchestrator *Orchestrator) {
	if r.Method == "OPTIONS" {
		WriteCORS(w, r)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[Workshop] Upgrade failed: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("[Workshop] Client connected to /ws/agent-run")

	var start StartRunRequest
	if err := conn.ReadJSON(&start); err != nil {
		_ = conn.WriteJSON(StreamEvent{
			Type: "run_error",
			Data: map[string]any{
				"message": "Failed to read run request.",
				"detail":  err.Error(),
			},
		})
		return
	}

	if start.Type != "start_run" {
		_ = conn.WriteJSON(StreamEvent{
			Type: "run_error",
			Data: map[string]any{
				"message": "First socket message must be start_run.",
			},
		})
		return
	}

	var writeMu sync.Mutex
	emit := func(event StreamEvent) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.WriteJSON(event)
	}

	if err := orchestrator.Run(r.Context(), start, emit); err != nil {
		_ = conn.WriteJSON(StreamEvent{
			RunID: start.RunID,
			Type:  "run_error",
			Data: map[string]any{
				"message": "Agent run failed.",
				"detail":  err.Error(),
			},
		})
	}
}

// HandleToolsCatalog serves the normalized tool catalog.
func HandleToolsCatalog(w http.ResponseWriter, r *http.Request) {
	WriteCORS(w, r)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	catalog := tools.GetNormalizedToolCatalog()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(catalog)
}

func WriteCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}
