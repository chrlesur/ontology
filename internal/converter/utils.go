package converter

import (
	"crypto/rand"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chrlesur/Ontology/internal/i18n"
)

// EscapeString escapes special characters in a string for safe use in output formats
func EscapeString(s string) string {
	return strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
		`/`, `\/`,
		"\b", `\b`,
		"\f", `\f`,
		"\n", `\n`,
		"\r", `\r`,
		"\t", `\t`,
	).Replace(s)
}

// IsValidURI checks if a string is a valid URI
func IsValidURI(uri string) bool {
	_, err := url.ParseRequestURI(uri)
	if err != nil {
		log.Debug(i18n.GetMessage("InvalidURI")) // Cette ligne est correcte car log.Debug n'attend qu'un seul argument
		return false
	}
	return true
}

// FormatDate formats a date string to a standard format (ISO 8601)
func FormatDate(date string) (string, error) {
	parsedDate, err := time.Parse(time.RFC3339, date)
	if err != nil {
		log.Warning(i18n.GetMessage("InvalidDateFormat"), err) // Ajout de l'erreur comme second argument
		return "", fmt.Errorf("%s: %w", i18n.GetMessage("ErrorParsingDate"), err)
	}
	return parsedDate.Format(time.RFC3339), nil
}

// GenerateUniqueID generates a unique identifier for entities without one
func GenerateUniqueID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Error(i18n.GetMessage("ErrorGeneratingUniqueID"), err)
		return ""
	}
	return fmt.Sprintf("%x", b)
}

// SplitIntoChunks divides a large dataset into smaller chunks for batch processing
func SplitIntoChunks(data []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}

// MergeMaps merges multiple maps into a single map
func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// TruncateString truncates a string to a specified length, adding an ellipsis if truncated
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}

// NormalizeWhitespace removes extra whitespace from a string
func NormalizeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// IsNumeric checks if a string contains only numeric characters
func IsNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
