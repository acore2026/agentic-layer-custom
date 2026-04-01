package agents

import (
	"agentic-layer-custom/pkg/tools"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// LoadSkills reads and concatenates all SKILL.md content from the specified directory.
func LoadSkills(skillDir string) (string, error) {
	var sb strings.Builder
	err := filepath.Walk(skillDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Base(path) == "SKILL.md" {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			sb.WriteString(fmt.Sprintf("--- SKILL SOURCE: %s ---\n", path))
			sb.WriteString(string(content))
			sb.WriteString("\n\n")
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

// DiscoverTools extracts unique tool names from the skill content using the CALL pattern.
func DiscoverTools(content string) []string {
	re := regexp.MustCompile(`CALL "([^"]+)"`)
	matches := re.FindAllStringSubmatch(content, -1)
	uniqueTools := make(map[string]bool)
	var tools []string
	for _, match := range matches {
		if len(match) > 1 {
			name := match[1]
			if !uniqueTools[name] {
				uniqueTools[name] = true
				tools = append(tools, name)
			}
		}
	}
	return tools
}

// NewConnectionAgent creates a worker agent for connection-related intents.
func NewConnectionAgent(m model.LLM, skillDir string) (agent.Agent, error) {
	// 1. Load Skills and Discover Tools
	skillContent, err := LoadSkills(skillDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load skills: %v", err)
	}
	toolNames := DiscoverTools(skillContent)

	// 2. Register Tools
	var agentTools []tool.Tool
	for _, name := range toolNames {
		localName := name // capture the loop variable
		t, err := functiontool.New(functiontool.Config{
			Name:        localName,
			Description: "Mock tool for " + localName + " (dynamically discovered)",
		}, func(ctx tool.Context, args map[string]any) (any, error) {
			return tools.UniversalMockTool(ctx, localName, args)
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create tool %s: %v", localName, err)
		}
		agentTools = append(agentTools, t)
	}

	// 3. Configure Agent
	config := llmagent.Config{
		Name:        "ConnectionAgent",
		Description: "A worker agent that handles connection-related intents using dynamically loaded signaling skills.",
		Instruction: fmt.Sprintf(`You are the Connection Agent of a 6G AI Core Network.
Your task is to process connection-related intents using the provided signaling skills.

DYNAMICALY LOADED SKILLS:
%s

ORCHESTRATION RULES:
1. Identify the relevant skill from the list above based on the user's intent.
2. Follow the pseudo-code workflow PRECISELY. Do not skip steps.
3. Use a ReAct (Reason + Act) approach for each step.
4. If a tool returns a 'token' or 'ue_id', you MUST pass it to subsequent tools that require it.
5. If any tool fails or returns a status other than SUCCESS, output "ABORT" and explain.

Once the workflow is DONE, provide a final summary to the user.`, skillContent),
		Model:           m,
		Tools:           agentTools,
		IncludeContents: llmagent.IncludeContentsDefault,
	}

	return llmagent.New(config)
}
