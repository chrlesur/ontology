package converter

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/chrlesur/Ontology/internal/i18n"
)


// Convert converts a document segment into QuickStatement format
func (qsc *QuickStatementConverter) Convert(segment []byte, context string, ontology string) (string, error) {
	log.Debug(i18n.GetMessage("ConvertStarted"))

	statements, err := qsc.parseSegment(segment)
	if err != nil {
		log.Error(i18n.GetMessage("FailedToParseSegment"), err)
		return "", fmt.Errorf("%s: %w", i18n.GetMessage("FailedToParseSegment"), err)
	}

	enrichedStatements, err := qsc.applyContextAndOntology(statements, context, ontology)
	if err != nil {
		log.Error(i18n.GetMessage("FailedToApplyContextAndOntology"), err)
		return "", fmt.Errorf("%s: %w", i18n.GetMessage("FailedToApplyContextAndOntology"), err)
	}

	result, err := qsc.toQuickStatementTSV(enrichedStatements)
	if err != nil {
		log.Error(i18n.GetMessage("FailedToConvertToQuickStatementTSV"), err)
		return "", fmt.Errorf("%s: %w", i18n.GetMessage("FailedToConvertToQuickStatementTSV"), err)
	}

	log.Debug(i18n.GetMessage("ConvertFinished"))
	return result, nil
}

func (qsc *QuickStatementConverter) parseSegment(segment []byte) ([]Statement, error) {
	log.Debug(i18n.GetMessage("ParsingSegment"))
	var statements []Statement
	scanner := bufio.NewScanner(bytes.NewReader(segment))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			return nil, fmt.Errorf(i18n.GetMessage("InvalidSegmentFormat"))
		}
		statement := Statement{
			Subject:  Entity{ID: parts[0]},
			Property: Property{ID: parts[1]},
			Object:   parts[2],
		}
		statements = append(statements, statement)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrorScanningSegment"), err)
	}
	return statements, nil
}

func (qsc *QuickStatementConverter) applyContextAndOntology(statements []Statement, context string, ontology string) ([]Statement, error) {
	log.Debug(i18n.GetMessage("ApplyingContextAndOntology"))
	// This is a placeholder implementation. In a real-world scenario, this function would
	// use the context and ontology to enrich the statements.
	for i := range statements {
		statements[i].Subject.Label = fmt.Sprintf("%s (from context)", statements[i].Subject.ID)
	}
	return statements, nil
}

func (qsc *QuickStatementConverter) toQuickStatementTSV(statements []Statement) (string, error) {
	log.Debug(i18n.GetMessage("ConvertingToQuickStatementTSV"))
	var buffer bytes.Buffer
	for _, stmt := range statements {
		line := fmt.Sprintf("%s\t%s\t%v\n", stmt.Subject.ID, stmt.Property.ID, stmt.Object)
		_, err := buffer.WriteString(line)
		if err != nil {
			return "", fmt.Errorf("%s: %w", i18n.GetMessage("ErrorWritingQuickStatement"), err)
		}
	}
	return buffer.String(), nil
}
