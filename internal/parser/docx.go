package parser

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
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
	log.Debug("Parse started for DOCX")

	// Lire tout le contenu en mémoire
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Error("Failed to read DOCX content: %v", err)
		return nil, fmt.Errorf("failed to read DOCX content: %w", err)
	}
	log.Debug("DOCX content read, size: %d bytes", len(content))

	// Créer un lecteur zip à partir du contenu
	zipReader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		log.Error("Failed to create zip reader: %v", err)
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
	}
	log.Debug("Zip reader created successfully")

	var textContent strings.Builder

	for _, file := range zipReader.File {
		log.Debug("Processing zip file: %s", file.Name)
		switch file.Name {
		case "word/document.xml":
			log.Debug("Found word/document.xml, extracting content")
			if err := p.extractContent(file, &textContent); err != nil {
				log.Error("Failed to extract content from word/document.xml: %v", err)
				return nil, err
			}
		case "docProps/core.xml":
			log.Debug("Found docProps/core.xml, extracting metadata")
			if err := p.extractMetadata(file); err != nil {
				log.Warning("Failed to extract metadata from docProps/core.xml: %v", err)
			}
		}
	}

	result := textContent.String()
	log.Debug("Extracted text content length: %d characters", len(result))
	log.Debug("Contenu extrait (premiers 100 caractères) : %s", string(result[:min(100, len(result))]))
	log.Debug("Contenu extrait (derniers 100 caractères) : %s", string(result[max(0, len(result)-100):]))
	log.Debug("Longueur totale du contenu extrait : %d", len(result))
	log.Info("DOCX parsing completed")
	return []byte(result), nil
}

func (p *DOCXParser) extractContent(file *zip.File, textContent *strings.Builder) error {
	log.Debug("Extracting content from: %s", file.Name)

	rc, err := file.Open()
	if err != nil {
		log.Error("Failed to open document.xml: %v", err)
		return fmt.Errorf("failed to open document.xml: %w", err)
	}
	defer rc.Close()

	content, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Error("Failed to read document.xml: %v", err)
		return fmt.Errorf("failed to read document.xml: %w", err)
	}
	log.Debug("Read %d bytes from document.xml", len(content))

	if len(content) > 1000 {
		log.Debug("First 1000 characters of document.xml: %s", string(content[:1000]))
	} else {
		log.Debug("Full content of document.xml: %s", string(content))
	}

	decoder := xml.NewDecoder(bytes.NewReader(content))
	var inTextElement bool
	var currentText string

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("Error decoding XML: %v", err)
			return err
		}

		switch se := token.(type) {
		case xml.StartElement:
			if se.Name.Local == "t" {
				inTextElement = true
			}
		case xml.EndElement:
			if se.Name.Local == "t" {
				inTextElement = false
				textContent.WriteString(currentText)
				textContent.WriteString(" ")
				currentText = ""
			} else if se.Name.Local == "p" {
				textContent.WriteString("\n")
			}
		case xml.CharData:
			if inTextElement {
				currentText += string(se)
			}
		}
	}

	log.Debug("Total extracted content length: %d", textContent.Len())
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
