package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/logger"
	"github.com/chrlesur/Ontology/internal/prompt"
)

const AIYOUAPIURL = "https://ai.dragonflygroup.fr/api"

type APICaller interface {
	Call(endpoint, method string, data interface{}, response interface{}) error
	SetToken(token string)
}

type AIYOUClient struct {
	apiCaller   APICaller
	AssistantID string
	Timeout     time.Duration
	config      *config.Config
	logger      *logger.Logger
}

type HTTPAPICaller struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

type Run struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Response string `json:"response"`
}

func NewAIYOUClient(assistantID string, email string, password string) (*AIYOUClient, error) {
	cfg := config.GetConfig()
	log := logger.GetLogger()

	if assistantID == "" {
		return nil, fmt.Errorf("AI.YOU assistant ID is required")
	}

	client := &AIYOUClient{
		apiCaller:   NewHTTPAPICaller(AIYOUAPIURL),
		AssistantID: assistantID,
		Timeout:     120 * time.Second,
		config:      cfg,
		logger:      log,
	}

	err := client.Login(email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to login to AI.YOU: %w", err)
	}

	return client, nil
}

func (c *AIYOUClient) Login(email, password string) error {
	c.logger.Info("Attempting to log in to AI.YOU")
	loginData := map[string]string{
		"email":    email,
		"password": password,
	}

	var loginResp struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expires_at"`
	}

	err := c.apiCaller.Call("/login", "POST", loginData, &loginResp)
	if err != nil {
		c.logger.Error(fmt.Sprintf("Login failed: %v", err))
		return fmt.Errorf("login failed: %w", err)
	}

	c.apiCaller.SetToken(loginResp.Token)
	c.logger.Info("Successfully logged in to AI.YOU")
	return nil
}

func (c *AIYOUClient) Translate(prompt string, context string) (string, error) {
	c.logger.Debug("Starting AI.YOU translation")

	threadID, err := c.CreateThread()
	if err != nil {
		return "", fmt.Errorf("failed to create thread: %w", err)
	}

	fullPrompt := fmt.Sprintf("%s\n\nContext: %s", prompt, context)
	response, err := c.ChatInThread(threadID, fullPrompt)
	if err != nil {
		return "", fmt.Errorf("error during chat: %w", err)
	}

	c.logger.Debug("AI.YOU translation completed")
	return response, nil
}

func (c *AIYOUClient) ProcessWithPrompt(promptTemplate *prompt.PromptTemplate, values map[string]string) (string, error) {
	c.logger.Debug("Processing with prompt using AI.YOU")

	formattedPrompt := promptTemplate.Format(values)
	return c.Translate(formattedPrompt, "")
}

func (c *AIYOUClient) CreateThread() (string, error) {
	c.logger.Debug("Creating new AI.YOU thread")

	var threadResp struct {
		ID string `json:"id"`
	}

	err := c.apiCaller.Call("/v1/threads", "POST", map[string]string{}, &threadResp)
	if err != nil {
		return "", fmt.Errorf("error creating thread: %w", err)
	}

	if threadResp.ID == "" {
		return "", fmt.Errorf("thread ID is empty in response")
	}

	c.logger.Debug(fmt.Sprintf("Thread created with ID: %s", threadResp.ID))
	return threadResp.ID, nil
}

func (c *AIYOUClient) ChatInThread(threadID, input string) (string, error) {
	c.logger.Debug(fmt.Sprintf("Chatting in thread %s", threadID))

	err := c.addMessage(threadID, input)
	if err != nil {
		return "", fmt.Errorf("failed to add message: %w", err)
	}

	runID, err := c.createRun(threadID)
	if err != nil {
		return "", fmt.Errorf("failed to create run: %w", err)
	}

	completedRun, err := c.waitForCompletion(threadID, runID)
	if err != nil {
		return "", fmt.Errorf("run failed: %w", err)
	}

	return completedRun.Response, nil
}

func (c *AIYOUClient) addMessage(threadID, content string) error {
	c.logger.Debug(fmt.Sprintf("Adding message to thread %s", threadID))

	messageData := map[string]string{
		"role":    "user",
		"content": content,
	}

	var response interface{}
	err := c.apiCaller.Call(fmt.Sprintf("/v1/threads/%s/messages", threadID), "POST", messageData, &response)
	if err != nil {
		return fmt.Errorf("error adding message: %w", err)
	}

	c.logger.Debug("Message added successfully")
	return nil
}

func (c *AIYOUClient) createRun(threadID string) (string, error) {
	c.logger.Debug(fmt.Sprintf("Creating run for thread %s", threadID))

	runData := map[string]string{
		"assistantId": c.AssistantID,
	}

	var runResp struct {
		ID string `json:"id"`
	}
	err := c.apiCaller.Call(fmt.Sprintf("/v1/threads/%s/runs", threadID), "POST", runData, &runResp)
	if err != nil {
		return "", fmt.Errorf("error creating run: %w", err)
	}

	c.logger.Debug(fmt.Sprintf("Run created with ID: %s", runResp.ID))
	return runResp.ID, nil
}

func (c *AIYOUClient) waitForCompletion(threadID, runID string) (*Run, error) {
	maxAttempts := 30
	delayBetweenAttempts := 2 * time.Second

	for i := 0; i < maxAttempts; i++ {
		c.logger.Debug(fmt.Sprintf("Attempt %d to retrieve run status", i+1))
		run, err := c.retrieveRun(threadID, runID)
		if err != nil {
			return nil, err
		}

		switch run.Status {
		case "completed":
			c.logger.Debug("Run completed successfully")
			return run, nil
		case "failed", "cancelled":
			return nil, fmt.Errorf("run failed with status: %s", run.Status)
		default:
			c.logger.Debug(fmt.Sprintf("Waiting for run completion. Pausing for %v", delayBetweenAttempts))
			time.Sleep(delayBetweenAttempts)
		}
	}

	return nil, fmt.Errorf("timeout waiting for run completion")
}

func (c *AIYOUClient) retrieveRun(threadID, runID string) (*Run, error) {
	c.logger.Debug(fmt.Sprintf("Retrieving run %s for thread %s", runID, threadID))

	var runStatus Run
	err := c.apiCaller.Call(fmt.Sprintf("/v1/threads/%s/runs/%s", threadID, runID), "POST", map[string]string{}, &runStatus)
	if err != nil {
		return nil, fmt.Errorf("error retrieving run: %w", err)
	}

	c.logger.Debug(fmt.Sprintf("Run status retrieved: %v", runStatus))
	return &runStatus, nil
}

func NewHTTPAPICaller(baseURL string) APICaller {
	return &HTTPAPICaller{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *HTTPAPICaller) Call(endpoint, method string, data interface{}, response interface{}) error {
	url := c.baseURL + endpoint
	var req *http.Request
	var err error

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("error marshaling request data: %w", err)
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API error: status code %d, body: %s", resp.StatusCode, string(body))
	}

	err = json.Unmarshal(body, response)
	if err != nil {
		return fmt.Errorf("error unmarshaling response: %w", err)
	}

	return nil
}

func (c *HTTPAPICaller) SetToken(token string) {
	c.token = token
}
