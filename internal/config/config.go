package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
	BaseURI      string `yaml:"base_uri"`
	OpenAIAPIURL string `yaml:"openai_api_url"`
	ClaudeAPIURL string `yaml:"claude_api_url"`
	OllamaAPIURL string `yaml:"ollama_api_url"`
	OpenAIAPIKey string `yaml:"openai_api_key"`
	ClaudeAPIKey string `yaml:"claude_api_key"`
	LogDirectory string `yaml:"log_directory"`
	LogLevel     string `yaml:"log_level"`
	MaxTokens    int    `yaml:"max_tokens"`
	ContextSize  int    `yaml:"context_size"`
	DefaultLLM   string `yaml:"default_llm"`
	DefaultModel string `yaml:"default_model"`
	OntologyName string `yaml:"ontology_name"`
	ExportRDF    bool   `yaml:"export_rdf"`
	ExportOWL    bool   `yaml:"export_owl"`
}

// GetConfig returns the singleton instance of Config
func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{
			OpenAIAPIURL: "https://api.openai.com/v1/chat/completions",
			ClaudeAPIURL: "https://api.anthropic.com/v1/messages",
			OllamaAPIURL: "http://localhost:11434/api/generate",
			BaseURI:      "http://www.wikidata.org/entity/",
			LogDirectory: "logs",
			LogLevel:     "info",
			MaxTokens:    4000,
			ContextSize:  500,
			DefaultLLM:   "openai",
			DefaultModel: "gpt-3.5-turbo",
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
