// internal/llm/factory.go

package llm

import (
	"fmt"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/i18n"

)

func GetClient(llmType string, model string) (Client, error) {
	cfg := config.GetConfig()

	var client Client
	var err error

	switch llmType {
	case "openai":
		client, err = NewOpenAIClient(cfg.OpenAIAPIKey, model)
	case "claude":
		client, err = NewClaudeClient(cfg.ClaudeAPIKey, model)
	case "ollama":
		client, err = NewOllamaClient(model)
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidLLMType, llmType)
	}

	if err != nil {
		return nil, err
	}

	return &contextCheckingClient{
		baseClient: client,
		model:      model,
	}, nil
}

type contextCheckingClient struct {
	baseClient Client
	model      string
}

func (c *contextCheckingClient) Translate(prompt string, context string) (string, error) {
	if err := CheckContextLength(c.model, context); err != nil {
		return "", err
	}
	return c.baseClient.Translate(prompt, context)
}
