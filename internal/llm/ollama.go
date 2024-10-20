// internal/llm/ollama.go

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

// OllamaClient implements the Client interface for Ollama
type OllamaClient struct {
	model  string
	client *http.Client
	config *config.Config
}

// supportedOllamaModels defines the list of supported Ollama models
var supportedOllamaModels = map[string]bool{
	"llama3.2:3B":       true,
	"llama3.1:8B":       true,
	"mistral-nemo:12B":  true,
	"mixtral:7B":        true,
	"mistral:7B":        true,
	"mistral-small:22B": true,
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(model string) (*OllamaClient, error) {
	log.Debug("Creating new Ollama client with model: %s", model)
	if !supportedOllamaModels[model] {
		log.Error("Unsupported Ollama model: %s", model)
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedModel, model)
	}

	return &OllamaClient{
		model:  model,
		client: &http.Client{Timeout: 30 * time.Second},
		config: config.GetConfig(),
	}, nil
}

// Translate sends a prompt to the Ollama API and returns the response
func (c *OllamaClient) Translate(prompt string, context string) (string, error) {
	log.Debug(i18n.Messages.TranslationStarted, "Ollama", c.model)
	log.Debug("Prompt length: %d, Context length: %d", len(prompt), len(context))

	var result string
	var err error
	maxRetries := 5
	baseDelay := time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Debug("Attempt %d of %d", attempt, maxRetries)
		result, err = c.makeRequest(prompt, context)
		if err == nil {
			log.Info(i18n.Messages.TranslationCompleted, "Ollama", c.model)
			log.Debug("Translation successful, result length: %d", len(result))
			return result, nil
		}

		log.Warning(i18n.Messages.TranslationRetry, attempt, err)
		time.Sleep(time.Duration(attempt) * baseDelay)
	}

	log.Error(i18n.Messages.TranslationFailed, err)
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

	log.Debug("Ollama API Request Body: %s", requestBody)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Error("Error creating request: %v", err)
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	log.Debug("Sending request to Ollama API")
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

	log.Debug("Successfully received and parsed response from Ollama API: %s", response.Response)
	return response.Response, nil
}

// ProcessWithPrompt processes a prompt template with the given values and sends it to the Ollama API
func (c *OllamaClient) ProcessWithPrompt(promptTemplate *prompt.PromptTemplate, values map[string]string) (string, error) {
	log.Debug("Processing prompt with Ollama")
	formattedPrompt := promptTemplate.Format(values)
	log.Debug("Formatted prompt: %s", formattedPrompt)

	// Utilisez la méthode Translate existante pour envoyer le prompt formatté
	return c.Translate(formattedPrompt, "")
}
