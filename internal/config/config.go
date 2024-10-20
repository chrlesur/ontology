package config

import (
	"sync"
)

var (
	once     sync.Once
	instance *Config
)

// Config structure definition
type Config struct {
	// ... autres champs existants ...
	BaseURI      string `yaml:"base_uri"`
	OpenAIAPIURL string `yaml:"openai_api_url"`
	ClaudeAPIURL string `yaml:"claude_api_url"`
}

// GetConfig returns the singleton instance of Config
func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{
			OpenAIAPIURL: "https://api.openai.com/v1/chat/completions",
			ClaudeAPIURL: "https://api.anthropic.com/v1/messages",
			BaseURI:      "http://www.wikidata.org/entity/",
		}
	})
	return instance
}

