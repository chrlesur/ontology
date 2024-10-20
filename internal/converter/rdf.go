package converter

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/i18n"
)

// ConvertToRDF converts a QuickStatement to RDF format
func (qsc *QuickStatementConverter) ConvertToRDF(quickstatement string) (string, error) {
	qsc.logger.Debug(i18n.GetMessage("ConvertToRDFStarted"))

	statements, err := qsc.parseQuickStatement(quickstatement)
	if err != nil {
		return "", fmt.Errorf("%s: %w", i18n.GetMessage("FailedToParseQuickStatement"), err)
	}

	var rdfStatements []string
	for _, stmt := range statements {
		rdfStmt, err := qsc.statementToRDF(stmt)
		if err != nil {
			return "", fmt.Errorf("%s: %w", i18n.GetMessage("FailedToConvertStatementToRDF"), err)
		}
		rdfStatements = append(rdfStatements, rdfStmt)
	}

	result := qsc.generateRDFDocument(rdfStatements)

	qsc.logger.Debug(i18n.GetMessage("ConvertToRDFFinished"))
	return result, nil
}

func (qsc *QuickStatementConverter) parseQuickStatementForRDF(quickstatement string) ([]Statement, error) {	qsc.logger.Debug(i18n.GetMessage("ParsingQuickStatement"))
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

func (qsc *QuickStatementConverter) statementToRDF(statement Statement) (string, error) {
	qsc.logger.Debug(i18n.GetMessage("ConvertingStatementToRDF"))
	subjectURI := fmt.Sprintf("<%s%s>", config.GetConfig().BaseURI, statement.Subject.ID)
	propertyURI := fmt.Sprintf("<%s%s>", config.GetConfig().BaseURI, statement.Property.ID)
	objectValue := statement.Object.(string) // Type assertion, be cautious in real implementation

	var objectRDF string
	if strings.HasPrefix(objectValue, "Q") {
		// If object is an entity
		objectRDF = fmt.Sprintf("<%s%s>", config.GetConfig().BaseURI, objectValue)
	} else {
		// If object is a literal
		objectRDF = fmt.Sprintf("\"%s\"", objectValue)
	}

	return fmt.Sprintf("%s %s %s .", subjectURI, propertyURI, objectRDF), nil
}

func (qsc *QuickStatementConverter) generateRDFDocument(rdfStatements []string) string {
	qsc.logger.Debug(i18n.GetMessage("GeneratingRDFDocument"))
	var buffer bytes.Buffer

	buffer.WriteString("@prefix wd: <" + config.GetConfig().BaseURI + "> .\n\n")
	for _, stmt := range rdfStatements {
		buffer.WriteString(stmt + "\n")
	}

	return buffer.String()
}
