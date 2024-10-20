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
	log.Debug("Input segment:\n%s", string(segment))

	statements, err := qsc.parseSegment(segment)
	if err != nil {
		log.Error(i18n.GetMessage("FailedToParseSegment"), err)
		return "", fmt.Errorf("%s: %w", i18n.GetMessage("FailedToParseSegment"), err)
	}

	log.Debug("Parsed %d statements", len(statements))

	var tsvBuilder strings.Builder
	for _, stmt := range statements {
		// Écrire la ligne TSV
		tsvLine := fmt.Sprintf("%s\t%s\t%s\n", stmt.Subject.ID, stmt.Property.ID, stmt.Object)
		tsvBuilder.WriteString(tsvLine)
	}

	result := tsvBuilder.String()
	log.Debug("Generated TSV output:\n%s", result)
	return result, nil
}

func (qsc *QuickStatementConverter) cleanAndNormalizeInput(input string) string {
	lines := strings.Split(input, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			cleanedLines = append(cleanedLines, trimmedLine)
		}
	}
	return strings.Join(cleanedLines, "\n")
}

func (qsc *QuickStatementConverter) parseSegment(segment []byte) ([]Statement, error) {
	log.Debug(i18n.GetMessage("ParsingSegment"))
	var statements []Statement
	scanner := bufio.NewScanner(bytes.NewReader(segment))
	for scanner.Scan() {
		line := scanner.Text()
		// Remplacer les doubles backslashes par un caractère temporaire
		line = strings.ReplaceAll(line, "\\\\", "\uFFFD")
		// Remplacer les \t par des tabulations réelles
		line = strings.ReplaceAll(line, "\\t", "\t")
		// Restaurer les doubles backslashes
		line = strings.ReplaceAll(line, "\uFFFD", "\\")

		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			log.Warning("Skipping invalid line: %s", line)
			continue
		}
		statement := Statement{
			Subject:  Entity{ID: strings.TrimSpace(parts[0])},
			Property: Property{ID: strings.TrimSpace(parts[1])},
			Object:   strings.TrimSpace(strings.Join(parts[2:], "\t")),
		}
		statements = append(statements, statement)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.GetMessage("ErrorScanningSegment"), err)
	}
	if len(statements) == 0 {
		return nil, fmt.Errorf(i18n.GetMessage("NoValidStatementsFound"))
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
