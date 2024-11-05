# Ontology

Ontology is a command-line tool for enriching ontologies from various document formats. It supports multiple input formats, can utilize different language models for analysis, and now includes support for S3 storage.

## Installation

### Prerequisites

- Go 1.16 or later
- AWS CLI (for S3 functionality)

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

## Configuration

Ontology uses a YAML configuration file. By default, it looks for `config.yaml` in the current directory. You can specify a different configuration file using the `--config` flag.

Example configuration:

```yaml
base_uri: "http://www.wikidata.org/entity/"
openai_api_url: "https://api.openai.com/v1/chat/completions"
claude_api_url: "https://api.anthropic.com/v1/messages"
ollama_api_url: "http://localhost:11434/api/generate"
log_directory: "logs"
log_level: "info"
max_tokens: 2000
context_size: 4000
default_llm: "claude"
default_model: "claude-3-5-sonnet-20240620"
include_positions: true
context_output: false
context_words: 30
storage:
  type: "s3"  # Can be "local" or "s3"
  local_path: "."  # Default path for local storage
  s3:
    bucket: "bucket"  # S3 bucket name
    region: "fr1"  # S3 region
    endpoint: "https://ctsscfabf9.s3.fr1.cloud-temple.com"  # Custom S3 endpoint (optional, for S3-compatible systems like Dell ECS)
    access_key_id: ""  # S3 access key (can be left empty if configured via environment variables)
    secret_access_key: ""  # S3 secret key (can be left empty if configured via environment variables)
```

## Usage

### Command Line Interface

To enrich an ontology from input documents:

```
ontology enrich [input] [flags]
```

Flags:
- `--output`: Output file for the enriched ontology
- `--format`: Input format (auto-detected if not specified)
- `--llm`: Language model to use for analysis
- `--llm-model`: Specific model for the chosen LLM
- `--passes`: Number of passes for ontology enrichment (default 1)
- `--recursive`: Process input directory recursively
- `--existing-ontology`: Path to an existing ontology file to enrich
- `--include-positions`: Include position information in the ontology (default true)
- `--context-output`: Enable context output in JSON format
- `--context-words`: Number of context words before and after each position (default 30)
- `--entity-prompt`: Additional prompt for entity extraction
- `--relation-prompt`: Additional prompt for relation extraction
- `--enrichment-prompt`: Additional prompt for ontology enrichment
- `--merge-prompt`: Additional prompt for ontology merging

S3-specific flags:
- `--aiyou-assistant-id`: AI.YOU Assistant ID
- `--aiyou-email`: AI.YOU Email
- `--aiyou-password`: AI.YOU Password

Example usage with S3:
```
ontology enrich s3://your-bucket/your-file.txt --output s3://your-bucket/output.tsv --context-output
```

To check the version of Ontology:

```
ontology version
```

## New Features

- S3 Storage Support: Ontology can now read from and write to S3 buckets.
- Improved Context Generation: Added options for including word positions and generating context JSON.
- AI.YOU Integration: Support for AI.YOU API for additional language model capabilities.
- Enhanced Metadata Generation: Improved metadata handling for both local and S3 files.

## Contributing

We welcome contributions to the Ontology project. Please read our [Contribution Guidelines](docs/contributing.md) for details on how to submit pull requests, report issues, and suggest improvements.

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.