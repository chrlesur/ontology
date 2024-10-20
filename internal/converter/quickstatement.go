// Package converter provides functionality for converting document segments
// into QuickStatement format and other ontology representations.
package converter

import (
	"github.com/chrlesur/Ontology/internal/logger"
)

// Converter defines the interface for QuickStatement conversion
type Converter interface {
	Convert(segment []byte, context string, ontology string) (string, error)
	ConvertToRDF(quickstatement string) (string, error)
	ConvertToOWL(quickstatement string) (string, error)
}

// QuickStatementConverter implements the Converter interface
type QuickStatementConverter struct {
	logger *logger.Logger
}

// NewQuickStatementConverter creates a new QuickStatementConverter
func NewQuickStatementConverter(log *logger.Logger) *QuickStatementConverter {
	return &QuickStatementConverter{
		logger: log,
	}
}

// Entity represents a Wikibase entity
type Entity struct {
	ID    string
	Label string
}

// Property represents a Wikibase property
type Property struct {
	ID       string
	DataType string
}

// Statement represents a complete QuickStatement
type Statement struct {
	Subject  Entity
	Property Property
	Object   interface{}
}
