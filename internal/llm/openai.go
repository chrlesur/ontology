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

type OpenAIClient struct {
	apiKey string
	model  string
	client *openai.Client
	config *config.Config
}

var supportedModels = map[string]bool{
	"gpt-4o":      true,
	"gpt-4o-mini": true,
	"o1-preview":  true,
	"o1-mini":     true,
}

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
		apiKey: apiKey,
		model:  model,
		client: client,
		config: config.GetConfig(),
	}, nil
}

func (c *OpenAIClient) Translate(prompt string, context string) (string, error) {
	log.Debug(i18n.Messages.TranslationStarted, "OpenAI", c.model)
	log.Debug("Starting Translate. Prompt length: %d, Context length: %d", len(prompt), len(context))

	var result string
	var err error
	maxRetries := 5
	baseDelay := time.Second * 10
	maxDelay := time.Minute * 2

	for attempt := 0; attempt < maxRetries; attempt++ {
		log.Debug("Attempt %d of %d", attempt+1, maxRetries)
		result, err = c.makeRequest(prompt, context)
		if err == nil {
			log.Debug(i18n.Messages.TranslationCompleted, "OpenAI", c.model)
			return result, nil
		}

		if !isRateLimitError(err) {
			log.Warning(i18n.Messages.TranslationRetry, attempt+1, err)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		delay := baseDelay * time.Duration(1<<uint(attempt))
		if delay > maxDelay {
			delay = maxDelay
		}
		log.Warning(i18n.Messages.RateLimitExceeded, delay)
		time.Sleep(delay)
	}

	log.Error(i18n.Messages.TranslationFailed, err)
	log.Debug("Translation completed. Result length: %d", len(result))

	return "", fmt.Errorf("%w: %v", ErrTranslationFailed, err)
}

func (c *OpenAIClient) makeRequest(prompt string, systemContext string) (string, error) {
	log.Debug("Making request to OpenAI API")

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemContext,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       c.model,
			Messages:    messages,
			MaxTokens:   c.config.MaxTokens,
			Temperature: 0.7,
		},
	)

	if err != nil {
		log.Error("Error creating chat completion: %v", err)
		return "", fmt.Errorf("error creating chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		log.Error("No content in response")
		return "", fmt.Errorf("no content in response")
	}

	log.Debug("Successfully received and parsed response from OpenAI API.")
	return resp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) ProcessWithPrompt(promptTemplate *prompt.PromptTemplate, values map[string]string) (string, error) {
	log.Debug("Processing prompt with OpenAI")
	formattedPrompt := promptTemplate.Format(values)

	return c.Translate(formattedPrompt, "")
}
