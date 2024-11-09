// string_utils.go

package pipeline

import (
    "strings"
    "unicode"
)

// truncateString tronque une chaîne à une longueur maximale donnée
func truncateString(s string, maxLength int) string {
    if len(s) <= maxLength {
        return s
    }
    return s[:maxLength] + "..."
}

// normalizeWord normalise un mot en le convertissant en minuscules et en ne gardant que les lettres, les chiffres et les apostrophes
func normalizeWord(word string) string {
    return strings.TrimSpace(strings.Map(func(r rune) rune {
        if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '\'' {
            return unicode.ToLower(r)
        }
        return ' '
    }, word))
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

// uniqueIntSlice supprime les doublons dans une slice d'entiers
func uniqueIntSlice(intSlice []int) []int {
    keys := make(map[int]bool)
    list := []int{}
    for _, entry := range intSlice {
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