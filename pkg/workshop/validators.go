package workshop

import (
	"agentic-layer-custom/pkg/tools"
	"fmt"
	"regexp"
	"strings"
)

var (
	// More robust front matter pattern - just check for --- at the start
	markdownFrontMatterPattern = regexp.MustCompile(`(?s)^---\s*[\r\n].*?---\s*[\r\n]`)
	workflowBlockPattern       = regexp.MustCompile("(?s)## Workflow[\\s\\S]*?```(?:python|py)\\s*([\\s\\S]*?)```")
	callPattern                = regexp.MustCompile(`CALL\s+(?:"([^"]+)"|([A-Za-z0-9_]+))`)
	// Restrict to lines starting with identifier and ending with newline or EOF, or (
	outputFormatToolPattern    = regexp.MustCompile(`(?m)^([A-Za-z0-9_]+)(?:\(|\s*$)`)
)

func validateMarkdownSkill(markdown string, catalog tools.NormalizedToolCatalog) []string {
	trimmed := strings.TrimSpace(markdown)
	var issues []string

	if trimmed == "" {
		return []string{"markdown output is empty"}
	}

	fmt.Printf("[Validator] Validating markdown (len: %d). First 50 chars: %q\n", len(trimmed), limitString(trimmed, 50))

	if !strings.HasPrefix(trimmed, "---") {
		issues = append(issues, "markdown front matter is required (must start with ---)")
	} else {
		// If it has prefix, check if it has the closing ---
		if count := strings.Count(trimmed, "---"); count < 2 {
			issues = append(issues, "markdown front matter is incomplete (missing closing ---)")
		}
	}

	requiredSections := []string{
		"# ",
		"## Overview",
		"## Tool Inventory",
		"## Workflow",
		"## Critical Rules",
		"## Output Format",
	}
	for _, section := range requiredSections {
		if !strings.Contains(trimmed, section) {
			issues = append(issues, fmt.Sprintf("required markdown section missing: %s", section))
		}
	}

	workflowMatch := workflowBlockPattern.FindStringSubmatch(trimmed)
	if len(workflowMatch) < 2 {
		issues = append(issues, "workflow section must contain a fenced python block (```python ... ```)")
	} else {
		workflowBody := workflowMatch[1]
		workflowCalls := extractWorkflowCallNames(workflowBody)
		if len(workflowCalls) == 0 {
			issues = append(issues, "workflow block must contain at least one CALL step")
		}
		for _, toolName := range workflowCalls {
			if _, ok := catalog.ByName[toolName]; !ok {
				issues = append(issues, fmt.Sprintf("workflow references unknown tool %q", toolName))
			}
		}
		if strings.Contains(workflowBody, "IF ") || strings.Contains(workflowBody, "ELSE:") || strings.Contains(workflowBody, `OUTPUT "ABORT"`) {
			issues = append(issues, "workflow must be linear and must not include IF, ELSE, or ABORT branches")
		}
		if !strings.Contains(workflowBody, `OUTPUT "DONE"`) {
			issues = append(issues, `workflow must include OUTPUT "DONE"`)
		}
	}

	// Only validate tools that are actually in sections where they should be
	toolNames := extractAllReferencedToolNames(trimmed)
	for _, toolName := range toolNames {
		if _, ok := catalog.ByName[toolName]; !ok {
			// Ignore "DONE" which is part of the output format but not a tool
			if strings.ToUpper(toolName) != "DONE" {
				issues = append(issues, fmt.Sprintf("markdown references unknown tool %q", toolName))
			}
		}
	}

	if len(issues) > 0 {
		fmt.Printf("[Validator] Found %d issues: %v\n", len(issues), issues)
	} else {
		fmt.Printf("[Validator] Validation passed.\n")
	}

	return dedupeIssues(append(issues, validateWorkflowProcess(trimmed, catalog)...))
}

func limitString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func findSectionIndex(sections []string, index int, text string) int {
	for i, s := range sections {
		if strings.Index(text, s) == index {
			return i
		}
	}
	return 0
}

func extractAllReferencedToolNames(markdown string) []string {
	var names []string
	
	// 1. Tools in inventory
	inventoryPattern := regexp.MustCompile("(?m)^-\\s+`([^`]+)`")
	if invStart := strings.Index(markdown, "## Tool Inventory"); invStart >= 0 {
		invEnd := strings.Index(markdown[invStart:], "##")
		if invEnd < 0 {
			invEnd = len(markdown) - invStart
		} else {
			invEnd += invStart
		}
		invSection := markdown[invStart:invEnd]
		for _, match := range inventoryPattern.FindAllStringSubmatch(invSection, -1) {
			names = append(names, strings.TrimSpace(match[1]))
		}
	}

	// 2. Tools in workflow
	if wfStart := strings.Index(markdown, "## Workflow"); wfStart >= 0 {
		wfEnd := strings.Index(markdown[wfStart:], "##")
		if wfEnd < 0 {
			wfEnd = len(markdown) - wfStart
		} else {
			wfEnd += wfStart
		}
		wfSection := markdown[wfStart:wfEnd]
		names = append(names, extractWorkflowCallNames(wfSection)...)
	}

	// 3. Tools in output format
	if outStart := strings.Index(markdown, "## Output Format"); outStart >= 0 {
		outSection := markdown[outStart:]
		for _, match := range outputFormatToolPattern.FindAllStringSubmatch(outSection, -1) {
			name := strings.TrimSpace(match[1])
			if name != "" && strings.ToUpper(name) != "DONE" {
				names = append(names, name)
			}
		}
	}

	return names
}

func extractWorkflowCallNames(text string) []string {
	matches := callPattern.FindAllStringSubmatch(text, -1)
	names := make([]string, 0, len(matches))
	for _, match := range matches {
		name := strings.TrimSpace(match[1])
		if name == "" {
			name = strings.TrimSpace(match[2])
		}
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

func validateWorkflowProcess(markdown string, catalog tools.NormalizedToolCatalog) []string {
	var issues []string

	workflowMatch := workflowBlockPattern.FindStringSubmatch(markdown)
	if len(workflowMatch) < 2 {
		return nil
	}
	workflowBody := workflowMatch[1]
	workflowCalls := extractWorkflowCallNames(workflowBody)
	
	// Extract inventory names specifically
	inventoryNames := []string{}
	inventoryPattern := regexp.MustCompile("(?m)^-\\s+`([^`]+)`")
	if invStart := strings.Index(markdown, "## Tool Inventory"); invStart >= 0 {
		invEnd := strings.Index(markdown[invStart:], "##")
		if invEnd < 0 { invEnd = len(markdown) - invStart } else { invEnd += invStart }
		invSection := markdown[invStart:invEnd]
		for _, match := range inventoryPattern.FindAllStringSubmatch(invSection, -1) {
			inventoryNames = append(inventoryNames, strings.TrimSpace(match[1]))
		}
	}

	// Extract output format names specifically
	outputNames := []string{}
	if outStart := strings.Index(markdown, "## Output Format"); outStart >= 0 {
		outSection := markdown[outStart:]
		for _, match := range outputFormatToolPattern.FindAllStringSubmatch(outSection, -1) {
			name := strings.TrimSpace(match[1])
			if name != "" && strings.ToUpper(name) != "DONE" {
				outputNames = append(outputNames, name)
			}
		}
	}

	if len(inventoryNames) > 0 {
		for _, toolName := range inventoryNames {
			if !containsString(workflowCalls, toolName) {
				issues = append(issues, fmt.Sprintf("tool inventory includes %q but workflow does not call it", toolName))
			}
		}
		for _, toolName := range workflowCalls {
			if !containsString(inventoryNames, toolName) {
				issues = append(issues, fmt.Sprintf("workflow calls %q but tool inventory does not list it", toolName))
			}
		}
	}

	if len(outputNames) > 0 {
		for _, toolName := range workflowCalls {
			if !containsString(outputNames, toolName) {
				issues = append(issues, fmt.Sprintf("workflow calls %q but output format does not include it", toolName))
			}
		}
	}

	return issues
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func dedupeIssues(issues []string) []string {
	if len(issues) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	deduped := make([]string, 0, len(issues))
	for _, issue := range issues {
		issue = strings.TrimSpace(issue)
		if issue == "" {
			continue
		}
		if _, ok := seen[issue]; ok {
			continue
		}
		seen[issue] = struct{}{}
		deduped = append(deduped, issue)
	}
	return deduped
}

func formatMarkdownIssues(issues []string) string {
	if len(issues) == 0 {
		return "No validation issues were found. Return the markdown unchanged if it is already valid."
	}
	return "- " + strings.Join(issues, "\n- ")
}
