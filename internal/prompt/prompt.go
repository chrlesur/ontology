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

// NewCustomPromptTemplate crée un nouveau PromptTemplate à partir d'une chaîne
func NewCustomPromptTemplate(template string) *PromptTemplate {
    return &PromptTemplate{Template: template}
}

// Format remplit le template avec les valeurs fournies
func (pt *PromptTemplate) Format(values map[string]string) string {
    result := pt.Template
    for key, value := range values {
        result = strings.Replace(result, fmt.Sprintf("{%s}", key), value, -1)
    }
    
    // Ajouter le prompt supplémentaire s'il existe
    if additionalPrompt, ok := values["additional_prompt"]; ok && additionalPrompt != "" {
        result += "\n\nAdditional instructions:\n" + additionalPrompt
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

Ensure that:
your extractions are relevant to creating an ontology and avoid including irrelevant or trivial information.
All spaces in entity names are replaced with underscores
Your extractions are relevant to creating an ontology
You avoid including irrelevant or trivial information
You use the original document language
You provide the output silently with no additional comments
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
Ensure that:

All spaces in entity names are replaced with underscores
Your extractions are relevant to creating an ontology
You avoid including irrelevant or trivial information
You use the original document language
You provide the output silently with no additional comments
`)

	OntologyEnrichmentPrompt = NewPromptTemplate(`
Vous êtes un expert en ontologies chargé d'enrichir et de raffiner une ontologie existante. Voici l'ontologie actuelle et de nouvelles informations à intégrer :

Ontologie actuelle :
{previous_result}

Nouveau texte à analyser :
{text}

Contexte supplémentaire :
{context}

Votre tâche :
1. Analyser le nouveau texte et le contexte.
2. Identifier les nouvelles entités et relations pertinentes.
3. Intégrer ces nouvelles informations dans l'ontologie existante.
4. Raffiner les entités et relations existantes si nécessaire.
5. Assurer la cohérence globale de l'ontologie.

Fournissez l'ontologie enrichie et raffinée dans le format suivant :
- Pour les entités : Nom_Entité\tType_Entité\tDescription
- Pour les relations : Entité_Source\tType_Relation\tEntité_Cible\tDescription

Assurez-vous que :

Tous les espaces dans les noms d'entités sont remplacés par des underscores.
Vos extractions sont pertinentes pour la création d'une ontologie
Vous évitez d'inclure des informations non pertinentes ou triviales
Vous utilisez la langue originale du document
Vous fournissez le résultat silencieusement sans commentaires supplémentaires
`)

	OntologyMergePrompt = NewPromptTemplate(`
Vous êtes un expert en fusion d'ontologies. Votre tâche est de fusionner intelligemment une ontologie existante avec de nouvelles informations pour créer une ontologie enrichie et cohérente.

Ontologie existante :
{previous_ontology}

Nouvelles informations à intégrer :
{new_ontology}

Directives pour la fusion :
1. Intégrez toutes les nouvelles entités et relations pertinentes de la nouvelle ontologie.
2. En cas de conflit ou de duplication, identifie les concepts qui sont essentiellement identiques ou très proches sémantiquement. Pour chaque groupe de concepts similaires, choisis le nom le plus approprié et représentatif.
Fusionne les descriptions en une seule, plus complète. Combine toutes les positions textuelles en une seule liste, sans doublons, triée par ordre croissant.
3. Assurez-vous que les relations entre les entités restent cohérentes.
4. Si une nouvelle information contredit une ancienne, privilégiez la nouvelle mais notez la contradiction si elle est significative.
5. Maintenez la structure et le format de l'ontologie existante.
6. Évitez les redondances et les informations en double.

Votre tâche :
- Analysez attentivement les deux ensembles d'informations.
- Fusionnez-les en une ontologie unique et cohérente.
- Assurez-vous que le résultat final est complet, sans perte d'information importante.

Format de sortie :
Présentez l'ontologie fusionnée dans le même format que l'ontologie existante, avec une entité ou une relation par ligne.
Pour les entités : Nom_Entité\tType_Entité\tDescription
Pour les relations : Entité_Source\tType_Relation\tEntité_Cible\tDescription

Procédez à la fusion de manière silencieuse, sans ajouter de commentaires ou d'explications supplémentaires.
`)
)
