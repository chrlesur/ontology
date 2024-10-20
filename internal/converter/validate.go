package converter

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/chrlesur/Ontology/internal/i18n"
)

// ValidateQuickStatement validates a QuickStatement string
func ValidateQuickStatement(statement string) bool {
	log.Debug(i18n.GetMessage("ValidateQuickStatementStarted"))

	err := validateQuickStatementSyntax(statement)
	if err != nil {
		log.Warning(i18n.GetMessage("InvalidQuickStatementSyntax"), err)
		return false
	}

	err = checkEntityReferences(statement)
	if err != nil {
		log.Warning(i18n.GetMessage("InvalidEntityReferences"), err)
		return false
	}

	log.Debug(i18n.GetMessage("ValidateQuickStatementFinished"))
	return true
}

// ValidateRDF validates an RDF string
func ValidateRDF(rdf string) bool {
	log.Debug(i18n.GetMessage("ValidateRDFStarted"))

	err := validateRDFSyntax(rdf)
	if err != nil {
		log.Warning(i18n.GetMessage("InvalidRDFSyntax"), err)
		return false
	}

	log.Debug(i18n.GetMessage("ValidateRDFFinished"))
	return true
}

// ValidateOWL validates an OWL string
func ValidateOWL(owl string) bool {
	log.Debug(i18n.GetMessage("ValidateOWLStarted"))

	err := validateOWLSyntax(owl)
	if err != nil {
		log.Warning(i18n.GetMessage("InvalidOWLSyntax"), err)
		return false
	}

	log.Debug(i18n.GetMessage("ValidateOWLFinished"))
	return true
}

func validateQuickStatementSyntax(statement string) error {
	scanner := bufio.NewScanner(strings.NewReader(statement))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			return fmt.Errorf(i18n.GetMessage("InvalidQuickStatementLine"), lineNum)
		}
		if !isValidEntity(parts[0]) || !isValidProperty(parts[1]) {
			return fmt.Errorf(i18n.GetMessage("InvalidEntityOrProperty"), lineNum)
		}
	}
	return scanner.Err()
}

func validateRDFSyntax(rdf string) error {
	if !strings.Contains(rdf, "<rdf:RDF") || !strings.Contains(rdf, "</rdf:RDF>") {
		return fmt.Errorf(i18n.GetMessage("MissingRDFTags"))
	}
	if !strings.Contains(rdf, "xmlns:rdf=\"http://www.w3.org/1999/02/22-rdf-syntax-ns#\"") {
		return fmt.Errorf(i18n.GetMessage("MissingRDFNamespace"))
	}
	return nil
}

func validateOWLSyntax(owl string) error {
	if !strings.Contains(owl, "<owl:Ontology") || !strings.Contains(owl, "</owl:Ontology>") {
		return fmt.Errorf(i18n.GetMessage("MissingOWLTags"))
	}
	if !strings.Contains(owl, "xmlns:owl=\"http://www.w3.org/2002/07/owl#\"") {
		return fmt.Errorf(i18n.GetMessage("MissingOWLNamespace"))
	}
	return nil
}

func checkEntityReferences(statement string) error {
	scanner := bufio.NewScanner(strings.NewReader(statement))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if !isValidEntity(parts[0]) {
			return fmt.Errorf(i18n.GetMessage("InvalidEntityReference"), parts[0], lineNum)
		}
	}
	return scanner.Err()
}

func isValidEntity(entity string) bool {
	return regexp.MustCompile(`^Q\d+$`).MatchString(entity)
}

func isValidProperty(property string) bool {
	return regexp.MustCompile(`^P\d+$`).MatchString(property)
}

func isValidURI(uri string) bool {
	// This is a simplified URI validation. In a real-world scenario, you might want to use a more comprehensive validation.
	return regexp.MustCompile(`^(http|https)://`).MatchString(uri)
}
