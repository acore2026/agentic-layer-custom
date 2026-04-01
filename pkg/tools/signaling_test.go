package tools

import (
	"testing"
)

func TestUniversalMockTool(t *testing.T) {
	args := map[string]any{
		"ue_id": "UE-01",
		"other": "param",
	}

	res, err := UniversalMockTool(nil, "Auth_tool", args)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultMap, ok := res.(map[string]any)
	if !ok {
		t.Fatalf("Expected map[string]any, got %T", res)
	}

	if resultMap["status"] != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %v", resultMap["status"])
	}

	if resultMap["ue_id"] != "UE-01" {
		t.Errorf("Expected ue_id UE-01, got %v", resultMap["ue_id"])
	}

	if resultMap["token"] == "" {
		t.Error("Expected non-empty token")
	}
}

func TestUniversalMockToolMissingUEID(t *testing.T) {
	args := map[string]any{
		"other": "param",
	}

	res, err := UniversalMockTool(nil, "Subscription_tool", args)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultMap, ok := res.(map[string]any)
	if !ok {
		t.Fatalf("Expected map[string]any, got %T", res)
	}

	if resultMap["ue_id"] != "UNKNOWN" {
		t.Errorf("Expected ue_id UNKNOWN, got %v", resultMap["ue_id"])
	}
}
