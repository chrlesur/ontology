package parser

import (
	"bytes"
	"fmt"
	"io"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/ledongthuc/pdf"
)

func init() {
	log.Debug("Registering PDF parser")
	RegisterParser(".pdf", NewPDFParser)
}

type PDFParser struct {
	metadata map[string]string
}

func NewPDFParser() Parser {
	log.Debug("Creating new PDF parser")
	return &PDFParser{
		metadata: make(map[string]string),
	}
}

func (p *PDFParser) Parse(path string) ([]byte, error) {
	log.Debug("Parsing PDF file: %s", path)
	content, err := ParsePDF(path)
	if err != nil {
		return nil, err
	}
	p.extractMetadata(path)
	return content, nil
}

func (p *PDFParser) GetMetadata() map[string]string {
	return p.metadata
}

func (p *PDFParser) extractMetadata(path string) {
	log.Debug("Extracting metadata from PDF file: %s", path)
	f, r, err := pdf.Open(path)
	if err != nil {
		log.Error("Failed to open PDF: %v", err)
		return
	}
	defer f.Close()

	// Extract basic information
	p.metadata["PageCount"] = fmt.Sprintf("%d", r.NumPage())

	log.Debug("Extracted %d metadata items", len(p.metadata))
}

// ParsePDF reads a PDF file and returns its content as a byte slice.
func ParsePDF(path string) ([]byte, error) {
	log.Debug(i18n.Messages.ParseStarted)

	f, r, err := pdf.Open(path)
	if err != nil {
		log.Error(i18n.Messages.ParseFailed, err)
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		log.Error("Failed to get plain text: %v", err)
		return nil, fmt.Errorf("failed to get plain text: %w", err)
	}

	_, err = io.Copy(&buf, b)
	if err != nil {
		log.Error("Failed to read content: %v", err)
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	content := buf.Bytes()
	log.Info(i18n.Messages.ParseCompleted)
	log.Debug("Total extracted content length: %d characters", len(content))
	return content, nil
}
