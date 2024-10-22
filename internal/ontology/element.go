package ontology

import (
	"fmt"
	"strconv"
	"strings"
)

// OntologyElement représente un élément unique dans l'ontologie
type OntologyElement struct {
	Name      string
	Type      string
	Positions []int
}

// NewOntologyElement crée un nouvel OntologyElement
func NewOntologyElement(name, elementType string) *OntologyElement {
	return &OntologyElement{
		Name:      name,
		Type:      elementType,
		Positions: []int{},
	}
}

// AddPosition ajoute une nouvelle position à l'élément
func (e *OntologyElement) AddPosition(position int) {
	e.Positions = append(e.Positions, position)
}

// SetPositions remplace toutes les positions existantes par un nouveau slice
func (e *OntologyElement) SetPositions(positions []int) {
	e.Positions = positions
}

// String retourne une représentation en chaîne de caractères de l'élément
func (e *OntologyElement) String() string {
	posStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(e.Positions)), ","), "[]")
	return fmt.Sprintf("%s|%s@%s", e.Name, e.Type, posStr)
}

// Ontology représente une collection d'éléments d'ontologie
type Ontology struct {
	Elements []*OntologyElement
}

// NewOntology crée une nouvelle instance d'Ontology
func NewOntology() *Ontology {
	return &Ontology{
		Elements: []*OntologyElement{},
	}
}

// AddElement ajoute un nouvel élément à l'ontologie ou met à jour un élément existant
func (o *Ontology) AddElement(element *OntologyElement) {
	existingElement := o.GetElementByName(element.Name)
	if existingElement != nil {
		// Mettre à jour l'élément existant
		existingElement.Type = element.Type
		existingElement.Positions = append(existingElement.Positions, element.Positions...)
		// Supprimer les doublons dans les positions
		existingElement.Positions = removeDuplicates(existingElement.Positions)
	} else {
		// Ajouter un nouvel élément
		o.Elements = append(o.Elements, element)
	}
}

// GetElementByName recherche un élément par son nom
func (o *Ontology) GetElementByName(name string) *OntologyElement {
	for _, element := range o.Elements {
		if element.Name == name {
			return element
		}
	}
	return nil
}

// LoadFromString charge une ontologie à partir d'une chaîne de caractères au format QuickStatement
func (o *Ontology) LoadFromString(content string) error {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			continue // Ignorer les lignes mal formatées
		}
		name := parts[0]
		typeParts := strings.Split(parts[1], "@")
		if len(typeParts) != 2 {
			continue // Ignorer les éléments sans positions
		}
		elementType := typeParts[0]
		positionsStr := strings.Split(typeParts[1], ",")

		element := NewOntologyElement(name, elementType)
		for _, posStr := range positionsStr {
			pos, err := strconv.Atoi(posStr)
			if err != nil {
				continue // Ignorer les positions non valides
			}
			element.AddPosition(pos)
		}
		o.AddElement(element)
	}
	return nil
}

// ToString convertit l'ontologie en chaîne de caractères au format QuickStatement
func (o *Ontology) ToString() string {
	var builder strings.Builder
	for _, element := range o.Elements {
		posStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(element.Positions)), ","), "[]")
		builder.WriteString(fmt.Sprintf("%s|%s@%s\n", element.Name, element.Type, posStr))
	}
	return builder.String()
}

// removeDuplicates supprime les doublons dans un slice d'entiers
func removeDuplicates(slice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
