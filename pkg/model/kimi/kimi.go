package kimi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"

	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"
)

type KimiModel struct {
	apiKey    string
	modelName string
	baseURL   string
}

func NewKimiModel(apiKey, modelName string) *KimiModel {
	return &KimiModel{
		apiKey:    apiKey,
		modelName: modelName,
		baseURL:   "https://api.moonshot.cn/v1/chat/completions",
	}
}

func NewModel(apiKey, modelName string) *KimiModel {
	return NewKimiModel(apiKey, modelName)
}

func (k *KimiModel) Name() string {
	return k.modelName
}

type openAIMessage struct {
	Role             string           `json:"role"`
	Content          string           `json:"content"`
	ReasoningContent string           `json:"reasoning_content,omitempty"`
	ToolCalls        []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID       string           `json:"tool_call_id,omitempty"`
}

type openAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openAITool struct {
	Type     string `json:"type"`
	Function any    `json:"function"`
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
	Tools    []openAITool    `json:"tools,omitempty"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
}

func (k *KimiModel) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		var messages []openAIMessage

		// Add system instruction if present in Config
		if req.Config != nil && req.Config.SystemInstruction != nil {
			var systemText string
			for _, p := range req.Config.SystemInstruction.Parts {
				systemText += p.Text
			}
			if systemText != "" {
				messages = append(messages, openAIMessage{
					Role:    "system",
					Content: systemText,
				})
			}
		}

		for _, content := range req.Contents {
			role := content.Role
			if role == "" || role == "user" {
				role = "user"
			} else if role == "model" {
				role = "assistant"
			}

			var text string
			var reasoningContent string
			var toolCalls []openAIToolCall
			var toolCallID string

			for _, part := range content.Parts {
				if part.Thought {
					reasoningContent = part.Text
					continue
				}
				if part.Text != "" {
					text += part.Text
				}
				if part.FunctionCall != nil {
					args, _ := json.Marshal(part.FunctionCall.Args)
					toolCalls = append(toolCalls, openAIToolCall{
						ID:   "call_" + part.FunctionCall.Name,
						Type: "function",
						Function: struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						}{Name: part.FunctionCall.Name, Arguments: string(args)},
					})
				}
				if part.FunctionResponse != nil {
					toolCallID = "call_" + part.FunctionResponse.Name
					respArgs, _ := json.Marshal(part.FunctionResponse.Response)
					text = string(respArgs)
					role = "tool"
				}
			}

			messages = append(messages, openAIMessage{
				Role:             role,
				Content:          text,
				ReasoningContent: reasoningContent,
				ToolCalls:        toolCalls,
				ToolCallID:       toolCallID,
			})
		}

		// RESILIENCE FIX: If adk-go didn't provide any messages, we send a "ping" or return a dummy response
		// This prevents the error from crashing the whole flow.
		if len(messages) == 0 {
			// If it's an empty request, we'll return a minimal successful "ready" response
			// so that adk-go can decide whether to try again or finish.
			llmResp := &model.LLMResponse{
				Content: &genai.Content{
					Role:  "model",
					Parts: []*genai.Part{{Text: "Ready."}},
				},
			}
			yield(llmResp, nil)
			return
		}

		oaReq := openAIRequest{
			Model:    k.modelName,
			Messages: messages,
		}

		if len(req.Tools) > 0 {
			for _, item := range req.Tools {
				t, ok := item.(tool.Tool)
				if !ok {
					continue
				}
				oaReq.Tools = append(oaReq.Tools, openAITool{
					Type: "function",
					Function: struct {
						Name       string         `json:"name"`
						Description string         `json:"description"`
						Parameters map[string]any `json:"parameters"`
					}{
						Name:        t.Name(),
						Description: t.Description(),
						Parameters: map[string]any{
							"type":                 "object",
							"properties":           map[string]any{}, // Generic map[string]any has no fixed properties
							"additionalProperties": true,
						},
					},
				})
			}
		}

		body, err := json.Marshal(oaReq)
		if err != nil {
			yield(nil, err)
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", k.baseURL, bytes.NewBuffer(body))
		if err != nil {
			yield(nil, err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+k.apiKey)

		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			yield(nil, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errResp map[string]any
			json.NewDecoder(resp.Body).Decode(&errResp)
			yield(nil, fmt.Errorf("Kimi API error (status %d): %v", resp.StatusCode, errResp))
			return
		}

		var oaResp openAIResponse
		if err := json.NewDecoder(resp.Body).Decode(&oaResp); err != nil {
			yield(nil, err)
			return
		}

		if len(oaResp.Choices) == 0 {
			yield(nil, fmt.Errorf("Kimi API returned no choices"))
			return
		}

		msg := oaResp.Choices[0].Message
		llmResp := &model.LLMResponse{
			Content: &genai.Content{
				Role: "model",
			},
		}

		if msg.ReasoningContent != "" {
			llmResp.Content.Parts = append(llmResp.Content.Parts, &genai.Part{
				Text:    msg.ReasoningContent,
				Thought: true,
			})
		}

		if msg.Content != "" {
			llmResp.Content.Parts = append(llmResp.Content.Parts, &genai.Part{Text: msg.Content})
		}

		for _, tc := range msg.ToolCalls {
			var args map[string]any
			json.Unmarshal([]byte(tc.Function.Arguments), &args)
			llmResp.Content.Parts = append(llmResp.Content.Parts, &genai.Part{
				FunctionCall: &genai.FunctionCall{
					Name: tc.Function.Name,
					Args: args,
				},
			})
		}

		yield(llmResp, nil)
	}
}
