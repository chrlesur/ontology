# Architecture of Ontology

## Overview

Ontology is structured as a command-line application with a modular architecture. The main components are:

1. CLI Interface
2. Configuration Management
3. Document Parsing
4. Segmentation
5. LLM Integration
6. Ontology Generation
7. Export Formats (QuickStatement, RDF, OWL)

## Component Details

### 1. CLI Interface
- Implemented using Cobra
- Handles user input and command execution
- Located in `cmd/ontology`

### 2. Configuration Management
- Manages application settings
- Uses YAML for configuration files
- Located in `internal/config`

### 3. Document Parsing
- Supports multiple file formats (TXT, PDF, Markdown, HTML, DOCX)
- Modular design allows easy addition of new formats
- Located in `internal/parser`

### 4. Segmentation
- Breaks large documents into manageable segments
- Ensures context preservation between segments
- Located in `internal/segmenter`

### 5. LLM Integration
- Supports multiple LLM providers (OpenAI, Claude, Ollama)
- Handles API communication and error management
- Located in `internal/llm`

### 6. Ontology Generation
- Processes LLM output to create ontology structures
- Manages multi-pass enrichment
- Part of `internal/pipeline`

### 7. Export Formats
- Generates QuickStatement TSV format
- Provides RDF and OWL export options
- Located in `internal/converter`

## Data Flow

1. User input → CLI Interface
2. CLI Interface → Configuration Management
3. Document Parsing → Segmentation
4. Segmentation → LLM Integration
5. LLM Integration → Ontology Generation
6. Ontology Generation → Export Formats
7. Export Formats → Output files

## Extension Points

- New document formats can be added by implementing the `Parser` interface in `internal/parser`
- Additional LLM providers can be integrated by implementing the `Client` interface in `internal/llm`
- New export formats can be added to the `internal/converter` package