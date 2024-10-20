// internal/llm/factory.go

package llm

import (
	"fmt"

	"github.com/chrlesur/Ontology/internal/config"
)

func GetClient(llmType string, model string) (Client, error) {
	cfg := config.GetConfig()

	switch llmType {
	case "openai":
		return NewOpenAIClient(cfg.OpenAIAPIKey, model)
	case "claude":
		return NewClaudeClient(cfg.ClaudeAPIKey, model)
	case "ollama":
		return NewOllamaClient(model)
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidLLMType, llmType)
	}
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
