package parser

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/chrlesur/Ontology/internal/i18n"
)

type DOCXParser struct {
	metadata map[string]string
}

func init() {
	RegisterParser(".docx", NewDOCXParser)
}

func NewDOCXParser() Parser {
	return &DOCXParser{
		metadata: make(map[string]string),
	}
}

func (p *DOCXParser) Parse(reader io.Reader) ([]byte, error) {
	log.Debug(i18n.Messages.ParseStarted, "DOCX")

	// Lire tout le contenu en mémoire
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Error(i18n.Messages.ParseFailed, "DOCX", err)
		return nil, fmt.Errorf("failed to read DOCX content: %w", err)
	}

	// Créer un lecteur zip à partir du contenu
	zipReader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		log.Error(i18n.Messages.ParseFailed, "DOCX", err)
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
	}

	var textContent strings.Builder

	for _, file := range zipReader.File {
		switch file.Name {
		case "word/document.xml":
			if err := p.extractContent(file, &textContent); err != nil {
				return nil, err
			}
		case "docProps/core.xml":
			if err := p.extractMetadata(file); err != nil {
				log.Warning(i18n.Messages.MetadataExtractionFailed, "DOCX", err)
			}
		}
	}

	log.Info(i18n.Messages.ParseCompleted, "DOCX")
	return []byte(textContent.String()), nil
}

func (p *DOCXParser) extractContent(file *zip.File, textContent *strings.Builder) error {
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open document.xml: %w", err)
	}
	defer rc.Close()

	content, err := ioutil.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("failed to read document.xml: %w", err)
	}

	var document struct {
		Body struct {
			Paragraphs []struct {
				Runs []struct {
					Text string `xml:"t"`
				} `xml:"r"`
			} `xml:"p"`
		}
	}

	err = xml.Unmarshal(content, &document)
	if err != nil {
		return fmt.Errorf("failed to unmarshal document.xml: %w", err)
	}

	for _, paragraph := range document.Body.Paragraphs {
		for _, run := range paragraph.Runs {
			textContent.WriteString(run.Text)
		}
		textContent.WriteString("\n")
	}

	return nil
}

func (p *DOCXParser) extractMetadata(file *zip.File) error {
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open core.xml: %w", err)
	}
	defer rc.Close()

	content, err := ioutil.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("failed to read core.xml: %w", err)
	}

	var coreProps struct {
		Title          string `xml:"title"`
		Subject        string `xml:"subject"`
		Creator        string `xml:"creator"`
		Keywords       string `xml:"keywords"`
		Description    string `xml:"description"`
		LastModifiedBy string `xml:"lastModifiedBy"`
		Revision       string `xml:"revision"`
		Created        string `xml:"created"`
		Modified       string `xml:"modified"`
	}

	err = xml.Unmarshal(content, &coreProps)
	if err != nil {
		return fmt.Errorf("failed to unmarshal core.xml: %w", err)
	}

	p.metadata["format"] = "DOCX"
	if coreProps.Title != "" {
		p.metadata["title"] = coreProps.Title
	}
	if coreProps.Subject != "" {
		p.metadata["subject"] = coreProps.Subject
	}
	if coreProps.Creator != "" {
		p.metadata["creator"] = coreProps.Creator
	}
	if coreProps.Keywords != "" {
		p.metadata["keywords"] = coreProps.Keywords
	}
	if coreProps.Description != "" {
		p.metadata["description"] = coreProps.Description
	}
	if coreProps.LastModifiedBy != "" {
		p.metadata["lastModifiedBy"] = coreProps.LastModifiedBy
	}
	if coreProps.Revision != "" {
		p.metadata["revision"] = coreProps.Revision
	}
	if coreProps.Created != "" {
		p.metadata["created"] = coreProps.Created
	}
	if coreProps.Modified != "" {
		p.metadata["modified"] = coreProps.Modified
	}

	return nil
}

func (p *DOCXParser) GetFormatMetadata() map[string]string {
	return p.metadata
}
