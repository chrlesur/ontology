package prompt

import (
	"fmt"
	"strings"
)

// PromptTemplate représente un template de prompt
type PromptTemplate struct {
	Template string
}

// NewPromptTemplate crée un nouveau PromptTemplate
func NewPromptTemplate(template string) *PromptTemplate {
	return &PromptTemplate{Template: template}
}

// Format remplit le template avec les valeurs fournies
func (pt *PromptTemplate) Format(values map[string]string) string {
	result := pt.Template
	for key, value := range values {
		result = strings.Replace(result, fmt.Sprintf("{%s}", key), value, -1)
	}
	return result
}

// Définition des templates de prompts
var (
	EntityExtractionPrompt = NewPromptTemplate(`
Analyze the following text and extract key entities (e.g., people, organizations, concepts) relevant to building an ontology:

{text}

For each entity, provide:
1. Entity name
2. Entity type (e.g., Person, Organization, Concept)
3. A brief description or context

Format your response as a list of entities, one per line, using tabs to separate fields like this:
EntityName\tEntityType\tDescription/Context

Ensure that your extractions are relevant to creating an ontology and avoid including irrelevant or trivial information.
Use the original document language
`)

	RelationExtractionPrompt = NewPromptTemplate(`
Based on the following text and the list of entities provided, identify relationships between these entities that would be relevant for an ontology:

Text:
{text}

Entities:
{entities}

For each relationship, provide:
1. Source Entity
2. Relationship Type
3. Target Entity
4. A brief description or context of the relationship

Format your response as a list of relationships, one per line, using tabs to separate fields like this:
SourceEntity\tRelationshipType\tTargetEntity\tDescription/Context

Focus on meaningful relationships that contribute to the structure of the ontology. Avoid trivial or overly generic relationships.
Use the original document language
`)

	OntologyEnrichmentPrompt = NewPromptTemplate(`
Given the following partial ontology and new text, suggest additions or modifications to enrich the ontology:

Current Ontology:
{current_ontology}

New Text:
{text}

Provide your suggestions in the following format:
1. Type of Change (Add/Modify/Remove)
2. Entity or Relationship affected
3. Proposed change
4. Justification for the change

Ensure that your suggestions maintain the coherence and relevance of the ontology while incorporating new information from the text.
Use the original document language

`)
)
