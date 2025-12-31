package llm

import (
	"context"
	"fmt"

	"investor/config"

	"github.com/go-resty/resty/v2"
)

type OpenAIProvider struct {
	client *resty.Client
	config config.LLMConfig
}

type openAIRequest struct {
	Model    string                   `json:"model"`
	Messages []Message                `json:"messages"`
	Tools    []map[string]interface{} `json:"tools,omitempty"`
}

type openAIResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func NewOpenAIProvider(cfg config.LLMConfig) *OpenAIProvider {
	return &OpenAIProvider{
		client: resty.New(),
		config: cfg,
	}
}

func (p *OpenAIProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	respMsg, err := p.ChatWithTools(ctx, messages, nil)
	if err != nil {
		return "", err
	}
	return respMsg.Content, nil
}

func (p *OpenAIProvider) ChatWithTools(ctx context.Context, messages []Message, tools []map[string]interface{}) (*Message, error) {
	// 1. Determine Model Name
	model := p.config.ModelName
	if model == "" {
		// Fallback defaults if not configured
		if p.config.Provider == "deepseek" {
			model = "deepseek-chat"
		} else {
			model = "gpt-3.5-turbo"
		}
	}

	reqBody := openAIRequest{
		Model:    model,
		Messages: messages,
		Tools:    tools,
	}

	var respBody openAIResponse

	resp, err := p.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+p.config.APIKey).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		SetResult(&respBody).
		Post(p.config.APIURL + "/chat/completions")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("LLM API error: %s", resp.String())
	}

	if len(respBody.Choices) == 0 {
		return nil, fmt.Errorf("empty response from LLM")
	}

	return &respBody.Choices[0].Message, nil
}
