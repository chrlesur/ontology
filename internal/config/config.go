package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/chrlesur/Ontology/internal/i18n"
	"gopkg.in/yaml.v2"
)

var (
	once     sync.Once
	instance *Config
)

// Config structure definition
type Config struct {
	BaseURI          string `yaml:"base_uri"`
	OpenAIAPIURL     string `yaml:"openai_api_url"`
	ClaudeAPIURL     string `yaml:"claude_api_url"`
	OllamaAPIURL     string `yaml:"ollama_api_url"`
	OpenAIAPIKey     string `yaml:"openai_api_key"`
	ClaudeAPIKey     string `yaml:"claude_api_key"`
	LogDirectory     string `yaml:"log_directory"`
	LogLevel         string `yaml:"log_level"`
	MaxTokens        int    `yaml:"max_tokens"`
	ContextSize      int    `yaml:"context_size"`
	DefaultLLM       string `yaml:"default_llm"`
	DefaultModel     string `yaml:"default_model"`
	OntologyName     string `yaml:"ontology_name"`
	Input            string `yaml:"input"`
	IncludePositions bool   `yaml:"include_positions"`
	ContextOutput    bool   `yaml:"context_output"`
	ContextWords     int    `yaml:"context_words"`
	AIYOUAPIURL      string `yaml:"aiyou_api_url"`
	AIYOUAssistantID string `yaml:"aiyou_assistant_id"`
	AIYOUEmail       string `yaml:"aiyou_email"`
	AIYOUPassword    string `yaml:"aiyou_password"`
}

// GetConfig returns the singleton instance of Config
func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{
			OpenAIAPIURL:     "https://api.openai.com/v1/chat/completions",
			ClaudeAPIURL:     "https://api.anthropic.com/v1/messages",
			OllamaAPIURL:     "http://localhost:11434/api/generate",
			BaseURI:          "http://www.wikidata.org/entity/",
			LogDirectory:     "logs",
			LogLevel:         "info",
			MaxTokens:        1000,
			ContextSize:      4000,
			DefaultLLM:       "claude",
			DefaultModel:     "claude-3-5-sonnet-20240620",
			IncludePositions: true,
			ContextOutput:    false,                           // Par défaut, la sortie de contexte est désactivée
			ContextWords:     30,                              // Par défaut, 30 mots de contexte
			AIYOUAssistantID: "asst_q2YbeHKeSxBzNr43KhIESkqj", // Secnumcloud
			AIYOUAPIURL:      "https://ai.dragonflygroup.fr/api",
		}
		instance.loadConfigFile()
		instance.loadEnvVariables()
	})
	return instance
}

// loadConfigFile loads the configuration from a YAML file
func (c *Config) loadConfigFile() {
	configPath := os.Getenv("ONTOLOGY_CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Printf(i18n.GetMessage("ErrReadConfigFile"), err)
		return
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		log.Printf(i18n.GetMessage("ErrParseConfigFile"), err)
	}
	if contextOutput := os.Getenv("ONTOLOGY_CONTEXT_OUTPUT"); contextOutput == "true" {
		c.ContextOutput = true
	}
	if contextWords := os.Getenv("ONTOLOGY_CONTEXT_WORDS"); contextWords != "" {
		if words, err := strconv.Atoi(contextWords); err == nil {
			c.ContextWords = words
		}
	}
}

// loadEnvVariables loads configuration from environment variables
func (c *Config) loadEnvVariables() {
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		c.OpenAIAPIKey = apiKey
	}
	if apiKey := os.Getenv("CLAUDE_API_KEY"); apiKey != "" {
		c.ClaudeAPIKey = apiKey
	}
	// Add more environment variables as needed
}

// ValidateConfig checks if the configuration is valid
func (c *Config) ValidateConfig() error {
	if c.OpenAIAPIKey == "" && c.ClaudeAPIKey == "" {
		return fmt.Errorf(i18n.GetMessage("ErrNoAPIKeys"))
	}
	// Add more validation checks as needed
	return nil
}

// Reload reloads the configuration from file and environment variables
func (c *Config) Reload() error {
	c.loadConfigFile()
	c.loadEnvVariables()
	return c.ValidateConfig()
}
