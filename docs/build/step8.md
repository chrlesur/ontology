
Étape 8 : Implémentation du système de configuration

Directives générales (à suivre impérativement pour toutes les étapes du projet) :
1. Utilisez exclusivement Go dans sa dernière version stable.
2. Assurez-vous qu'aucun fichier de code source ne dépasse 3000 tokens.
3. Limitez chaque package à un maximum de 10 méthodes exportées.
4. Aucune méthode ne doit dépasser 80 lignes de code.
5. Suivez les meilleures pratiques et les modèles idiomatiques de Go.
6. Tous les messages visibles par l'utilisateur doivent être en anglais.
7. Chaque fonction, méthode et type exporté doit avoir un commentaire de documentation conforme aux standards GoDoc.
8. Utilisez le package 'internal/logger' pour toute journalisation. Implémentez les niveaux de log : debug, info, warning, et error.
9. Toutes les valeurs configurables doivent être définies dans le package 'internal/config'.
10. Gérez toutes les erreurs de manière appropriée, en utilisant error wrapping lorsque c'est pertinent.
11. Pour les messages utilisateur, utilisez les constantes définies dans le package 'internal/i18n'.
12. Assurez-vous que le code est prêt pour de futurs efforts de localisation.
13. Optimisez le code pour la performance, particulièrement pour le traitement de grands documents.
14. Implémentez des tests unitaires pour chaque nouvelle fonction ou méthode.
15. Veillez à ce que le code soit sécurisé, en particulier lors du traitement des entrées utilisateur.

Tâche spécifique : Amélioration et extension du système de configuration

1. Dans le package internal/config, mettez à jour le fichier config.go avec le contenu suivant :

```go
package config

import (
    "fmt"
    "io/ioutil"
    "os"
    "sync"

    "gopkg.in/yaml.v2"
    "github.com/chrlesur/Ontology/internal/i18n"
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

// LoadConfig loads the configuration from a YAML file
func (c *Config) loadConfigFile() {
    configPath := os.Getenv("ONTOLOGY_CONFIG_PATH")
    if configPath == "" {
        configPath = "config.yaml"
    }

    data, err := ioutil.ReadFile(configPath)
    if err != nil {
        fmt.Printf(i18n.GetMessage("ErrReadConfigFile"), err)
        return
    }

    err = yaml.Unmarshal(data, c)
    if err != nil {
        fmt.Printf(i18n.GetMessage("ErrParseConfigFile"), err)
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
```

2. Créez un fichier config_test.go dans le même package et implémentez des tests unitaires pour toutes les fonctions de configuration.

3. Créez un fichier config.yaml à la racine du projet avec un exemple de configuration :

```yaml
base_uri: "http://www.wikidata.org/entity/"
openai_api_url: "https://api.openai.com/v1/chat/completions"
claude_api_url: "https://api.anthropic.com/v1/messages"
ollama_api_url: "http://localhost:11434/api/generate"
log_directory: "logs"
log_level: "info"
max_tokens: 4000
context_size: 500
default_llm: "openai"
default_model: "gpt-3.5-turbo"
```

4. Mettez à jour le fichier internal/i18n/messages.go pour inclure les nouveaux messages d'erreur :

```go
const (
    // ... autres constantes existantes ...
    ErrReadConfigFile  = "Failed to read config file: %v"
    ErrParseConfigFile = "Failed to parse config file: %v"
    ErrNoAPIKeys       = "No API keys provided for any LLM service"
)
```

5. Mettez à jour la fonction main() dans cmd/ontology/main.go pour charger et valider la configuration au démarrage :

```go
func main() {
    config := config.GetConfig()
    if err := config.ValidateConfig(); err != nil {
        fmt.Printf("Configuration error: %v\n", err)
        os.Exit(1)
    }
    ontology.Execute()
}
```

6. Mettez à jour les autres packages (llm, parser, segmenter, etc.) pour utiliser les valeurs de configuration plutôt que des valeurs codées en dur.

7. Implémentez une fonction pour recharger la configuration à chaud (si nécessaire) :

```go
func (c *Config) Reload() error {
    c.loadConfigFile()
    c.loadEnvVariables()
    return c.ValidateConfig()
}
```

8. Assurez-vous que toutes les fonctions respectent les limites de taille (pas plus de 80 lignes par fonction).

9. Documentez toutes les fonctions exportées avec des commentaires GoDoc.

10. Implémentez des tests pour vérifier que la configuration est correctement chargée depuis le fichier YAML et les variables d'environnement.

Après avoir terminé ces tâches, exécutez tous les tests unitaires et assurez-vous que le code compile sans erreur. Vérifiez également que la configuration fonctionne correctement dans différents scénarios (fichier de configuration manquant, variables d'environnement, etc.).
