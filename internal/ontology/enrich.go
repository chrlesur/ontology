package ontology

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/pipeline"
	"github.com/spf13/cobra"
)

var (
	output           string
	format           string
	llm              string
	llmModel         string
	passes           int
	rdf              bool
	owl              bool
	recursive        bool
	existingOntology string
)

// enrichCmd represents the enrich command
var enrichCmd = &cobra.Command{
	Use:   "enrich [input]",
	Short: i18n.Messages.EnrichCmdShortDesc,
	Long:  i18n.Messages.EnrichCmdLongDesc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]
		log.Info(i18n.Messages.StartingEnrichProcess)

		// Vérifier si l'input est un fichier ou un répertoire
		fileInfo, err := os.Stat(input)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.Messages.ErrAccessInput, err)
		}

		if fileInfo.IsDir() {
			return processDirectory(input)
		}
		return processFile(input)
	},
}

func init() {
	rootCmd.AddCommand(enrichCmd)

	enrichCmd.Flags().StringVar(&output, "output", "", i18n.Messages.OutputFlagUsage)
	enrichCmd.Flags().StringVar(&format, "format", "", i18n.Messages.FormatFlagUsage)
	enrichCmd.Flags().StringVar(&llm, "llm", "", i18n.Messages.LLMFlagUsage)
	enrichCmd.Flags().StringVar(&llmModel, "llm-model", "", i18n.Messages.LLMModelFlagUsage)
	enrichCmd.Flags().IntVar(&passes, "passes", 1, i18n.Messages.PassesFlagUsage)
	enrichCmd.Flags().BoolVar(&rdf, "rdf", false, i18n.Messages.RDFFlagUsage)
	enrichCmd.Flags().BoolVar(&owl, "owl", false, i18n.Messages.OWLFlagUsage)
	enrichCmd.Flags().BoolVar(&recursive, "recursive", false, i18n.Messages.RecursiveFlagUsage)
	enrichCmd.Flags().StringVar(&existingOntology, "existing-ontology", "", i18n.Messages.ExistingOntologyFlagUsage)
}

func processDirectory(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if !recursive && path != dirPath {
				return filepath.SkipDir
			}
			return nil
		}
		return processFile(path)
	})
}

func processFile(filePath string) error {
    outputPath := output
    if outputPath == "" {
        // Générer le nom de fichier de sortie dans le même répertoire que le fichier d'entrée
        dir := filepath.Dir(filePath)
        baseName := filepath.Base(filePath)
        baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
        
        var extension string
        if owl {
            extension = ".owl"
        } else if rdf {
            extension = ".rdf"
        } else {
            extension = ".tsv"
        }
        
        outputPath = filepath.Join(dir, baseName+extension)
    }

    p, err := pipeline.NewPipeline()
    if err != nil {
        return fmt.Errorf("%s: %w", i18n.Messages.ErrorCreatingPipeline, err)
    }

    // Passer le chemin de sortie à ExecutePipeline
    err = p.ExecutePipeline(filePath, outputPath, passes, existingOntology)
    if err != nil {
        return fmt.Errorf("%s: %w", i18n.Messages.ErrorExecutingPipeline, err)
    }

    log.Info(fmt.Sprintf("File processed: %s, output: %s", filePath, outputPath))
    return nil
}

func generateOutputFilename(input string, owl, rdf bool) string {
	baseName := filepath.Base(input)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))

	var extension string
	if owl {
		extension = ".owl"
	} else if rdf {
		extension = ".rdf"
	} else {
		extension = ".tsv"
	}

	return baseName + extension
}
