package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/chrlesur/Ontology/internal/i18n"
	"gopkg.in/yaml.v2"
)

type MarkdownParser struct {
	metadata map[string]string
}

func init() {
	RegisterParser(".md", NewMarkdownParser)
}

func NewMarkdownParser() Parser {
	return &MarkdownParser{
		metadata: make(map[string]string),
	}
}

func (p *MarkdownParser) Parse(reader io.Reader) ([]byte, error) {
	log.Debug(i18n.Messages.ParseStarted, "Markdown")

	var content bytes.Buffer
	scanner := bufio.NewScanner(reader)
	inFrontMatter := false
	var frontMatter strings.Builder
	lineCount, wordCount, charCount, headerCount := 0, 0, 0, 0

	for scanner.Scan() {
		line := scanner.Text()

		if lineCount == 0 && line == "---" {
			inFrontMatter = true
			continue
		}

		if inFrontMatter {
			if line == "---" {
				inFrontMatter = false
				p.parseFrontMatter(frontMatter.String())
			} else {
				frontMatter.WriteString(line + "\n")
			}
			continue
		}

		content.WriteString(line + "\n")
		lineCount++
		wordCount += len(strings.Fields(line))
		charCount += len(line)

		if strings.HasPrefix(line, "#") {
			headerCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	p.metadata["format"] = "Markdown"
	p.metadata["lineCount"] = fmt.Sprintf("%d", lineCount)
	p.metadata["wordCount"] = fmt.Sprintf("%d", wordCount)
	p.metadata["charCount"] = fmt.Sprintf("%d", charCount)
	p.metadata["headerCount"] = fmt.Sprintf("%d", headerCount)

	log.Info(i18n.Messages.ParseCompleted, "Markdown")
	return content.Bytes(), nil
}

func (p *MarkdownParser) parseFrontMatter(frontMatter string) {
	var fm map[string]interface{}
	err := yaml.Unmarshal([]byte(frontMatter), &fm)
	if err != nil {
		log.Warning("Failed to parse YAML front matter: %v", err)
		return
	}

	for key, value := range fm {
		p.metadata[key] = fmt.Sprintf("%v", value)
	}
}

func (p *MarkdownParser) GetFormatMetadata() map[string]string {
	return p.metadata
}
