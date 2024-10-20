// internal/llm/openai.go

package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/sashabaranov/go-openai"
)

// OpenAIClient implements the Client interface for OpenAI
type OpenAIClient struct {
	client *openai.Client
	model  string
}

// supportedModels defines the list of supported OpenAI models
var supportedModels = map[string]bool{
	"GPT-4o":      true,
	"GPT-4o mini": true,
	"o1-preview":  true,
	"o1-mini":     true,
}


// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey string, model string) (*OpenAIClient, error) {
	if apiKey == "" {
		return nil, ErrAPIKeyMissing
	}

	if !supportedModels[model] {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedModel, model)
	}

	client := openai.NewClient(apiKey)
	return &OpenAIClient{
		client: client,
		model:  model,
	}, nil
}

// Translate sends a prompt to the OpenAI API and returns the response
func (c *OpenAIClient) Translate(prompt string, context string) (string, error) {
	log.Debug(i18n.TranslationStarted, "OpenAI", c.model)

	var result string
	var err error
	for attempt := 1; attempt <= 5; attempt++ {
		result, err = c.makeRequest(prompt, context)
		if err == nil {
			break
		}
		log.Warning(i18n.TranslationRetry, attempt, err)
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTranslationFailed, err)
	}

	log.Info(i18n.TranslationCompleted, "OpenAI", c.model)
	return result, nil
}

func (c *OpenAIClient) makeRequest(prompt string, systemContext string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemContext,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	ctx := context.Background()
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("error creating chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return resp.Choices[0].Message.Content, nil
}
