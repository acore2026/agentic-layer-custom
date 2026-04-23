package model

import "testing"

func TestNormalizeOpenAIBaseURL(t *testing.T) {
	t.Run("defaults when empty", func(t *testing.T) {
		if got := NormalizeOpenAIBaseURL(""); got != DefaultGLMBaseURL {
			t.Fatalf("NormalizeOpenAIBaseURL() = %q", got)
		}
	})

	t.Run("strips chat completions suffix", func(t *testing.T) {
		raw := "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
		if got := NormalizeOpenAIBaseURL(raw); got != DefaultGLMBaseURL {
			t.Fatalf("NormalizeOpenAIBaseURL(%q) = %q", raw, got)
		}
	})

	t.Run("keeps base path", func(t *testing.T) {
		raw := "https://example.com/v1/"
		if got := NormalizeOpenAIBaseURL(raw); got != "https://example.com/v1" {
			t.Fatalf("NormalizeOpenAIBaseURL(%q) = %q", raw, got)
		}
	})
}

func TestNewGLM5LLMFromEnv(t *testing.T) {
	t.Run("requires API key", func(t *testing.T) {
		t.Setenv("GLM_API_KEY", "")
		if _, err := NewGLM5LLMFromEnv(); err == nil {
			t.Fatal("expected error when GLM_API_KEY is missing")
		}
	})

	t.Run("applies defaults and disables thinking", func(t *testing.T) {
		t.Setenv("GLM_API_KEY", "test-key")
		t.Setenv("GLM_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions")
		t.Setenv("GLM_MODEL", "")

		llm, err := NewGLM5LLMFromEnv()
		if err != nil {
			t.Fatalf("NewGLM5LLMFromEnv() error = %v", err)
		}
		if got := llm.Name(); got != DefaultGLMModel {
			t.Fatalf("Name() = %q", got)
		}
		if got := llm.baseURL; got != DefaultGLMBaseURL {
			t.Fatalf("baseURL = %q", got)
		}
		if llm.thinking == nil || llm.thinking.Type != "disabled" {
			t.Fatalf("thinking = %#v", llm.thinking)
		}
	})
}
