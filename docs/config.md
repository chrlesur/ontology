# Configuration Guide

Ontology uses a YAML configuration file for managing various settings. By default, it looks for a file named `config.yaml` in the current directory. You can also specify a different configuration file path using the environment variable `ONTOLOGY_CONFIG_PATH`.

## Configuration Options

```yaml
base_uri: "http://www.wikidata.org/entity/"
openai_api_url: "https://api.openai.com/v1/chat/completions"
claude_api_url: "https://api.anthropic.com/v1/messages"
ollama_api_url: "http://localhost:11434/api/generate"
log_directory: "logs"
log_level: "info"
max_tokens: 8000
context_size: 4000
default_llm: "claude"
default_model: "claude-3-5-sonnet-20240620"
Explanation of Options
base_uri: The base URI for entity identifiers
openai_api_url: API endpoint for OpenAI
claude_api_url: API endpoint for Claude
ollama_api_url: API endpoint for Ollama
log_directory: Directory for storing log files
log_level: Logging level (debug, info, warning, error)
max_tokens: Maximum number of tokens per segment
context_size: Size of context to maintain between segments
default_llm: Default LLM provider to use
default_model: Default model for the chosen LLM provider
```

## Environment Variables
Some configuration options can be overridden using environment variables:

```
OPENAI_API_KEY: API key for OpenAI
CLAUDE_API_KEY: API key for Claude
```

## Command-line Overrides
Many configuration options can be overridden via command-line flags. See the usage documentation for more details.