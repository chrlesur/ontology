// internal/llm/claude.go

package llm

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"

    "github.com/chrlesur/Ontology/internal/logger"
    "github.com/chrlesur/Ontology/internal/i18n"
    "github.com/chrlesur/Ontology/internal/config"
)

// ClaudeClient implements the Client interface for Claude
type ClaudeClient struct {
    apiKey string
    model  string
    client *http.Client
}

// supportedClaudeModels defines the list of supported Claude models
var supportedClaudeModels = map[string]bool{
    "claude-3-5-sonnet-20240620": true,
    "claude-3-opus-20240229":     true,
    "claude-3-haiku-20240307":    true,
}

// NewClaudeClient creates a new Claude client
func NewClaudeClient(apiKey string, model string) (*ClaudeClient, error) {
    if apiKey == "" {
        return nil, ErrAPIKeyMissing
    }

    if !supportedClaudeModels[model] {
        return nil, fmt.Errorf("%w: %s", ErrUnsupportedModel, model)
    }

    return &ClaudeClient{
        apiKey: apiKey,
        model:  model,
        client: &http.Client{Timeout: 30 * time.Second},
    }, nil
}

// Translate sends a prompt to the Claude API and returns the response
func (c *ClaudeClient) Translate(prompt string, context string) (string, error) {
    logger.Debug(i18n.TranslationStarted, "Claude", c.model)

    var result string
    var err error
    for attempt := 1; attempt <= 5; attempt++ {
        result, err = c.makeRequest(prompt, context)
        if err == nil {
            break
        }
        logger.Warning(i18n.TranslationRetry, attempt, err)
        time.Sleep(time.Duration(attempt) * time.Second)
    }

    if err != nil {
        return "", fmt.Errorf("%w: %v", ErrTranslationFailed, err)
    }

    logger.Info(i18n.TranslationCompleted, "Claude", c.model)
    return result, nil
}

func (c *ClaudeClient) makeRequest(prompt string, context string) (string, error) {
    url := config.GetConfig().ClaudeAPIURL

    requestBody, err := json.Marshal(map[string]interface{}{
        "model": c.model,
        "messages": []map[string]string{
            {"role": "system", "content": context},
            {"role": "user", "content": prompt},
        },
    })
    if err != nil {
        return "", fmt.Errorf("error marshalling request: %w", err)
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
    if err != nil {
        return "", fmt.Errorf("error creating request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-api-key", c.apiKey)

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
        Content []struct {
            Text string `json:"text"`
        } `json:"content"`
    }

    err = json.Unmarshal(body, &response)
    if err != nil {
        return "", fmt.Errorf("error unmarshalling response: %w", err)
    }

    if len(response.Content) == 0 {
        return "", fmt.Errorf("no content in response")
    }

    return response.Content[0].Text, nil
}