package parser

import (
	"archive/zip"
	"encoding/xml"
	"io/ioutil"
	"strings"

	"github.com/chrlesur/Ontology/internal/i18n"
)

func init() {
	RegisterParser(".docx", NewDOCXParser)
}

// DOCXParser implémente l'interface Parser pour les fichiers DOCX
type DOCXParser struct {
	metadata map[string]string
}

// NewDOCXParser crée une nouvelle instance de DOCXParser
func NewDOCXParser() Parser {
	return &DOCXParser{
		metadata: make(map[string]string),
	}
}

// Parse extrait le contenu textuel d'un fichier DOCX
func (p *DOCXParser) Parse(path string) ([]byte, error) {
	log.Debug(i18n.ParseStarted, "DOCX", path)

	reader, err := zip.OpenReader(path)
	if err != nil {
		log.Error(i18n.ParseFailed, "DOCX", path, err)
		return nil, err
	}
	defer reader.Close()

	var textContent strings.Builder
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			if err != nil {
				log.Error(i18n.ParseFailed, "DOCX", path, err)
				return nil, err
			}
			defer rc.Close()

			content, err := ioutil.ReadAll(rc)
			if err != nil {
				log.Error(i18n.ParseFailed, "DOCX", path, err)
				return nil, err
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
				log.Error(i18n.ParseFailed, "DOCX", path, err)
				return nil, err
			}

			for _, paragraph := range document.Body.Paragraphs {
				for _, run := range paragraph.Runs {
					textContent.WriteString(run.Text)
				}
				textContent.WriteString("\n")
			}
			break
		}
	}

	p.extractMetadata(reader)
	log.Info(i18n.ParseCompleted, "DOCX", path)
	return []byte(textContent.String()), nil
}

// GetMetadata retourne les métadonnées du fichier DOCX
func (p *DOCXParser) GetMetadata() map[string]string {
	return p.metadata
}

// extractMetadata extrait les métadonnées du DOCX
func (p *DOCXParser) extractMetadata(reader *zip.ReadCloser) {
	for _, file := range reader.File {
		if file.Name == "docProps/core.xml" {
			rc, err := file.Open()
			if err != nil {
				log.Warning(i18n.MetadataExtractionFailed, "DOCX", err)
				return
			}
			defer rc.Close()

			content, err := ioutil.ReadAll(rc)
			if err != nil {
				log.Warning(i18n.MetadataExtractionFailed, "DOCX", err)
				return
			}

			var coreProps struct {
				Title          string `xml:"title"`
				Subject        string `xml:"subject"`
				Creator        string `xml:"creator"`
				Keywords       string `xml:"keywords"`
				Description    string `xml:"description"`
				LastModifiedBy string `xml:"lastModifiedBy"`
				Revision       string `xml:"revision"`
			}

			err = xml.Unmarshal(content, &coreProps)
			if err != nil {
				log.Warning(i18n.MetadataExtractionFailed, "DOCX", err)
				return
			}

			p.metadata["title"] = coreProps.Title
			p.metadata["subject"] = coreProps.Subject
			p.metadata["creator"] = coreProps.Creator
			p.metadata["keywords"] = coreProps.Keywords
			p.metadata["description"] = coreProps.Description
			p.metadata["lastModifiedBy"] = coreProps.LastModifiedBy
			p.metadata["revision"] = coreProps.Revision
		}
	}
}
