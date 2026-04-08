package workshop

import (
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// State keys for the workshop workflow
const (
	StateConversationTranscript = "service.conversation_transcript"
	StateCurrentSkillMarkdown   = "service.current_skill_markdown"
	StateLatestUserRequest      = "service.latest_user_request"
	StateNormalizedPrompt       = "service.normalized_prompt"
	StateToolCatalogJSON        = "service.tool_catalog_json"
	StateToolShortlist          = "service.tool_shortlist"
	StateIntentCategory         = "service.intent_category"
	StateIntentAnalysisSummary  = "service.intent_analysis_summary"
	StateWriterMarkdownDraft    = "service.writer_markdown_draft"
	StateCheckerIssues          = "service.checker_issues"
	StateCheckerAttemptCount    = "service.checker_attempt_count"
	StateSkillMarkdown          = "service.skill_markdown"
	StateKnowledgeCase          = "service.knowledge_case"
	StateKnowledgeBrief         = "service.knowledge_brief"
)

// ServiceAgents holds the agents used in the service generation workflow
type ServiceAgents struct {
	Pipeline agent.Agent
	Checker  agent.Agent
}

// BuildServiceAgents creates the multi-agent workflow for service generation
func BuildServiceAgents(analysisLLM model.LLM, llm model.LLM) (*ServiceAgents, error) {
	intentAnalysisAgent, err := llmagent.New(llmagent.Config{
		Name:                "intent_analysis_agent",
		Description:         "Analyzes the request, confirms the intent category, and summarizes the workflow direction.",
		Model:               analysisLLM,
		InstructionProvider: IntentAnalysisInstructionProvider,
		GenerateContentConfig: &genai.GenerateContentConfig{
			MaxOutputTokens: 80,
		},
		OutputKey: StateIntentAnalysisSummary,
	})
	if err != nil {
		return nil, fmt.Errorf("create intent analysis agent: %w", err)
	}

	writerAgent, err := llmagent.New(llmagent.Config{
		Name:                "skill_writer_agent",
		Description:         "Writes the markdown skill document from the request, category, and domain knowledge.",
		Model:               llm,
		InstructionProvider: WriterInstructionProvider,
		OutputKey:           StateWriterMarkdownDraft,
	})
	if err != nil {
		return nil, fmt.Errorf("create skill writer agent: %w", err)
	}

	checkerAgent, err := llmagent.New(llmagent.Config{
		Name:                "markdown_format_checker_agent",
		Description:         "Checks and repairs markdown format and consistency before finalizing the skill document.",
		Model:               llm,
		InstructionProvider: CheckerInstructionProvider,
		OutputKey:           StateSkillMarkdown,
	})
	if err != nil {
		return nil, fmt.Errorf("create markdown format checker agent: %w", err)
	}

	rootAgent, err := sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:        "skill_generation_workflow",
			Description: "Runs intent analysis and markdown skill writing in a fixed order.",
			SubAgents:   []agent.Agent{intentAnalysisAgent, writerAgent},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create root workflow agent: %w", err)
	}

	return &ServiceAgents{
		Pipeline: rootAgent,
		Checker:  checkerAgent,
	}, nil
}

// Instruction Providers ported from the workshop backend

func IntentAnalysisInstructionProvider(ctx agent.ReadonlyContext) (string, error) {
	stateValues := collectStateValues(ctx.ReadonlyState())
	return fmt.Sprintf(
		"You are the Intent Analysis agent for a markdown skill generator.\n\nWrite one very short user-facing analysis summary before skill drafting begins.\n\nRules:\n- Output concise prose only.\n- Do not output YAML or markdown skill content.\n- Do not narrate tool-by-tool in detail.\n- Do not repeat the request.\n- Start writing the summary immediately. Stream the lines directly as you compose them.\n- Stop after exactly 3 short lines.\n- Each line must be under 14 words.\n\nOutput format:\nCategory: <fixed category>\nGoal: <one sentence>\nWorkflow: <one sentence>\n\nNormalized prompt:\n%s\n\nFixed intent category:\n%s\n\nMatched domain knowledge:\n%s\n\nLikely tool shortlist:\n%s",
		formatStateString(stateValues[StateNormalizedPrompt], "No prompt provided."),
		formatStateString(stateValues[StateIntentCategory], "ACN"),
		formatStateString(stateValues[StateKnowledgeBrief], "No domain reference matched the current request."),
		formatStateString(stateValues[StateToolShortlist], "No tools available."),
	), nil
}

func WriterInstructionProvider(ctx agent.ReadonlyContext) (string, error) {
	stateValues := collectStateValues(ctx.ReadonlyState())
	return fmt.Sprintf(
		"You are the Skill Writer for a markdown skill generator.\n\nWrite the final markdown skill document directly.\n\nRules:\n- Output markdown only.\n- The response itself must be the markdown document.\n- Do not wrap the full document in ``` fences, ```yaml fences, ```markdown fences, or any surrounding code block.\n- The document must begin immediately with the YAML front matter line `---`.\n- Use only tool names from the provided tool catalog.\n- Do not output YAML IR.\n- Do not invent parameter wiring, parameter placeholders, or parameter values.\n- Prefer the shortest valid workflow that matches the intent and domain guidance.\n- For now, keep the workflow strictly linear: no IF, no ELSE, no ABORT branch.\n- End the workflow with OUTPUT \"DONE\".\n\nThe markdown document must use this exact section order:\n1. YAML front matter\n2. H1 title\n3. ## Overview\n4. ## Tool Inventory\n5. ## Workflow\n6. ## Critical Rules\n7. ## Output Format\n\nFormatting rules:\n- Front matter must contain name and description.\n- Title must be exactly '# {name} Skill'.\n- Tool Inventory must list only the selected tools.\n- Workflow must be a fenced python pseudo-code block that uses only ordered CALL \"ToolName\" lines followed by OUTPUT \"DONE\".\n- Critical Rules must be concise and operational.\n- Output Format must list the same ordered tool names on separate lines followed by DONE.\n\nIntent category:\n%s\n\nIntent analysis summary:\n%s\n\nMatched domain knowledge:\n%s\n\nCurrent markdown skill:\n%s\n\nTool catalog summary:\n%s",
		formatStateString(stateValues[StateIntentCategory], "ACN"),
		formatStateString(stateValues[StateIntentAnalysisSummary], "No prior analysis summary available."),
		formatStateString(stateValues[StateKnowledgeBrief], "No domain reference matched the current request."),
		formatStateString(stateValues[StateCurrentSkillMarkdown], "No current markdown skill document exists yet."),
		formatStateString(stateValues[StateToolCatalogJSON], "No tools available."),
	), nil
}

func CheckerInstructionProvider(ctx agent.ReadonlyContext) (string, error) {
	stateValues := collectStateValues(ctx.ReadonlyState())
	return fmt.Sprintf(
		"You are the Markdown Format Checker for a markdown skill generator.\n\nValidate the current markdown draft and return a corrected markdown document.\n\nRules:\n- Output markdown only.\n- The response itself must be the markdown document.\n- Do not wrap the full document in ``` fences, ```yaml fences, ```markdown fences, or any surrounding code block.\n- The document must begin immediately with the YAML front matter line `---`.\n- If the draft is already valid, return it unchanged.\n- Preserve intent and tool sequence.\n- Fix structure, formatting, tool grounding, and workflow consistency.\n- Do not emit YAML IR, explanations, or extra prose outside the markdown document.\n- Keep the workflow strictly linear: no IF, no ELSE, no ABORT branch.\n\nRequired markdown structure:\n1. YAML front matter\n2. H1 title\n3. ## Overview\n4. ## Tool Inventory\n5. ## Workflow\n6. ## Critical Rules\n7. ## Output Format\n\nFormatting rules:\n- Workflow must be a fenced python pseudo-code block.\n- Use only tools from the catalog.\n- Workflow lines must be ordered CALL \"ToolName\" entries followed by OUTPUT \"DONE\".\n- Output Format must list the same ordered tool names followed by DONE.\n\nIntent category:\n%s\n\nMatched domain knowledge:\n%s\n\nValidation issues to fix:\n%s\n\nCurrent markdown draft:\n%s\n\nTool catalog summary:\n%s",
		formatStateString(stateValues[StateIntentCategory], "ACN"),
		formatStateString(stateValues[StateKnowledgeBrief], "No domain reference matched the current request."),
		formatStateString(stateValues[StateCheckerIssues], "No issues provided. Return the markdown unchanged if it is valid."),
		formatStateString(stateValues[StateWriterMarkdownDraft], ""),
		formatStateString(stateValues[StateToolCatalogJSON], "No tools available."),
	), nil
}

// Helper functions ported from the workshop backend

func collectStateValues(state session.ReadonlyState) map[string]any {
	values := map[string]any{}
	for key, value := range state.All() {
		values[key] = value
	}
	return values
}
