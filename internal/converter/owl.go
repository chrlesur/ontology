package converter

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/i18n"
)

// ConvertToOWL converts a QuickStatement to OWL format
func (qsc *QuickStatementConverter) ConvertToOWL(quickstatement string) (string, error) {
	qsc.logger.Debug(i18n.GetMessage("ConvertToOWLStarted"))

	statements, err := qsc.parseQuickStatement(quickstatement)
	if err != nil {
		return "", fmt.Errorf("%s: %w", i18n.GetMessage("FailedToParseQuickStatement"), err)
	}

	var owlStatements []string
	for _, stmt := range statements {
		owlStmt, err := qsc.statementToOWL(stmt)
		if err != nil {
			return "", fmt.Errorf("%s: %w", i18n.GetMessage("FailedToConvertStatementToOWL"), err)
		}
		owlStatements = append(owlStatements, owlStmt)
	}

	result := qsc.generateOWLDocument(owlStatements)

	qsc.logger.Debug(i18n.GetMessage("ConvertToOWLFinished"))
	return result, nil
}

func (qsc *QuickStatementConverter) parseQuickStatement(quickstatement string) ([]Statement, error) {
	qsc.logger.Debug(i18n.GetMessage("ParsingQuickStatement"))
	var statements []Statement
	scanner := bufio.NewScanner(strings.NewReader(quickstatement))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			return nil, fmt.Errorf(i18n.GetMessage("InvalidQuickStatementFormat"))
		}
		statement := Statement{
			Subject:  Entity{ID: parts[0]},
			Property: Property{ID: parts[1]},
			Object:   parts[2],
		}
		statements = append(statements, statement)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrorScanningQuickStatement"), err)
	}
	return statements, nil
}

func (qsc *QuickStatementConverter) statementToOWL(statement Statement) (string, error) {
	qsc.logger.Debug(i18n.GetMessage("ConvertingStatementToOWL"))
	subjectURI := fmt.Sprintf(":%s", statement.Subject.ID)
	propertyURI := fmt.Sprintf(":%s", statement.Property.ID)
	objectValue := statement.Object.(string) // Type assertion, be cautious in real implementation

	var owlStatement string
	if strings.HasPrefix(objectValue, "Q") {
		// If object is an entity
		objectURI := fmt.Sprintf(":%s", objectValue)
		owlStatement = fmt.Sprintf("ObjectPropertyAssertion(%s %s %s)", propertyURI, subjectURI, objectURI)
	} else {
		// If object is a literal
		owlStatement = fmt.Sprintf("DataPropertyAssertion(%s %s \"%s\"^^xsd:string)", propertyURI, subjectURI, objectValue)
	}

	return owlStatement, nil
}

func (qsc *QuickStatementConverter) generateOWLDocument(owlStatements []string) string {
	qsc.logger.Debug(i18n.GetMessage("GeneratingOWLDocument"))
	var buffer bytes.Buffer

	buffer.WriteString("Prefix(:=<" + config.GetConfig().BaseURI + ">)\n")
	buffer.WriteString("Prefix(owl:=<http://www.w3.org/2002/07/owl#>)\n")
	buffer.WriteString("Prefix(rdf:=<http://www.w3.org/1999/02/22-rdf-syntax-ns#>)\n")
	buffer.WriteString("Prefix(xml:=<http://www.w3.org/XML/1998/namespace>)\n")
	buffer.WriteString("Prefix(xsd:=<http://www.w3.org/2001/XMLSchema#>)\n")
	buffer.WriteString("Prefix(rdfs:=<http://www.w3.org/2000/01/rdf-schema#>)\n\n")
	buffer.WriteString("Ontology(<" + config.GetConfig().BaseURI + ">\n\n")

	for _, stmt := range owlStatements {
		buffer.WriteString(stmt + "\n")
	}

	buffer.WriteString(")")

	return buffer.String()
}
