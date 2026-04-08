package workshop

import (
	"agentic-layer-custom/pkg/model"
	"agentic-layer-custom/pkg/model/kimi"
	"agentic-layer-custom/pkg/tools"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/adk/agent"
	adkmodel "google.golang.org/adk/model"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

const maxCheckerAttempts = 3

// ServiceAgent manages the lifecycle of a skill generation run.
type ServiceAgent struct {
	appName        string
	sessionService session.Service
}

func NewServiceAgent() *ServiceAgent {
	return &ServiceAgent{
		appName:        "Service-Agent",
		sessionService: session.InMemoryService(),
	}
}

func (s *ServiceAgent) Run(ctx context.Context, req StartRunRequest, emit func(StreamEvent) error) error {
	runID := strings.TrimSpace(req.RunID)
	if runID == "" {
		runID = uuid.NewString()
	}

	userPrompt := latestUserPrompt(req.Messages)
	sessionID := "service-" + runID
	userID := "service-user"

	catalog := tools.GetNormalizedToolCatalog()
	
	// Use environment variables for LLM configuration
	provider := os.Getenv("LLM_PROVIDER")
	var baseLLM adkmodel.LLM

	if provider == "kimi" {
		apiKey := os.Getenv("KIMI_API_KEY")
		modelName := os.Getenv("KIMI_MODEL")
		if modelName == "" {
			modelName = "moonshot-v1-8k"
		}
		baseLLM = kimi.NewKimiModel(apiKey, modelName)
	} else {
		apiKey := os.Getenv("OPENAI_API_KEY")
		baseURL := os.Getenv("OPENAI_BASE_URL")
		modelName := os.Getenv("OPENAI_MODEL_NAME")
		if modelName == "" {
			modelName = "gpt-4o"
		}
		baseLLM = model.NewOpenAICompatibleLLM(modelName, baseURL, apiKey)
	}

	analysisLLM := baseLLM
	writerLLM := baseLLM
	if compatible, ok := baseLLM.(*model.OpenAICompatibleLLM); ok {
		analysisLLM = compatible.WithThinkingEnabled(false)
		writerLLM = compatible.WithThinkingEnabled(req.ReasoningEnabled)
	}

	serviceAgents, err := BuildServiceAgents(analysisLLM, writerLLM)
	if err != nil {
		return err
	}

	pipelineRunner, err := runner.New(runner.Config{
		AppName:        s.appName,
		Agent:          serviceAgents.Pipeline,
		SessionService: s.sessionService,
	})
	if err != nil {
		return fmt.Errorf("create pipeline runner: %w", err)
	}

	checkerRunner, err := runner.New(runner.Config{
		AppName:        s.appName,
		Agent:          serviceAgents.Checker,
		SessionService: s.sessionService,
	})
	if err != nil {
		return fmt.Errorf("create checker runner: %w", err)
	}

	initialState := buildInitialState(req, catalog)

	if _, err := s.sessionService.Create(ctx, &session.CreateRequest{
		AppName:   s.appName,
		UserID:    userID,
		SessionID: sessionID,
		State:     initialState,
	}); err != nil {
		return fmt.Errorf("create service session: %w", err)
	}

	log.Printf("[ServiceAgent] run started: run_id=%s session_id=%s", runID, sessionID)
	if err := emit(StreamEvent{
		RunID:     runID,
		Type:      "run_started",
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      map[string]any{"mode": "adk_runner"},
	}); err != nil {
		return err
	}

	if err := emitStatusSessionEvent(runID, "intent_analysis_agent", "Analyzing request.", emit); err != nil {
		return err
	}

	if err := s.runAgent(ctx, pipelineRunner, userID, sessionID, userPrompt, nil, runID, emit); err != nil {
		return err
	}

	runState, err := s.loadSessionState(ctx, userID, sessionID)
	if err != nil {
		return err
	}

	rawDraft := formatStateString(runState[StateWriterMarkdownDraft], "")
	currentDraft := extractMarkdownArtifact(rawDraft)
	if strings.TrimSpace(currentDraft) == "" {
		log.Printf("[ServiceAgent] ERROR: skill writer agent did not emit a markdown draft. rawDraft content: %q", rawDraft)
		return fmt.Errorf("skill writer agent did not emit a markdown draft (raw length: %d)", len(rawDraft))
	}

	log.Printf("[ServiceAgent] Validating draft (length: %d). content preview: %q", len(currentDraft), limitString(currentDraft, 100))
	issues := validateMarkdownSkill(currentDraft, catalog)
	if len(issues) == 0 {
		if err := emitStatusSessionEvent(runID, "markdown_format_checker_agent", "Skill format check passed.", emit); err != nil {
			return err
		}
		if err := emit(StreamEvent{
			RunID:     runID,
			Type:      "session_event",
			Timestamp: time.Now().Format(time.RFC3339),
			Data: NormalizedSessionEvent{
				ID:            uuid.NewString(),
				Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
				InvocationID:  uuid.NewString(),
				Author:        "markdown_format_checker_agent",
				Partial:       false,
				TurnComplete:  true,
				FinalResponse: true,
				StateDelta: map[string]any{
					StateSkillMarkdown: currentDraft,
				},
			},
		}); err != nil {
			return err
		}
		log.Printf("[ServiceAgent] run completed without checker rewrite: run_id=%s session_id=%s", runID, sessionID)
		return emit(StreamEvent{
			RunID:     runID,
			Type:      "run_complete",
			Timestamp: time.Now().Format(time.RFC3339),
			Data:      map[string]any{"status": "completed"},
		})
	}

	for attempt := 1; attempt <= maxCheckerAttempts; attempt++ {
		log.Printf("[ServiceAgent] checker attempt %d: issues=%v", attempt, issues)
		statusText := "Checking skill format."
		if len(issues) > 0 && attempt > 1 {
			statusText = "Fixing skill format."
		}
		if err := emitStatusSessionEvent(runID, "markdown_format_checker_agent", statusText, emit); err != nil {
			return err
		}

		stateDelta := map[string]any{
			StateWriterMarkdownDraft: currentDraft,
			StateCheckerAttemptCount: attempt,
			StateCheckerIssues:       formatMarkdownIssues(issues),
		}
		if err := s.runAgent(ctx, checkerRunner, userID, sessionID, "Validate and correct the markdown skill draft.", stateDelta, runID, emit); err != nil {
			return err
		}

		runState, err = s.loadSessionState(ctx, userID, sessionID)
		if err != nil {
			return err
		}

		rawChecked := formatStateString(runState[StateSkillMarkdown], "")
		checkedDraft := extractMarkdownArtifact(rawChecked)
		if strings.TrimSpace(checkedDraft) == "" {
			log.Printf("[ServiceAgent] ERROR: checker agent did not emit a markdown skill document. rawChecked content: %q", rawChecked)
			return fmt.Errorf("markdown format checker agent did not emit a markdown skill document (raw length: %d)", len(rawChecked))
		}

		log.Printf("[ServiceAgent] Validating checked draft (length: %d). content preview: %q", len(checkedDraft), limitString(checkedDraft, 100))
		issues = validateMarkdownSkill(checkedDraft, catalog)
		if len(issues) == 0 {
			log.Printf("[ServiceAgent] run completed: run_id=%s session_id=%s", runID, sessionID)
			return emit(StreamEvent{
				RunID:     runID,
				Type:      "run_complete",
				Timestamp: time.Now().Format(time.RFC3339),
				Data:      map[string]any{"status": "completed"},
			})
		}

		currentDraft = checkedDraft
	}

	return fmt.Errorf("markdown format checker could not produce valid markdown after %d attempts: %s", maxCheckerAttempts, strings.Join(issues, "; "))
}

func (s *ServiceAgent) runAgent(
	ctx context.Context,
	adkRunner *runner.Runner,
	userID string,
	sessionID string,
	prompt string,
	stateDelta map[string]any,
	runID string,
	emit func(StreamEvent) error,
) error {
	runConfig := agent.RunConfig{StreamingMode: agent.StreamingModeSSE}
	userContent := genai.NewContentFromText(prompt, genai.RoleUser)
	analysisCompleted := false
	writerStarted := false

	runOptions := []runner.RunOption{}
	if len(stateDelta) > 0 {
		runOptions = append(runOptions, runner.WithStateDelta(stateDelta))
	}

	for event, err := range adkRunner.Run(ctx, userID, sessionID, userContent, runConfig, runOptions...) {
		if err != nil {
			return fmt.Errorf("runner execution failed: %w", err)
		}
		if event == nil || event.Author == "user" {
			continue
		}
		
		normalized := normalizeADKEvent(event)
		
		if normalized.Author == "skill_writer_agent" && !writerStarted {
			if err := emitStatusSessionEvent(runID, "skill_writer_agent", "Starting skill draft.", emit); err != nil {
				return err
			}
			writerStarted = true
		}

		if err := emit(StreamEvent{
			RunID:     runID,
			Type:      "session_event",
			Timestamp: time.Now().Format(time.RFC3339),
			Data:      normalized,
		}); err != nil {
			return err
		}

		if normalized.Author == "intent_analysis_agent" && normalized.FinalResponse && !analysisCompleted {
			if err := emitStatusSessionEvent(runID, "intent_analysis_agent", "Analysis completed.", emit); err != nil {
				return err
			}
			if !writerStarted {
				if err := emitStatusSessionEvent(runID, "skill_writer_agent", "Starting skill draft.", emit); err != nil {
					return err
				}
				writerStarted = true
			}
			analysisCompleted = true
		}

		if normalized.Author == "skill_writer_agent" && normalized.FinalResponse {
			if err := emitStatusSessionEvent(runID, "skill_writer_agent", "Skill draft completed.", emit); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *ServiceAgent) loadSessionState(ctx context.Context, userID string, sessionID string) (map[string]any, error) {
	sessionResponse, err := s.sessionService.Get(ctx, &session.GetRequest{
		AppName:   s.appName,
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		return nil, fmt.Errorf("load service session: %w", err)
	}

	values := map[string]any{}
	for key, value := range sessionResponse.Session.State().All() {
		values[key] = value
	}
	return values, nil
}

// Helpers

func buildInitialState(req StartRunRequest, catalog tools.NormalizedToolCatalog) map[string]any {
	userPrompt := latestUserPrompt(req.Messages)
	category, knowledge := resolveKnowledgeCase(userPrompt)
	
	return map[string]any{
		StateConversationTranscript: formatMessages(req.Messages),
		StateCurrentSkillMarkdown:   strings.TrimSpace(req.CurrentSkillMarkdown),
		StateLatestUserRequest:      userPrompt,
		StateNormalizedPrompt:       normalizePrompt(userPrompt),
		StateToolCatalogJSON:        formatToolCatalogForPrompt(catalog),
		StateToolShortlist:          formatToolCatalogForPrompt(catalog), // Simplified for now
		StateIntentCategory:         string(category),
		StateKnowledgeCase:          knowledge.ID,
		StateKnowledgeBrief:         formatKnowledgeBrief(knowledge),
		StateCheckerAttemptCount:    0,
	}
}

func latestUserPrompt(messages []ChatMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if strings.EqualFold(strings.TrimSpace(messages[i].Role), "user") && strings.TrimSpace(messages[i].Content) != "" {
			return strings.TrimSpace(messages[i].Content)
		}
	}
	return "Generate a 6G signaling skill."
}

func normalizePrompt(prompt string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(prompt)), " ")
}

func formatMessages(messages []ChatMessage) string {
	var sb strings.Builder
	for _, m := range messages {
		sb.WriteString(fmt.Sprintf("[%s] %s\n\n", m.Role, m.Content))
	}
	return sb.String()
}

func formatToolCatalogForPrompt(catalog tools.NormalizedToolCatalog) string {
	var sb strings.Builder
	for _, t := range catalog.Tools {
		sb.WriteString(fmt.Sprintf("- %s: %s (Params: %s)\n", t.Name, t.Description, strings.Join(t.AllParams, ", ")))
	}
	return sb.String()
}

func extractMarkdownArtifact(text string) string {
	trimmed := strings.TrimSpace(text)
	// Only extract if the entire text is wrapped in a markdown/md code block.
	// We check for the starting ``` and ending ```.
	if strings.HasPrefix(trimmed, "```") && strings.HasSuffix(trimmed, "```") {
		re := regexp.MustCompile("(?s)^```(?:markdown|md)?\\s*(.*?)```$")
		if match := re.FindStringSubmatch(trimmed); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return trimmed
}

func normalizeADKEvent(event *session.Event) NormalizedSessionEvent {
	text := ""
	thought := ""
	for _, part := range event.Content.Parts {
		if part.Thought {
			thought += part.Text
		} else {
			text += part.Text
		}
	}
	
	return NormalizedSessionEvent{
		ID:            uuid.NewString(),
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		InvocationID:  uuid.NewString(),
		Author:        event.Author,
		Text:          text,
		Thought:       thought,
		Partial:       false, // Assuming non-stream for simplicity in initial port
		TurnComplete:  event.TurnComplete,
		FinalResponse: event.TurnComplete,
	}
}

func emitStatusSessionEvent(runID string, author string, text string, emit func(StreamEvent) error) error {
	return emit(StreamEvent{
		RunID:     runID,
		Type:      "session_event",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: NormalizedSessionEvent{
			ID:            uuid.NewString(),
			Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
			InvocationID:  uuid.NewString(),
			Author:        author,
			Partial:       false,
			TurnComplete:  false,
			FinalResponse: false,
			Text:          text,
		},
	})
}

func formatStateString(value any, fallback string) string {
	if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
		return strings.TrimSpace(text)
	}
	return fallback
}
