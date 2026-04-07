package workshop

// StartRunRequest represents the initial request to start a skill generation run.
type StartRunRequest struct {
	Type                 string        `json:"type"`
	RunID                string        `json:"run_id"`
	Messages             []ChatMessage `json:"messages"`
	ReasoningEnabled     bool          `json:"reasoning_enabled"`
	CurrentSkillMarkdown string        `json:"current_skill_markdown"`
	CurrentSkillYAML     string        `json:"current_skill_yaml"` // Legacy support
}

// ChatMessage represents a single message in the conversation history.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// StreamEvent represents a generic event sent over the WebSocket.
type StreamEvent struct {
	RunID     string `json:"run_id"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Data      any    `json:"data"`
}

// NormalizedSessionEvent is a UI-friendly version of an ADK event.
type NormalizedSessionEvent struct {
	ID            string         `json:"id"`
	Timestamp     string         `json:"timestamp"`
	InvocationID  string         `json:"invocation_id"`
	Author        string         `json:"author"`
	Text          string         `json:"text,omitempty"`
	Thought       string         `json:"thought,omitempty"`
	Partial       bool           `json:"partial"`
	TurnComplete  bool           `json:"turn_complete"`
	FinalResponse bool           `json:"final_response"`
	StateDelta    map[string]any `json:"state_delta,omitempty"`
}
