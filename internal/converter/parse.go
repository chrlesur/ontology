package converter

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/knakk/rdf"
)

// ParseOntology parses an ontology string and returns a structured representation
func ParseOntology(ontology string) (map[string]interface{}, error) {
	log.Debug(i18n.GetMessage("ParseOntologyStarted"))

	format := detectOntologyFormat(ontology)
	var result map[string]interface{}
	var err error

	switch format {
	case "QuickStatement":
		result, err = parseQuickStatementOntology(ontology)
	case "RDF":
		result, err = parseRDFOntology(ontology)
	case "OWL":
		result, err = parseOWLOntology(ontology)
	default:
		return nil, fmt.Errorf(i18n.GetMessage("UnknownOntologyFormat"))
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("FailedToParseOntology"), err)
	}

	log.Debug(i18n.GetMessage("ParseOntologyFinished"))
	return result, nil
}

func detectOntologyFormat(ontology string) string {
	if strings.Contains(ontology, "Q") && strings.Contains(ontology, "P") && strings.Contains(ontology, "\t") {
		return "QuickStatement"
	}
	if strings.Contains(ontology, "<rdf:RDF") {
		return "RDF"
	}
	if strings.Contains(ontology, "<Ontology") {
		return "OWL"
	}
	return "Unknown"
}

func parseQuickStatementOntology(ontology string) (map[string]interface{}, error) {
	log.Debug(i18n.GetMessage("ParseQuickStatementOntologyStarted"))

	result := make(map[string]interface{})
	entities := make(map[string]map[string]interface{})

	lines := strings.Split(ontology, "\n")
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			return nil, fmt.Errorf(i18n.GetMessage("InvalidQuickStatementLine"))
		}

		subject, predicate, object := parts[0], parts[1], parts[2]
		if _, exists := entities[subject]; !exists {
			entities[subject] = make(map[string]interface{})
		}
		entities[subject][predicate] = object
	}

	result["entities"] = entities
	log.Debug(i18n.GetMessage("ParseQuickStatementOntologyFinished"))
	return result, nil
}

func parseRDFOntology(ontology string) (map[string]interface{}, error) {
	log.Debug(i18n.GetMessage("ParseRDFOntologyStarted"))

	result := make(map[string]interface{})
	entities := make(map[string]map[string]interface{})

	dec := rdf.NewTripleDecoder(strings.NewReader(ontology), rdf.RDFXML)
	for {
		triple, err := dec.Decode()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrorDecodingRDF"), err)
		}

		subject := triple.Subj.String()
		predicate := triple.Pred.String()
		object := triple.Obj.String()

		if _, exists := entities[subject]; !exists {
			entities[subject] = make(map[string]interface{})
		}
		entities[subject][predicate] = object
	}

	result["entities"] = entities
	log.Debug(i18n.GetMessage("ParseRDFOntologyFinished"))
	return result, nil
}

func parseOWLOntology(ontology string) (map[string]interface{}, error) {
	log.Debug(i18n.GetMessage("ParseOWLOntologyStarted"))

	result := make(map[string]interface{})

	// Simplified OWL parsing
	var owlData struct {
		XMLName xml.Name `xml:"Ontology"`
		Classes []struct {
			IRI   string `xml:"IRI,attr"`
			Label string `xml:"label,attr"`
		} `xml:"Declaration>Class"`
	}

	err := xml.Unmarshal([]byte(ontology), &owlData)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrorParsingOWL"), err)
	}

	classes := make(map[string]interface{})
	for _, class := range owlData.Classes {
		classes[class.IRI] = map[string]interface{}{
			"label": class.Label,
		}
	}
	result["classes"] = classes

	log.Debug(i18n.GetMessage("ParseOWLOntologyFinished"))
	return result, nil
}
