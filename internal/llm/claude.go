// internal/llm/claude.go

package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/prompt"
)

// ClaudeClient implements the Client interface for Claude
type ClaudeClient struct {
	apiKey string
	model  string
	client *http.Client
	config *config.Config
}

// supportedClaudeModels defines the list of supported Claude models
var supportedClaudeModels = map[string]bool{
	"claude-3-5-sonnet-20240620": true,
	"claude-3-opus-20240229":     true,
	"claude-3-haiku-20240307":    true,
}

// NewClaudeClient creates a new Claude client
func NewClaudeClient(apiKey string, model string) (*ClaudeClient, error) {
	log.Debug("Creating new Claude client with model: %s", model)
	if apiKey == "" {
		log.Error("API key is missing for Claude client")
		return nil, ErrAPIKeyMissing
	}

	if !supportedClaudeModels[model] {
		log.Error("Unsupported Claude model: %s", model)
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedModel, model)
	}

	return &ClaudeClient{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
		config: config.GetConfig(),
	}, nil
}

// Translate sends a prompt to the Claude API and returns the response
func (c *ClaudeClient) Translate(prompt string, context string) (string, error) {
	log.Debug(i18n.Messages.TranslationStarted, "Claude", c.model)
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
			log.Debug(i18n.Messages.TranslationCompleted, "Claude", c.model)
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

func isRateLimitError(err error) bool {
	return strings.Contains(err.Error(), "rate_limit_error")
}

func (c *ClaudeClient) makeRequest(prompt string, context string) (string, error) {
	log.Debug("Making request to Claude API")
	url := config.GetConfig().ClaudeAPIURL

	requestBody, err := json.Marshal(map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"system":     context,
		"max_tokens": c.config.MaxTokens,
	})
	if err != nil {
		log.Error("Error marshalling request: %v", err)
		return "", fmt.Errorf("error marshalling request: %w", err)
	}
	//log.Debug("Claude API Request Body : %s", requestBody)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Error("Error creating request: %v", err)
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	//log.Debug("Sending request to Claude API : %s", prompt)
	resp, err := c.client.Do(req)
	if err != nil {
		log.Error("Error sending request: %v", err)
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error reading response: %v", err)
		return "", fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Error("API request failed with status code %d: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Error("Error unmarshalling response: %v", err)
		return "", fmt.Errorf("error unmarshalling response: %w", err)
	}

	if len(response.Content) == 0 {
		log.Error("No content in response")
		return "", fmt.Errorf("no content in response")
	}
	//log.Debug("Claude API Response : %s", response.Content)
	log.Debug("Successfully received and parsed response from Claude API.")
	return response.Content[0].Text, nil
}

// ProcessWithPrompt processes a prompt template with the given values and sends it to the Claude API
func (c *ClaudeClient) ProcessWithPrompt(promptTemplate *prompt.PromptTemplate, values map[string]string) (string, error) {
	log.Debug("Processing prompt with Claude")
	formattedPrompt := promptTemplate.Format(values)

	// Utilisez la méthode Translate existante pour envoyer le prompt formatté
	return c.Translate(formattedPrompt, "")
}
