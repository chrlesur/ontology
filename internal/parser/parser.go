package parser

import "io"

// Parser définit l'interface pour tous les analyseurs de documents
type Parser interface {
    // Parse prend le chemin d'un fichier et retourne son contenu en bytes
    Parse(path string) ([]byte, error)
    // GetMetadata retourne les métadonnées du document sous forme de map
    GetMetadata() map[string]string
}

// FormatParser est une fonction qui crée un Parser spécifique à un format
type FormatParser func() Parser

// formatParsers stocke les fonctions de création de Parser pour chaque format supporté
var formatParsers = make(map[string]FormatParser)

// RegisterParser enregistre un nouveau parser pour un format donné
func RegisterParser(format string, parser FormatParser) {
    formatParsers[format] = parser
}

// GetParser retourne le parser approprié basé sur le format spécifié
func GetParser(format string) (Parser, error) {
    parserFunc, ok := formatParsers[format]
    if !ok {
        return nil, fmt.Errorf("unsupported format: %s", format)
    }
    return parserFunc(), nil
}

// ParseDirectory parcourt un répertoire et parse tous les fichiers supportés
func ParseDirectory(path string, recursive bool) ([][]byte, error) {
    var results [][]byte
    err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            if !recursive && filePath != path {
                return filepath.SkipDir
            }
            return nil
        }
        ext := strings.ToLower(filepath.Ext(filePath))
        parser, err := GetParser(ext)
        if err != nil {
            return nil // Skip unsupported files
        }
        content, err := parser.Parse(filePath)
        if err != nil {
            logger.Warning("Failed to parse file: %s, error: %v", filePath, err)
            return nil
        }
        results = append(results, content)
        return nil
    })
    return results, err
}