package agents

import (
	"strings"
	"testing"
)

func TestLoadSkills(t *testing.T) {
	// The skills are in the project root's skill/ directory
	// In the test context, we may need to adjust the path depending on where go test is run
	// However, we can test with the actual directory structure since it's a fixed part of the project
	content, err := LoadSkills("../../skill")
	if err != nil {
		t.Fatalf("Failed to load skills: %v", err)
	}

	if content == "" {
		t.Fatal("Expected non-empty skill content")
	}

	expectedSkills := []string{"acn", "init-registration", "pdu-session-establishment"}
	for _, skill := range expectedSkills {
		if !strings.Contains(content, skill) {
			t.Errorf("Expected skill content to contain %q", skill)
		}
	}
}

func TestDiscoverTools(t *testing.T) {
	mockContent := `
    # Step 1: Subscription Status Check
    CALL "Subscription_tool"

    # Step 2: Create/Update Subnet Context
    CALL "Create_Or_Update_Subnet_Context_tool"

    # Step 3: Issue Access Token
    CALL "Issue_Access_Token_tool"
	`

	tools := DiscoverTools(mockContent)
	expectedTools := []string{"Subscription_tool", "Create_Or_Update_Subnet_Context_tool", "Issue_Access_Token_tool"}

	if len(tools) != len(expectedTools) {
		t.Fatalf("Expected %d tools, got %d: %v", len(expectedTools), len(tools), tools)
	}

	for i, tool := range tools {
		if tool != expectedTools[i] {
			t.Errorf("Expected tool at index %d to be %q, got %q", i, expectedTools[i], tool)
		}
	}
}

func TestDiscoverToolsDeDuplication(t *testing.T) {
	mockContent := `
    CALL "Auth_tool"
    CALL "Subscription_tool"
    CALL "Auth_tool"
	`

	tools := DiscoverTools(mockContent)
	if len(tools) != 2 {
		t.Fatalf("Expected 2 unique tools, got %d: %v", len(tools), tools)
	}

	if tools[0] != "Auth_tool" || tools[1] != "Subscription_tool" {
		t.Errorf("Expected [Auth_tool, Subscription_tool], got %v", tools)
	}
}
