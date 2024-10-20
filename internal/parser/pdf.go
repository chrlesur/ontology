package parser

import (
    "github.com/unidoc/unipdf/v3/model"
    "github.com/chrlesur/Ontology/internal/logger"
    "github.com/chrlesur/Ontology/internal/i18n"
)

func init() {
    RegisterParser(".pdf", NewPDFParser)
}

// PDFParser implémente l'interface Parser pour les fichiers PDF
type PDFParser struct {
    metadata map[string]string
}

// NewPDFParser crée une nouvelle instance de PDFParser
func NewPDFParser() Parser {
    return &PDFParser{
        metadata: make(map[string]string),
    }
}

// Parse extrait le contenu textuel d'un fichier PDF
func (p *PDFParser) Parse(path string) ([]byte, error) {
    logger.Debug(i18n.ParseStarted, "PDF", path)
    f, err := os.Open(path)
    if err != nil {
        logger.Error(i18n.ParseFailed, "PDF", path, err)
        return nil, err
    }
    defer f.Close()

    pdfReader, err := model.NewPdfReader(f)
    if err != nil {
        logger.Error(i18n.ParseFailed, "PDF", path, err)
        return nil, err
    }

    var content []byte
    numPages, err := pdfReader.GetNumPages()
    if err != nil {
        logger.Error(i18n.ParseFailed, "PDF", path, err)
        return nil, err
    }

    for i := 1; i <= numPages; i++ {
        page, err := pdfReader.GetPage(i)
        if err != nil {
            logger.Warning(i18n.PageParseFailed, i, path, err)
            continue
        }
        text, err := page.GetAllText()
        if err != nil {
            logger.Warning(i18n.TextExtractionFailed, i, path, err)
            continue
        }
        content = append(content, []byte(text)...)
    }

    p.extractMetadata(pdfReader)
    logger.Info(i18n.ParseCompleted, "PDF", path)
    return content, nil
}

// GetMetadata retourne les métadonnées du fichier PDF
func (p *PDFParser) GetMetadata() map[string]string {
    return p.metadata
}

// extractMetadata extrait les métadonnées du PDF
func (p *PDFParser) extractMetadata(pdfReader *model.PdfReader) {
    info, err := pdfReader.GetInfo()
    if err != nil {
        logger.Warning(i18n.MetadataExtractionFailed, "PDF", err)
        return
    }
    p.metadata["title"] = info.Title.String()
    p.metadata["author"] = info.Author.String()
    p.metadata["subject"] = info.Subject.String()
    p.metadata["keywords"] = info.Keywords.String()
    p.metadata["creator"] = info.Creator.String()
    p.metadata["producer"] = info.Producer.String()
}