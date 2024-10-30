package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/prompt"
)

type OllamaClient struct {
	model  string
	client *http.Client
	config *config.Config
}

var supportedOllamaModels = map[string]bool{
	"llama3.2":      true,
	"llama3.1":      true,
	"mistral-nemo":  true,
	"mixtral":       true,
	"mistral":       true,
	"mistral-small": true,
}

func NewOllamaClient(model string) (*OllamaClient, error) {
	log.Debug("Creating new Ollama client with model: %s", model)
	if !supportedOllamaModels[model] {
		log.Error("Unsupported Ollama model: %s", model)
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedModel, model)
	}

	return &OllamaClient{
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
		config: config.GetConfig(),
	}, nil
}

func (c *OllamaClient) Translate(prompt string, context string) (string, error) {
	log.Debug(i18n.Messages.TranslationStarted, "Ollama", c.model)
	log.Debug("Starting Translate. Prompt length: %d, Context length: %d", len(prompt), len(context))

	var result string
	var err error
	maxRetries := 5
	baseDelay := time.Second * 60
	maxDelay := time.Minute * 5

	for attempt := 0; attempt < maxRetries; attempt++ {
		log.Debug("Attempt %d of %d", attempt+1, maxRetries)
		result, err = c.makeRequest(prompt, context)
		if err == nil {
			log.Debug(i18n.Messages.TranslationCompleted, "Ollama", c.model)
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

func (c *OllamaClient) makeRequest(prompt string, context string) (string, error) {
	log.Debug("Making request to Ollama API")
	url := c.config.OllamaAPIURL

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":  c.model,
		"prompt": prompt,
		"system": context,
		"stream": false,
	})
	if err != nil {
		log.Error("Error marshalling request: %v", err)
		return "", fmt.Errorf("error marshalling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Error("Error creating request: %v", err)
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
		Response string `json:"response"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Error("Error unmarshalling response: %v", err)
		return "", fmt.Errorf("error unmarshalling response: %w", err)
	}

	log.Debug("Successfully received and parsed response from Ollama API.")
	return response.Response, nil
}

func (c *OllamaClient) ProcessWithPrompt(promptTemplate *prompt.PromptTemplate, values map[string]string) (string, error) {
	log.Debug("Processing prompt with Ollama")
	formattedPrompt := promptTemplate.Format(values)

	return c.Translate(formattedPrompt, "")
}
