package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/chrlesur/Ontology/internal/i18n"
)

type TextParser struct {
	metadata map[string]string
}

func init() {
	RegisterParser(".txt", NewTextParser)
}

func NewTextParser() Parser {
	return &TextParser{
		metadata: make(map[string]string),
	}
}

func (p *TextParser) Parse(reader io.Reader) ([]byte, error) {
	log.Debug(i18n.Messages.ParseStarted, "Text")

	var content strings.Builder
	scanner := bufio.NewScanner(reader)
	lineCount := 0
	wordCount := 0
	charCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		content.WriteString(line)
		content.WriteString("\n")

		lineCount++
		words := strings.Fields(line)
		wordCount += len(words)
		charCount += utf8.RuneCountInString(line)
	}

	if err := scanner.Err(); err != nil {
		log.Error(i18n.Messages.ParseFailed, "Text", err)
		return nil, fmt.Errorf("failed to read text content: %w", err)
	}

	p.extractMetadata(lineCount, wordCount, charCount)

	log.Info(i18n.Messages.ParseCompleted, "Text")
	return []byte(content.String()), nil
}

func (p *TextParser) extractMetadata(lineCount, wordCount, charCount int) {
	p.metadata["format"] = "Text"
	p.metadata["lineCount"] = fmt.Sprintf("%d", lineCount)
	p.metadata["wordCount"] = fmt.Sprintf("%d", wordCount)
	p.metadata["charCount"] = fmt.Sprintf("%d", charCount)
}

func (p *TextParser) GetFormatMetadata() map[string]string {
	return p.metadata
}
