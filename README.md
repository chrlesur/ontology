# Ontology

Ontology is a command-line tool for enriching ontologies from various document formats. It supports multiple input formats and can utilize different language models for analysis.

## Installation

### Prerequisites

- Go 1.16 or later

### Building from Source

1. Clone the repository:
   ```
   git clone https://github.com/chrlesur/Ontology.git
   cd Ontology
   ```

2. Build the project:
   ```
   go build -o ontology ./cmd/ontology
   ```
   This will create the `ontology` executable in the current directory.

3. (Optional) Install the binary to your $GOPATH/bin:
   ```
   go install ./cmd/ontology
   ```

### Cross-compilation

To build for a specific operating system, use:

```
GOOS=<target-os> go build -o ontology-<target-os> ./cmd/ontology
```

Replace `<target-os>` with `linux`, `darwin`, or `windows`.

Examples:
```
GOOS=linux go build -o ontology-linux ./cmd/ontology
GOOS=darwin go build -o ontology-darwin ./cmd/ontology
GOOS=windows go build -o ontology-windows.exe ./cmd/ontology
```

### Running the Application

After building, you can run the application using:

```
./ontology
```

Or, if you've installed it to your $GOPATH/bin:

```
ontology
```

## Configuration

Ontology uses a YAML configuration file. By default, it looks for `.ontology.yaml` in the current directory. You can specify a different configuration file using the `--config` flag.

Example configuration:

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
```

For more details on configuration options, see the [Configuration Guide](docs/configuration.md).

## Usage

### Command Line Interface

To enrich an ontology from input documents:

```
ontology enrich --input <input_file_or_directory> --output <output_file> [flags]
```

Flags:
- `--input`: Input file or directory (required)
- `--output`: Output file for the enriched ontology (required)
- `--format`: Input format (auto-detected if not specified)
- `--llm`: Language model to use for analysis
- `--llm-model`: Specific model for the chosen LLM
- `--passes`: Number of passes for ontology enrichment
- `--recursive`: Process input directory recursively

To check the version of Ontology:

```
ontology version
```

For more examples and detailed usage instructions, see the [Usage Guide](docs/usage.md).

## Contributing

We welcome contributions to the Ontology project. Please read our [Contribution Guidelines](docs/contributing.md) for details on how to submit pull requests, report issues, and suggest improvements.

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.
