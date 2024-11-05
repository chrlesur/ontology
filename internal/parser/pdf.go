package parser

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/ledongthuc/pdf"
)

type PDFParser struct {
	metadata map[string]string
}

func init() {
	RegisterParser(".pdf", NewPDFParser)
}

func NewPDFParser() Parser {
	return &PDFParser{
		metadata: make(map[string]string),
	}
}

func (p *PDFParser) Parse(reader io.Reader) ([]byte, error) {
	log.Debug(i18n.Messages.ParseStarted, "PDF")

	content, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Error(i18n.Messages.ParseFailed, "PDF", err)
		return nil, fmt.Errorf("failed to read PDF content: %w", err)
	}

	pdfReader, err := pdf.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		log.Error(i18n.Messages.ParseFailed, "PDF", err)
		return nil, fmt.Errorf("failed to create PDF reader: %w", err)
	}

	var textContent bytes.Buffer
	for i := 1; i <= pdfReader.NumPage(); i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			log.Warning("Failed to extract text from page %d: %v", i, err)
			continue
		}
		textContent.WriteString(text)
	}

	p.extractMetadata(pdfReader)

	log.Info(i18n.Messages.ParseCompleted, "PDF")
	return textContent.Bytes(), nil
}

func (p *PDFParser) extractMetadata(pdfReader *pdf.Reader) {
	p.metadata["format"] = "PDF"
	p.metadata["pageCount"] = fmt.Sprintf("%d", pdfReader.NumPage())

	info := pdfReader.Trailer().Key("Info")
	if info.IsNull() {
		return
	}

	metadataFields := []struct {
		key      string
		pdfField string
	}{
		{"title", "Title"},
		{"author", "Author"},
		{"subject", "Subject"},
		{"keywords", "Keywords"},
		{"creator", "Creator"},
		{"producer", "Producer"},
		{"creationDate", "CreationDate"},
		{"modificationDate", "ModDate"},
	}

	for _, field := range metadataFields {
		value := info.Key(field.pdfField)
		if !value.IsNull() {
			p.metadata[field.key] = value.String()
		}
	}
}

func (p *PDFParser) GetFormatMetadata() map[string]string {
	return p.metadata
}
