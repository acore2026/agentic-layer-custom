package telemetry

import (
	"sync"
)

// EventType defines the possible types of telemetry events.
type EventType string

const (
	EventTypeAIPayload        EventType = "ai_payload"
	EventTypeLLMThought       EventType = "llm_thought"
	EventTypeNetworkPCAP      EventType = "network_pcap"
	EventTypeWorkflowComplete EventType = "workflow_complete"
)

// TelemetryEvent is the top-level JSON structure sent over the WebSocket.
type TelemetryEvent struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data"`
}

// AIPayloadData represents LLM prompt and response payloads.
type AIPayloadData struct {
	Agent   string `json:"agent"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMThoughtData represents granular reasoning chunks.
type LLMThoughtData struct {
	Agent string `json:"agent"`
	Chunk string `json:"chunk"`
}

// NetworkPCAPData represents simulated signaling traffic.
type NetworkPCAPData struct {
	Direction   string      `json:"direction"`
	Source      string      `json:"source"`
	Destination string      `json:"destination"`
	Protocol    string      `json:"protocol"`
	Info        string      `json:"info"`
	Details     interface{} `json:"details"`
}

// WorkflowCompleteData indicates the end of the orchestration.
type WorkflowCompleteData struct {
	Status       string `json:"status"`
	FinalMessage string `json:"finalMessage"`
}

// TelemetryHub handles thread-safe event broadcasting.
type TelemetryHub struct {
	mu        sync.Mutex
	listeners []chan TelemetryEvent
}

var globalHub = &TelemetryHub{}

// GetHub returns the global telemetry hub.
func GetHub() *TelemetryHub {
	return globalHub
}

// Emit broadcasts an event to all registered listeners.
func (h *TelemetryHub) Emit(event TelemetryEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ch := range h.listeners {
		// Non-blocking send to avoid hanging agents if a listener is slow
		select {
		case ch <- event:
		default:
		}
	}
}

// Subscribe registers a new listener channel.
func (h *TelemetryHub) Subscribe() chan TelemetryEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(chan TelemetryEvent, 100)
	h.listeners = append(h.listeners, ch)
	return ch
}

// Unsubscribe removes a listener channel.
func (h *TelemetryHub) Unsubscribe(ch chan TelemetryEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, listener := range h.listeners {
		if listener == ch {
			h.listeners = append(h.listeners[:i], h.listeners[i+1:]...)
			close(ch)
			break
		}
	}
}
