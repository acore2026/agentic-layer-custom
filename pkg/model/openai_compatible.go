package model

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"regexp"
	"strings"
	"time"

	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

type OpenAICompatibleLLM struct {
	name       string
	baseURL    string
	apiKey     string
	httpClient *http.Client
	thinking   *openAIThinkingConfig
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model       string                `json:"model"`
	Messages    []openAIMessage       `json:"messages"`
	Tools       []openAITool          `json:"tools,omitempty"`
	ToolChoice  string                `json:"tool_choice,omitempty"`
	Temperature *float32              `json:"temperature,omitempty"`
	MaxTokens   *int32                `json:"max_tokens,omitempty"`
	Thinking    *openAIThinkingConfig `json:"thinking,omitempty"`
	Stream      bool                  `json:"stream"`
}

type openAIThinkingConfig struct {
	Type string `json:"type"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content          string           `json:"content"`
			ReasoningContent string           `json:"reasoning_content"`
			ToolCalls        []openAIToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Model string `json:"model"`
}

type openAIStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content          string                 `json:"content"`
			ReasoningContent string                 `json:"reasoning_content"`
			ToolCalls        []openAIStreamToolCall `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Model string `json:"model"`
}

type openAITool struct {
	Type     string             `json:"type"`
	Function openAIFunctionSpec `json:"function"`
}

type openAIFunctionSpec struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters,omitempty"`
}

type openAIToolCall struct {
	ID       string             `json:"id,omitempty"`
	Type     string             `json:"type,omitempty"`
	Function openAIFunctionCall `json:"function"`
}

type openAIStreamToolCall struct {
	Index    int                      `json:"index"`
	ID       string                   `json:"id,omitempty"`
	Type     string                   `json:"type,omitempty"`
	Function openAIStreamFunctionCall `json:"function"`
}

type openAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIStreamFunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

var functionCallMarkupPattern = regexp.MustCompile(`(?s)<function_calls>\s*(.*?)\s*</function_calls>`)
var invokePattern = regexp.MustCompile(`(?s)<invoke\s+name="([^"]+)">\s*(.*?)\s*</invoke>`)
var parameterPattern = regexp.MustCompile(`(?s)<parameter\s+name="([^"]+)">\s*(.*?)\s*</parameter>`)

func NewOpenAICompatibleLLM(name, baseURL, apiKey string) *OpenAICompatibleLLM {
	return &OpenAICompatibleLLM{
		name:       name,
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
	}
}

func (m *OpenAICompatibleLLM) Name() string {
	return m.name
}

func (m *OpenAICompatibleLLM) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		if !stream {
			response, err := m.generate(ctx, req, false)
			yield(response, err)
			return
		}

		if err := m.generateStream(ctx, req, yield); err != nil {
			yield(nil, err)
		}
	}
}

func (m *OpenAICompatibleLLM) generate(ctx context.Context, req *model.LLMRequest, stream bool) (*model.LLMResponse, error) {
	httpResponse, err := m.doChatCompletionRequest(ctx, req, stream)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	var decoded openAIResponse
	if err := json.NewDecoder(httpResponse.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode openai-compatible response: %w", err)
	}

	if httpResponse.StatusCode >= 400 {
		return nil, fmt.Errorf("openai-compatible model returned status %d", httpResponse.StatusCode)
	}
	if len(decoded.Choices) == 0 {
		return nil, fmt.Errorf("openai-compatible model returned no choices")
	}

	content := buildContentFromOpenAI(
		strings.TrimSpace(decoded.Choices[0].Message.Content),
		strings.TrimSpace(decoded.Choices[0].Message.ReasoningContent),
		decoded.Choices[0].Message.ToolCalls,
	)
	if content == nil || len(content.Parts) == 0 {
		return nil, fmt.Errorf("openai-compatible model returned empty content")
	}

	return &model.LLMResponse{
		Content:      content,
		ModelVersion: decoded.Model,
		Partial:      false,
		TurnComplete: true,
	}, nil
}

func (m *OpenAICompatibleLLM) generateStream(
	ctx context.Context,
	req *model.LLMRequest,
	yield func(*model.LLMResponse, error) bool,
) error {
	httpResponse, err := m.doChatCompletionRequest(ctx, req, true)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(httpResponse.Body)
		return fmt.Errorf("openai-compatible model returned status %d: %s", httpResponse.StatusCode, strings.TrimSpace(string(errorBody)))
	}

	scanner := bufio.NewScanner(httpResponse.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var builder strings.Builder
	var reasoningBuilder strings.Builder
	modelVersion := m.name
	toolCalls := map[int]*openAIToolCall{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}

		payloadLine := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payloadLine == "[DONE]" {
			break
		}

		var chunk openAIStreamResponse
		if err := json.Unmarshal([]byte(payloadLine), &chunk); err != nil {
			return fmt.Errorf("decode openai-compatible stream chunk: %w", err)
		}
		if strings.TrimSpace(chunk.Model) != "" {
			modelVersion = chunk.Model
		}

		for _, choice := range chunk.Choices {
			reasoningDelta := choice.Delta.ReasoningContent
			if reasoningDelta != "" {
				reasoningBuilder.WriteString(reasoningDelta)
				if !yield(&model.LLMResponse{
					Content:      &genai.Content{Role: genai.RoleModel, Parts: []*genai.Part{{Text: reasoningDelta, Thought: true}}},
					ModelVersion: modelVersion,
					Partial:      true,
					TurnComplete: false,
				}, nil) {
					return nil
				}
			}
			delta := choice.Delta.Content
			if delta != "" {
				builder.WriteString(delta)
				if !yield(&model.LLMResponse{
					Content:      genai.NewContentFromText(delta, genai.RoleModel),
					ModelVersion: modelVersion,
					Partial:      true,
					TurnComplete: false,
				}, nil) {
					return nil
				}
			}
			for _, toolDelta := range choice.Delta.ToolCalls {
				call := toolCalls[toolDelta.Index]
				if call == nil {
					call = &openAIToolCall{}
					toolCalls[toolDelta.Index] = call
				}
				if strings.TrimSpace(toolDelta.ID) != "" {
					call.ID = toolDelta.ID
				}
				if strings.TrimSpace(toolDelta.Type) != "" {
					call.Type = toolDelta.Type
				}
				if strings.TrimSpace(toolDelta.Function.Name) != "" {
					call.Function.Name = toolDelta.Function.Name
				}
				if toolDelta.Function.Arguments != "" {
					call.Function.Arguments += toolDelta.Function.Arguments
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read openai-compatible stream: %w", err)
	}

	finalText := strings.TrimSpace(builder.String())
	finalReasoning := strings.TrimSpace(reasoningBuilder.String())
	finalToolCalls := orderedToolCalls(toolCalls)
	content := buildContentFromOpenAI(finalText, finalReasoning, finalToolCalls)
	if content == nil || len(content.Parts) == 0 {
		return fmt.Errorf("openai-compatible model returned empty content")
	}

	if !yield(&model.LLMResponse{
		Content:      content,
		ModelVersion: modelVersion,
		Partial:      false,
		TurnComplete: true,
	}, nil) {
		return nil
	}

	return nil
}

func (m *OpenAICompatibleLLM) doChatCompletionRequest(ctx context.Context, req *model.LLMRequest, stream bool) (*http.Response, error) {
	chatMessages := make([]openAIMessage, 0, len(req.Contents)+1)
	if req.Config != nil && req.Config.SystemInstruction != nil {
		chatMessages = append(chatMessages, openAIMessage{
			Role:    "system",
			Content: contentText(req.Config.SystemInstruction),
		})
	}

	for _, content := range req.Contents {
		if content == nil {
			continue
		}
		chatMessages = append(chatMessages, openAIMessage{
			Role:    mapOpenAIRole(string(content.Role)),
			Content: contentText(content),
		})
	}

	payload := openAIRequest{
		Model:    m.name,
		Messages: chatMessages,
		Stream:   stream,
	}
	if m.thinking != nil {
		payload.Thinking = m.thinking
	}
	if req.Config != nil && len(req.Config.Tools) > 0 {
		payload.Tools = toOpenAITools(req.Config.Tools)
		if len(payload.Tools) > 0 {
			payload.ToolChoice = "auto"
		}
	}
	if req.Config != nil && req.Config.Temperature != nil {
		temp := float32(*req.Config.Temperature)
		payload.Temperature = &temp
	}
	if req.Config != nil && req.Config.MaxOutputTokens > 0 {
		maxTokens := req.Config.MaxOutputTokens
		payload.MaxTokens = &maxTokens
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal openai-compatible request: %w", err)
	}

	backoffs := []time.Duration{0, 2 * time.Second, 5 * time.Second}
	var lastErr error
	for attempt, backoff := range backoffs {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, m.baseURL+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("build openai-compatible request: %w", err)
		}
		httpRequest.Header.Set("Content-Type", "application/json")
		httpRequest.Header.Set("Authorization", "Bearer "+m.apiKey)

		httpResponse, err := m.httpClient.Do(httpRequest)
		if err != nil {
			lastErr = fmt.Errorf("call openai-compatible model: %w", err)
			continue
		}

		if httpResponse.StatusCode == http.StatusTooManyRequests || httpResponse.StatusCode >= 500 {
			errorBody, _ := io.ReadAll(httpResponse.Body)
			_ = httpResponse.Body.Close()
			lastErr = fmt.Errorf("openai-compatible model returned status %d: %s", httpResponse.StatusCode, strings.TrimSpace(string(errorBody)))
			continue
		}

		return httpResponse, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("openai-compatible request failed without a response")
}

func (m *OpenAICompatibleLLM) WithThinkingEnabled(enabled bool) *OpenAICompatibleLLM {
	clone := *m
	if enabled {
		clone.thinking = nil
	} else {
		clone.thinking = &openAIThinkingConfig{Type: "disabled"}
	}
	return &clone
}

func mapOpenAIRole(role string) string {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "", "user":
		return "user"
	case "model", "assistant":
		return "assistant"
	case "tool", "function":
		return "tool"
	case "system":
		return "system"
	default:
		return "user"
	}
}

func toOpenAITools(tools []*genai.Tool) []openAITool {
	var converted []openAITool
	for _, tool := range tools {
		if tool == nil {
			continue
		}
		for _, declaration := range tool.FunctionDeclarations {
			if declaration == nil || strings.TrimSpace(declaration.Name) == "" {
				continue
			}
			parameters := declaration.ParametersJsonSchema
			if parameters == nil && declaration.Parameters != nil {
				parameters = declaration.Parameters
			}
			converted = append(converted, openAITool{
				Type: "function",
				Function: openAIFunctionSpec{
					Name:        declaration.Name,
					Description: declaration.Description,
					Parameters:  parameters,
				},
			})
		}
	}
	return converted
}

func buildContentFromOpenAI(text string, reasoning string, toolCalls []openAIToolCall) *genai.Content {
	cleanText := strings.TrimSpace(text)
	resolvedCalls := toolCalls
	if len(resolvedCalls) == 0 {
		cleanText, resolvedCalls = extractToolCallsFromMarkup(cleanText)
	}

	parts := make([]*genai.Part, 0, 2+len(resolvedCalls))
	if strings.TrimSpace(reasoning) != "" {
		parts = append(parts, &genai.Part{
			Text:    strings.TrimSpace(reasoning),
			Thought: true,
		})
	}
	if cleanText != "" {
		parts = append(parts, genai.NewPartFromText(cleanText))
	}

	for _, call := range resolvedCalls {
		args := map[string]any{}
		if strings.TrimSpace(call.Function.Arguments) != "" {
			_ = json.Unmarshal([]byte(call.Function.Arguments), &args)
		}
		part := genai.NewPartFromFunctionCall(call.Function.Name, args)
		if part.FunctionCall != nil {
			part.FunctionCall.ID = call.ID
		}
		parts = append(parts, part)
	}

	if len(parts) == 0 {
		return nil
	}
	return &genai.Content{
		Role:  genai.RoleModel,
		Parts: parts,
	}
}

func extractToolCallsFromMarkup(text string) (string, []openAIToolCall) {
	matches := functionCallMarkupPattern.FindStringSubmatch(text)
	if len(matches) < 2 {
		return text, nil
	}

	var toolCalls []openAIToolCall
	for _, invokeMatch := range invokePattern.FindAllStringSubmatch(matches[1], -1) {
		if len(invokeMatch) < 3 {
			continue
		}
		args := map[string]any{}
		for _, parameterMatch := range parameterPattern.FindAllStringSubmatch(invokeMatch[2], -1) {
			if len(parameterMatch) < 3 {
				continue
			}
			args[parameterMatch[1]] = strings.TrimSpace(parameterMatch[2])
		}
		arguments, _ := json.Marshal(args)
		toolCalls = append(toolCalls, openAIToolCall{
			ID:   "",
			Type: "function",
			Function: openAIFunctionCall{
				Name:      strings.TrimSpace(invokeMatch[1]),
				Arguments: string(arguments),
			},
		})
	}

	cleanText := strings.TrimSpace(functionCallMarkupPattern.ReplaceAllString(text, ""))
	return cleanText, toolCalls
}

func orderedToolCalls(source map[int]*openAIToolCall) []openAIToolCall {
	if len(source) == 0 {
		return nil
	}
	maxIndex := -1
	for index := range source {
		if index > maxIndex {
			maxIndex = index
		}
	}
	ordered := make([]openAIToolCall, 0, len(source))
	for i := 0; i <= maxIndex; i++ {
		call := source[i]
		if call == nil || strings.TrimSpace(call.Function.Name) == "" {
			continue
		}
		ordered = append(ordered, *call)
	}
	return ordered
}

func contentText(content *genai.Content) string {
	if content == nil {
		return ""
	}

	var parts []string
	for _, part := range content.Parts {
		if part == nil || strings.TrimSpace(part.Text) == "" {
			continue
		}
		parts = append(parts, part.Text)
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}
