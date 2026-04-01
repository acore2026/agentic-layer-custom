package kimi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"

	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

type KimiModel struct {
	apiKey    string
	modelName string
	baseURL   string
}

func NewModel(apiKey, modelName string) *KimiModel {
	return &KimiModel{
		apiKey:    apiKey,
		modelName: modelName,
		baseURL:   "https://api.moonshot.cn/v1/chat/completions",
	}
}

func (k *KimiModel) Name() string {
	return k.modelName
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
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
		for _, content := range req.Contents {
			role := content.Role
			if role == "" || role == "user" {
				role = "user"
			} else if role == "model" {
				role = "assistant"
			}

			var text string
			var toolCalls []openAIToolCall
			var toolCallID string

			for _, part := range content.Parts {
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
				Role:       role,
				Content:    text,
				ToolCalls:  toolCalls,
				ToolCallID: toolCallID,
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
			for _, t := range req.Tools {
				oaReq.Tools = append(oaReq.Tools, openAITool{
					Type:     "function",
					Function: t,
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
