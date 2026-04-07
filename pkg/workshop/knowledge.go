package workshop

import (
	"fmt"
	"strings"
)

type intentCategory string

const (
	intentCategoryACN         intentCategory = "ACN"
	intentCategoryQoS         intentCategory = "QoS"
	intentCategoryComputation intentCategory = "Computation"
)

type knowledgeCase struct {
	ID               string
	Category         intentCategory
	MatchAll         []string
	Name             string
	Description      string
	Overview         string
	StepTitles       map[string]string
	ToolHints        []string
	FailureLabel     string
	SuccessLabel     string
	ContinuityRules  []string
	CriticalRules    []string
	SuccessCondition string
}

var acnKnowledgeCase = knowledgeCase{
	ID:          "acn_embodied_agent_subnet",
	Category:    intentCategoryACN,
	MatchAll:    []string{"embodied agent", "network subnet"},
	Name:        "ACN",
	Description: "Process of connecting embodied agents to the network.",
	Overview:    "Connect embodied agents to a dedicated network subnet through a short, ordered tool workflow.",
	ToolHints:   []string{"Start by checking subscription eligibility.", "Set up the subnet context before issuing access credentials.", "Issue and then validate access credentials before session creation.", "Finish by creating the subnet PDU session."},
}

var qosKnowledgeCase = knowledgeCase{
	ID:          "qos_generic",
	Category:    intentCategoryQoS,
	Name:        "QoS",
	Description: "Quality-of-service workflow requests such as bandwidth, latency, or priority handling.",
}

var computationKnowledgeCase = knowledgeCase{
	ID:          "computation_generic",
	Category:    intentCategoryComputation,
	Name:        "Computation",
	Description: "Computation workflow requests such as offloading, placement, or resource selection.",
}

func detectIntentCategory(prompt string) intentCategory {
	normalized := strings.ToLower(strings.TrimSpace(prompt))
	if normalized == "" {
		return intentCategoryACN
	}

	isACN := true
	for _, fragment := range acnKnowledgeCase.MatchAll {
		if !strings.Contains(normalized, fragment) {
			isACN = false
			break
		}
	}
	if isACN {
		return intentCategoryACN
	}

	qosKeywords := []string{"qos", "latency", "bandwidth", "throughput", "priority", "turbo mode", "gaming"}
	for _, keyword := range qosKeywords {
		if strings.Contains(normalized, keyword) {
			return intentCategoryQoS
		}
	}

	computationKeywords := []string{"compute", "computation", "offload", "placement", "gpu", "cpu", "workload", "resource"}
	for _, keyword := range computationKeywords {
		if strings.Contains(normalized, keyword) {
			return intentCategoryComputation
		}
	}

	return intentCategoryACN
}

func resolveKnowledgeCase(prompt string) (intentCategory, *knowledgeCase) {
	category := detectIntentCategory(prompt)
	switch category {
	case intentCategoryQoS:
		return category, &qosKnowledgeCase
	case intentCategoryComputation:
		return category, &computationKnowledgeCase
	default:
		return intentCategoryACN, &acnKnowledgeCase
	}
}

func formatKnowledgeBrief(knowledge *knowledgeCase) string {
	if knowledge == nil {
		return "No domain reference matched the current request."
	}

	brief := fmt.Sprintf("Domain: %s\nReference: %s", knowledge.Name, knowledge.Description)
	if len(knowledge.ToolHints) > 0 {
		brief += fmt.Sprintf("\nWorkflow hints: %s", strings.Join(knowledge.ToolHints, " | "))
	}
	return brief
}
