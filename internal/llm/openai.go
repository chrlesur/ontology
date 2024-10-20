// internal/llm/openai.go

package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/prompt"
	"github.com/sashabaranov/go-openai"
)

// OpenAIClient implements the Client interface for OpenAI
type OpenAIClient struct {
	client *openai.Client
	model  string
	config *config.Config
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
	log.Debug("Creating new OpenAI client with model: %s", model)
	if apiKey == "" {
		log.Error("API key is missing for OpenAI client")
		return nil, ErrAPIKeyMissing
	}

	if !supportedModels[model] {
		log.Error("Unsupported OpenAI model: %s", model)
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedModel, model)
	}

	client := openai.NewClient(apiKey)
	return &OpenAIClient{
		client: client,
		model:  model,
		config: config.GetConfig(),
	}, nil
}

// Translate sends a prompt to the OpenAI API and returns the response
func (c *OpenAIClient) Translate(prompt string, context string) (string, error) {
	log.Debug(i18n.Messages.TranslationStarted, "OpenAI", c.model)
	log.Debug("Prompt length: %d, Context length: %d", len(prompt), len(context))

	var result string
	var err error
	maxRetries := 5
	baseDelay := time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Debug("Attempt %d of %d", attempt, maxRetries)
		result, err = c.makeRequest(prompt, context)
		if err == nil {
			log.Info(i18n.Messages.TranslationCompleted, "OpenAI", c.model)
			log.Debug("Translation successful, result length: %d", len(result))
			return result, nil
		}

		log.Warning(i18n.Messages.TranslationRetry, attempt, err)
		time.Sleep(time.Duration(attempt) * baseDelay)
	}

	log.Error(i18n.Messages.TranslationFailed, err)
	return "", fmt.Errorf("%w: %v", ErrTranslationFailed, err)
}

func (c *OpenAIClient) makeRequest(prompt string, systemContext string) (string, error) {
	log.Debug("Making request to OpenAI API")

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
		MaxTokens: c.config.MaxTokens,
	}

	ctx := context.Background()
	log.Debug("Sending request to OpenAI API")
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Error("Error creating chat completion: %v", err)
		return "", fmt.Errorf("error creating chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		log.Error("No choices in response")
		return "", fmt.Errorf("no choices in response")
	}

	log.Debug("Successfully received and parsed response from OpenAI API")
	return resp.Choices[0].Message.Content, nil
}

// ProcessWithPrompt processes a prompt template with the given values and sends it to the OpenAI API
func (c *OpenAIClient) ProcessWithPrompt(promptTemplate *prompt.PromptTemplate, values map[string]string) (string, error) {
	log.Debug("Processing prompt with OpenAI")
	formattedPrompt := promptTemplate.Format(values)
	log.Debug("Formatted prompt: %s", formattedPrompt)

	// Utilisez la méthode Translate existante pour envoyer le prompt formatté
	return c.Translate(formattedPrompt, "")
}
