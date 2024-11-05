package parser

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/chrlesur/Ontology/internal/i18n"
	"golang.org/x/net/html"
)

type HTMLParser struct {
	metadata map[string]string
}

func init() {
	RegisterParser(".html", NewPDFParser)
}

func NewHTMLParser() Parser {
	return &HTMLParser{
		metadata: make(map[string]string),
	}
}

func (p *HTMLParser) Parse(reader io.Reader) ([]byte, error) {
	log.Debug(i18n.Messages.ParseStarted, "HTML")

	content, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Error(i18n.Messages.ParseFailed, "HTML", err)
		return nil, fmt.Errorf("failed to read HTML content: %w", err)
	}

	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		log.Error(i18n.Messages.ParseFailed, "HTML", err)
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var textContent strings.Builder
	p.extractContent(doc, &textContent)
	p.extractMetadata(doc)

	log.Info(i18n.Messages.ParseCompleted, "HTML")
	return []byte(textContent.String()), nil
}

func (p *HTMLParser) extractContent(n *html.Node, textContent *strings.Builder) {
	if n.Type == html.TextNode {
		textContent.WriteString(strings.TrimSpace(n.Data))
		textContent.WriteString(" ")
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.extractContent(c, textContent)
	}
}

func (p *HTMLParser) extractMetadata(n *html.Node) {
	p.metadata["format"] = "HTML"

	var extractMetaTag func(*html.Node)
	extractMetaTag = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "name", "property":
					name = attr.Val
				case "content":
					content = attr.Val
				}
			}
			if name != "" && content != "" {
				p.metadata[name] = content
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractMetaTag(c)
		}
	}

	var extractTitle func(*html.Node)
	extractTitle = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				p.metadata["title"] = n.FirstChild.Data
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractTitle(c)
		}
	}

	extractMetaTag(n)
	extractTitle(n)
}

func (p *HTMLParser) GetFormatMetadata() map[string]string {
	return p.metadata
}
