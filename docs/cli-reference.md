# CLI Reference

## Global Flags

- `--config string`: Path to config file (default is ./config.yaml)
- `--debug`: Enable debug mode
- `--silent`: Silent mode, only show errors

## Commands

### enrich

Enriches an ontology from input documents.

Usage:
```
ontology enrich [flags]
```

Flags:
- `--input string`: Input file or directory (required)
- `--output string`: Output file for the enriched ontology (required)
- `--format string`: Input format (auto-detected if not specified)
- `--llm string`: Language model to use for analysis
- `--llm-model string`: Specific model for the chosen LLM
- `--passes int`: Number of passes for ontology enrichment (default 1)
- `--rdf`: Export ontology in RDF format
- `--owl`: Export ontology in OWL format
- `--recursive`: Process input directory recursively

Example:
```
ontology enrich --input ./documents --output enriched_ontology.tsv --llm openai --passes 2 --recursive
```

### version

Displays the current version of Ontology.

Usage:
```
ontology version
```

Example:
```
ontology version
```

Output:
```
Ontology version 1.0.0
```
