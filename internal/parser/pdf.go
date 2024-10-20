package parser

import (
	"os"
	"strings"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

// ParsePDF reads a PDF file and returns its content as a byte slice.
func ParsePDF(path string) ([]byte, error) {
	log.Debug(i18n.ParseStarted)

	f, err := os.Open(path)
	if err != nil {
		log.Error(i18n.ParseFailed, err)
		return nil, err
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		log.Error(i18n.ParseFailed, err)
		return nil, err
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		log.Error(i18n.ParseFailed, err)
		return nil, err
	}

	var content strings.Builder
	for pageNum := 0; pageNum < numPages; pageNum++ {
		page, err := pdfReader.GetPage(pageNum + 1)
		if err != nil {
			log.Error(i18n.PageParseFailed, err)
			continue
		}

		ex, err := extractor.New(page)
		if err != nil {
			log.Error(i18n.TextExtractionFailed, err)
			continue
		}

		text, err := ex.ExtractText()
		if err != nil {
			log.Error(i18n.TextExtractionFailed, err)
			continue
		}

		content.WriteString(text)
	}

	log.Info(i18n.ParseCompleted)
	return []byte(content.String()), nil
}

// GetPDFMetadata extracts metadata from a PDF file.
func GetPDFMetadata(pdfReader *model.PdfReader) map[string]string {
	metadata := make(map[string]string)

	trailer, err := pdfReader.GetTrailer()
	if err != nil {
		log.Error(i18n.MetadataExtractionFailed, err)
		return metadata
	}

	if trailer.Get("Info") != nil {
		infoDict, ok := trailer.Get("Info").(*core.PdfObjectDictionary)
		if ok {
			for _, key := range infoDict.Keys() {
				val := infoDict.Get(key)
				if strObj, isStr := val.(*core.PdfObjectString); isStr {
					metadata[string(key)] = strObj.String()
				}
			}
		}
	}

	return metadata
}
