
# Development Guide

This guide provides information for developers who want to contribute to the Ontology project.

## Setting Up the Development Environment

1. Ensure you have Go 1.16 or later installed.
2. Clone the repository:
   ```
   git clone https://github.com/chrlesur/Ontology.git
   cd Ontology
   ```
3. Install dependencies:
   ```
   go mod download
   ```

## Project Structure

- `cmd/ontology`: Main application entry point
- `internal/`: Internal packages
  - `config/`: Configuration management
  - `converter/`: Ontology format converters
  - `i18n/`: Internationalization
  - `llm/`: Language Model integrations
  - `logger/`: Logging utilities
  - `parser/`: Document parsers
  - `pipeline/`: Main processing pipeline
  - `segmenter/`: Document segmentation
- `docs/`: Project documentation

## Coding Standards

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Ensure all exported functions, types, and variables have proper documentation comments
- Keep functions under 80 lines where possible
- Use meaningful variable and function names

## Testing

- Write unit tests for all new functionality
- Run tests using:
  ```
  go test ./...
  ```
- Aim for at least 80% test coverage for new code

## Adding New Features

1. Create a new branch for your feature
2. Implement the feature, ensuring it adheres to the project's architecture
3. Add appropriate unit tests
4. Update documentation as necessary
5. Submit a pull request with a clear description of the changes

## Submitting Pull Requests

1. Ensure your code passes all tests
2. Update the CHANGELOG.md file with your changes
3. Submit the pull request with a clear title and description
4. Be prepared to respond to feedback and make necessary changes

## Reporting Issues

- Use the GitHub issue tracker to report bugs
- Provide a clear description of the issue, including steps to reproduce
- Include relevant logs and error messages if applicable

