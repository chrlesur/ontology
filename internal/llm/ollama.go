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
)

// OllamaClient implements the Client interface for Ollama
type OllamaClient struct {
	model  string
	client *http.Client
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
	if !supportedOllamaModels[model] {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedModel, model)
	}

	return &OllamaClient{
		model:  model,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Translate sends a prompt to the Ollama API and returns the response
func (c *OllamaClient) Translate(prompt string, context string) (string, error) {
	log.Debug(i18n.TranslationStarted, "Ollama", c.model)

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

	log.Info(i18n.TranslationCompleted, "Ollama", c.model)
	return result, nil
}

func (c *OllamaClient) makeRequest(prompt string, context string) (string, error) {
	url := config.GetConfig().OllamaAPIURL

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":  c.model,
		"prompt": prompt,
		"system": context,
		"stream": false,
	})
	if err != nil {
		return "", fmt.Errorf("error marshalling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Response string `json:"response"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response: %w", err)
	}

	return response.Response, nil
}
