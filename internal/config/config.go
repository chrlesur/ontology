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
    BaseURI          string        `yaml:"base_uri"`
    OpenAIAPIURL     string        `yaml:"openai_api_url"`
    ClaudeAPIURL     string        `yaml:"claude_api_url"`
    OllamaAPIURL     string        `yaml:"ollama_api_url"`
    OpenAIAPIKey     string        `yaml:"openai_api_key"`
    ClaudeAPIKey     string        `yaml:"claude_api_key"`
    LogDirectory     string        `yaml:"log_directory"`
    LogLevel         string        `yaml:"log_level"`
    MaxTokens        int           `yaml:"max_tokens"`
    ContextSize      int           `yaml:"context_size"`
    DefaultLLM       string        `yaml:"default_llm"`
    DefaultModel     string        `yaml:"default_model"`
    OntologyName     string        `yaml:"ontology_name"`
    Input            string        `yaml:"input"`
    IncludePositions bool          `yaml:"include_positions"`
    ContextOutput    bool          `yaml:"context_output"`
    ContextWords     int           `yaml:"context_words"`
    AIYOUAPIURL      string        `yaml:"aiyou_api_url"`
    AIYOUAssistantID string        `yaml:"aiyou_assistant_id"`
    AIYOUEmail       string        `yaml:"aiyou_email"`
    AIYOUPassword    string        `yaml:"aiyou_password"`
    Storage          StorageConfig `yaml:"storage"`
}

// StorageConfig contient la configuration pour le stockage
type StorageConfig struct {
    Type     string  `yaml:"type"`
    LocalPath string `yaml:"local_path"`
    S3       S3Config `yaml:"s3"`
}

// S3Config contient la configuration spécifique à S3
type S3Config struct {
    Bucket          string `yaml:"bucket"`
    Region          string `yaml:"region"`
    Endpoint        string `yaml:"endpoint"`
    AccessKeyID     string `yaml:"access_key_id"`
    SecretAccessKey string `yaml:"secret_access_key"`
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
            ContextOutput:    false,
            ContextWords:     30,
            AIYOUAssistantID: "asst_q2YbeHKeSxBzNr43KhIESkqj",
            AIYOUAPIURL:      "https://ai.dragonflygroup.fr/api",
            Storage: StorageConfig{
                Type: "local",
                LocalPath: ".",
            },
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
    if s3AccessKey := os.Getenv("S3_ACCESS_KEY_ID"); s3AccessKey != "" {
        c.Storage.S3.AccessKeyID = s3AccessKey
    }
    if s3SecretKey := os.Getenv("S3_SECRET_ACCESS_KEY"); s3SecretKey != "" {
        c.Storage.S3.SecretAccessKey = s3SecretKey
    }
    if s3Endpoint := os.Getenv("S3_ENDPOINT"); s3Endpoint != "" {
        c.Storage.S3.Endpoint = s3Endpoint
    }
    // Add more environment variables as needed
}

// ValidateConfig checks if the configuration is valid
func (c *Config) ValidateConfig() error {
    if c.OpenAIAPIKey == "" && c.ClaudeAPIKey == "" {
        return fmt.Errorf(i18n.GetMessage("ErrNoAPIKeys"))
    }
    if c.Storage.Type != "local" && c.Storage.Type != "s3" {
        return fmt.Errorf("invalid storage type: %s", c.Storage.Type)
    }
    if c.Storage.Type == "s3" {
        if c.Storage.S3.Bucket == "" {
            return fmt.Errorf("S3 bucket is required when using S3 storage")
        }
        if c.Storage.S3.Region == "" {
            return fmt.Errorf("S3 region is required when using S3 storage")
        }
        // Nous ne validons pas l'endpoint ici car il peut être optionnel (utilisation de S3 standard)
        // mais nous pourrions ajouter un avertissement si l'endpoint n'est pas défini
        if c.Storage.S3.Endpoint == "" {
            log.Println("Warning: S3 endpoint is not set. Using default S3 endpoint.")
        }
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