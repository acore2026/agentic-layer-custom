package kimi_test

import (
	"context"
	"testing"

	"agentic-layer-custom/pkg/model/kimi"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

func TestKimiProvider(t *testing.T) {
	ctx := context.Background()
	k := kimi.NewModel("mock-api-key", "moonshot-v1-8k")

	if k.Name() != "moonshot-v1-8k" {
		t.Errorf("Expected model name 'moonshot-v1-8k', got %q", k.Name())
	}

	t.Run("GenerateContent should yield once even on error (integration test required for full check)", func(t *testing.T) {
		req := &model.LLMRequest{
			Contents: []*genai.Content{
				{
					Role:  "user",
					Parts: []*genai.Part{{Text: "Hello"}},
				},
			},
		}

		// Since we don't have a real API key or network in unit tests easily, we verify it attempts to call
		// In a real environment, we'd mock the HTTP client but for this PoC we verify it doesn't panic and is typed correctly.
		
		// This should fail due to invalid API key, but we want to check it yields.
		yielded := false
		for _, err := range k.GenerateContent(ctx, req, false) {
			yielded = true
			if err == nil {
				t.Error("Expected error with mock-api-key, got nil")
			}
		}

		if !yielded {
			t.Error("Expected at least one yield from GenerateContent")
		}
	})
}
