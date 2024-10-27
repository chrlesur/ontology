package model

// OntologyElement représente un élément unique dans l'ontologie
type OntologyElement struct {
	Name      string // Nom de l'élément
	Type      string // Type de l'élément
	Positions []int  // Positions de l'élément dans le document source
	Description string 

}

// Relation représente une relation entre deux éléments de l'ontologie
type Relation struct {
	Source      string
	Type        string
	Target      string
	Description string
}

// Ontology représente une collection d'éléments d'ontologie
type Ontology struct {
	Elements  []*OntologyElement // Liste des éléments de l'ontologie
	Relations []*Relation
}

// SetPositions définit les positions de l'élément
func (e *OntologyElement) SetPositions(positions []int) {
    e.Positions = positions
}

// NewOntologyElement crée un nouvel élément d'ontologie
func NewOntologyElement(name, elementType string) *OntologyElement {
	return &OntologyElement{
		Name:      name,
		Type:      elementType,
		Positions: []int{}, // Initialise un slice vide pour les positions
	}
}

// NewOntology crée une nouvelle instance d'Ontology
func NewOntology() *Ontology {
	return &Ontology{
		Elements: []*OntologyElement{}, // Initialise un slice vide pour les éléments
	}
}

// AddElement ajoute un nouvel élément à l'ontologie
func (o *Ontology) AddElement(element *OntologyElement) {
	o.Elements = append(o.Elements, element)
}

// GetElementByName recherche un élément par son nom
func (o *Ontology) GetElementByName(name string) *OntologyElement {
	for _, element := range o.Elements {
		if element.Name == name {
			return element
		}
	}
	return nil // Retourne nil si l'élément n'est pas trouvé
}

// AddRelation ajoute une nouvelle relation à l'ontologie
func (o *Ontology) AddRelation(relation *Relation) {
	o.Relations = append(o.Relations, relation)
}
