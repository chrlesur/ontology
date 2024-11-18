package pipeline

import (
	"github.com/chrlesur/Ontology/internal/logger"
)

var log = logger.GetLogger()

// ContextEntry représente le contexte pour une position spécifique dans le document
type ContextEntry struct {
    Position       int      `json:"position"`
    FileID         string   `json:"file_id"`
    FilePosition   int      `json:"file_position"`
    Before         []string `json:"before"`
    After          []string `json:"after"`
    Element        string   `json:"element"`
    Length         int      `json:"length"`
}
