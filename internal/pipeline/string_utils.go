// string_utils.go

package pipeline

import (
	"strings"
)

// truncateString tronque une chaîne à une longueur maximale donnée
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

// generateArticleVariants génère des variantes d'un mot avec différents articles
func generateArticleVariants(word string) []string {
	variants := []string{word}
	lowercaseWord := strings.ToLower(word)

	// Ajouter des variantes avec et sans apostrophe
	if !strings.HasPrefix(lowercaseWord, "l'") && !strings.HasPrefix(lowercaseWord, "d'") {
		variants = append(variants, "l'"+lowercaseWord, "d'"+lowercaseWord, "l "+lowercaseWord, "d "+lowercaseWord)
	}

	// Ajouter une variante sans underscore si le mot en contient
	if strings.Contains(word, "_") {
		spaceVariant := strings.ReplaceAll(word, "_", " ")
		variants = append(variants, spaceVariant)
		variants = append(variants, "l'"+spaceVariant, "d'"+spaceVariant, "l "+spaceVariant, "d "+spaceVariant)
	}

	return variants
}

func UniqueStringSlice(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// min retourne le minimum entre deux entiers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max retourne le maximum entre deux entiers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}
