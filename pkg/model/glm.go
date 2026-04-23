package model

import (
	"fmt"
	"os"
	"strings"
)

const (
	ProviderGLM5      = "glm5"
	DefaultGLMBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	DefaultGLMModel   = "glm-5"
)

func NormalizeOpenAIBaseURL(raw string) string {
	normalized := strings.TrimSpace(raw)
	if normalized == "" {
		return DefaultGLMBaseURL
	}
	normalized = strings.TrimRight(normalized, "/")
	if strings.HasSuffix(normalized, "/chat/completions") {
		return strings.TrimSuffix(normalized, "/chat/completions")
	}
	return normalized
}

func NewGLM5LLM(apiKey, baseURL, modelName string) *OpenAICompatibleLLM {
	if strings.TrimSpace(modelName) == "" {
		modelName = DefaultGLMModel
	}

	llm := NewOpenAICompatibleLLM(modelName, NormalizeOpenAIBaseURL(baseURL), apiKey)
	return llm.WithThinkingEnabled(false)
}

func NewGLM5LLMFromEnv() (*OpenAICompatibleLLM, error) {
	apiKey := strings.TrimSpace(os.Getenv("GLM_API_KEY"))
	if apiKey == "" {
		return nil, fmt.Errorf("GLM_API_KEY must be set when using %s provider", ProviderGLM5)
	}

	baseURL := os.Getenv("GLM_BASE_URL")
	modelName := os.Getenv("GLM_MODEL")
	return NewGLM5LLM(apiKey, baseURL, modelName), nil
}
