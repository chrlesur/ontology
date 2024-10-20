package parser

import (
    "io/ioutil"
    "strings"

    "golang.org/x/net/html"
    "github.com/chrlesur/Ontology/internal/logger"
    "github.com/chrlesur/Ontology/internal/i18n"
)

func init() {
    RegisterParser(".html", NewHTMLParser)
}

// HTMLParser implémente l'interface Parser pour les fichiers HTML
type HTMLParser struct {
    metadata map[string]string
}

// NewHTMLParser crée une nouvelle instance de HTMLParser
func NewHTMLParser() Parser {
    return &HTMLParser{
        metadata: make(map[string]string),
    }
}

// Parse extrait le contenu textuel d'un fichier HTML
func (p *HTMLParser) Parse(path string) ([]byte, error) {
    logger.Debug(i18n.ParseStarted, "HTML", path)
    content, err := ioutil.ReadFile(path)
    if err != nil {
        logger.Error(i18n.ParseFailed, "HTML", path, err)
        return nil, err
    }

    doc, err := html.Parse(strings.NewReader(string(content)))
    if err != nil {
        logger.Error(i18n.ParseFailed, "HTML", path, err)
        return nil, err
    }

    var textContent strings.Builder
    var extractText func(*html.Node)
    extractText = func(n *html.Node) {
        if n.Type == html.TextNode {
            textContent.WriteString(n.Data)
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            extractText(c)
        }
    }
    extractText(doc)

    p.extractMetadata(doc)
    logger.Info(i18n.ParseCompleted, "HTML", path)
    return []byte(textContent.String()), nil
}

// GetMetadata retourne les métadonnées du fichier HTML
func (p *HTMLParser) GetMetadata() map[string]string {
    return p.metadata
}

// extractMetadata extrait les métadonnées du HTML
func (p *HTMLParser) extractMetadata(doc *html.Node) {
    var f func(*html.Node)
    f = func(n *html.Node) {
        if n.Type == html.ElementNode && n.Data == "meta" {
            var name, content string
            for _, a := range n.Attr {
                if a.Key == "name" {
                    name = a.Val
                } else if a.Key == "content" {
                    content = a.Val
                }
            }
            if name != "" && content != "" {
                p.metadata[name] = content
            }
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }
    f(doc)
}