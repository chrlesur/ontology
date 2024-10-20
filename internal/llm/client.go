package llm

import (
	"github.com/chrlesur/Ontology/internal/prompt"
)

// Client defines the interface for LLM clients
type Client interface {
	// Translate takes a prompt and context, and returns the LLM's response
	Translate(prompt string, context string) (string, error)
	ProcessWithPrompt(promptTemplate *prompt.PromptTemplate, values map[string]string) (string, error)
}
