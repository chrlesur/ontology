package ontology

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/logger"
	"github.com/chrlesur/Ontology/internal/model"
	"github.com/chrlesur/Ontology/internal/pipeline"

	"github.com/spf13/cobra"
)

var (
	output                   string
	format                   string
	llm                      string
	llmModel                 string
	passes                   int
	recursive                bool
	existingOntology         string
	entityExtractionPrompt   string
	relationExtractionPrompt string
	ontologyEnrichmentPrompt string
	ontologyMergePrompt      string
	maxThreads               int
	aiyouAssistantID         string
)

// enrichCmd represents the enrich command
var enrichCmd = &cobra.Command{
	Use:   "enrich [input]",
	Short: i18n.Messages.EnrichCmdShortDesc,
	Long:  i18n.Messages.EnrichCmdLongDesc,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]
		log := logger.GetLogger()
		log.Info(i18n.Messages.StartingEnrichProcess)
		aiyouAssistantID, _ := cmd.Flags().GetString("aiyou-assistant-id")
		aiyouEmail, _ := cmd.Flags().GetString("aiyou-email")
		aiyouPassword, _ := cmd.Flags().GetString("aiyou-password")

		// Mettre à jour la configuration si les flags sont fournis
		cfg := config.GetConfig() // Obtenez l'instance de configuration
		if aiyouAssistantID != "" {
			cfg.AIYOUAssistantID = aiyouAssistantID
		}
		if aiyouEmail != "" {
			cfg.AIYOUEmail = aiyouEmail
		}
		if aiyouPassword != "" {
			cfg.AIYOUPassword = aiyouPassword
		}

		// Utiliser le chemin absolu pour l'entrée
		var absInput string
		if strings.HasPrefix(strings.ToLower(input), "s3://") {
			absInput = input
		} else {
			var err error
			absInput, err = filepath.Abs(input)
			if err != nil {
				return fmt.Errorf("error getting absolute path: %w", err)
			}
		}

		// Déterminer le nom de fichier de sortie si non spécifié
		if output == "" {
			if strings.HasPrefix(strings.ToLower(absInput), "s3://") {
				// Pour les entrées S3, conserver le format S3 pour la sortie
				output = strings.TrimSuffix(absInput, filepath.Ext(absInput)) + ".tsv"
			} else {
				// Pour les entrées locales, utiliser un chemin local
				output = filepath.Join(filepath.Dir(absInput), filepath.Base(absInput)+".tsv")
			}
		} else if !strings.HasPrefix(strings.ToLower(output), "s3://") && strings.HasPrefix(strings.ToLower(absInput), "s3://") {
			// Si l'entrée est S3 mais pas la sortie, convertir la sortie en format S3
			output = strings.TrimSuffix(absInput, filepath.Ext(absInput)) + ".tsv"
		}

		p, err := pipeline.NewPipeline(includePositions, contextOutput, contextWords, entityExtractionPrompt, relationExtractionPrompt, ontologyEnrichmentPrompt, ontologyMergePrompt, llm, llmModel, absInput, maxThreads, aiyouAssistantID)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.Messages.ErrorCreatingPipeline, err)
		}

		p.SetProgressCallback(func(info pipeline.ProgressInfo) {
			switch info.CurrentStep {
			case "Starting Pass":
				log.Info("Starting pass %d of %d", info.CurrentPass, info.TotalPasses)
			case "Segmenting":
				log.Info("Segmenting input into %d parts", info.TotalSegments)
			case "Processing Segment":
				log.Debug("Processing segment %d of %d", info.ProcessedSegments, info.TotalSegments)
			}
		})

		ontology := model.NewOntology() // Utiliser le nouveau package model

		err = p.ExecutePipeline(absInput, output, passes, existingOntology, ontology)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.Messages.ErrorExecutingPipeline, err)
		}

		log.Info(i18n.Messages.EnrichProcessCompleted)
		log.Info("File processed: %s, output: %s", absInput, output)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(enrichCmd)

	enrichCmd.Flags().StringVar(&output, "output", "", i18n.Messages.OutputFlagUsage)
	enrichCmd.Flags().StringVar(&format, "format", "", i18n.Messages.FormatFlagUsage)
	enrichCmd.Flags().StringVar(&llm, "llm", "", i18n.Messages.LLMFlagUsage)
	enrichCmd.Flags().StringVar(&llmModel, "llm-model", "", i18n.Messages.LLMModelFlagUsage)
	enrichCmd.Flags().IntVar(&passes, "passes", 1, i18n.Messages.PassesFlagUsage)
	enrichCmd.Flags().BoolVar(&recursive, "recursive", false, i18n.Messages.RecursiveFlagUsage)
	enrichCmd.Flags().StringVar(&existingOntology, "existing-ontology", "", i18n.Messages.ExistingOntologyFlagUsage)

	enrichCmd.Flags().StringVarP(&entityExtractionPrompt, "entity-prompt", "e", "", "Additional prompt for entity extraction")
	enrichCmd.Flags().StringVarP(&relationExtractionPrompt, "relation-prompt", "r", "", "Additional prompt for relation extraction")
	enrichCmd.Flags().StringVarP(&ontologyEnrichmentPrompt, "enrichment-prompt", "n", "", "Additional prompt for ontology enrichment")
	enrichCmd.Flags().StringVarP(&ontologyMergePrompt, "merge-prompt", "m", "", "Additional prompt for ontology merging")
	enrichCmd.Flags().IntVarP(&maxThreads, "max-threads", "t", 10, "Maximum number of concurrent threads for processing")

}

func ExecuteEnrichCommand(input, output string, passes int, existingOntology string, includePositions, contextOutput bool, contextWords int, entityPrompt, relationPrompt, enrichmentPrompt, mergePrompt string) error {
	log := logger.GetLogger()
	log.Info(i18n.Messages.StartingEnrichProcess)

	absInput := input // Garder l'input tel quel s'il est déjà en format S3

	if !strings.HasPrefix(strings.ToLower(input), "s3://") {
		var err error
		absInput, err = filepath.Abs(input)
		if err != nil {
			return fmt.Errorf("error getting absolute path: %w", err)
		}
	}

	if output == "" {
		// Générer le nom de fichier de sortie en conservant le format S3 si l'entrée est S3
		if strings.HasPrefix(strings.ToLower(absInput), "s3://") {
			// Pour les entrées S3, construire un chemin de sortie S3
			output = strings.TrimSuffix(absInput, filepath.Ext(absInput)) + ".tsv"
		} else {
			// Pour les entrées locales, utiliser le chemin local
			output = filepath.Join(filepath.Dir(absInput), filepath.Base(absInput)+".tsv")
		}
	} else if !strings.HasPrefix(strings.ToLower(output), "s3://") && strings.HasPrefix(strings.ToLower(absInput), "s3://") {
		// Si l'entrée est S3 mais pas la sortie, convertir la sortie en format S3
		output = "s3://" + strings.TrimPrefix(filepath.ToSlash(output), "/")
	}

	p, err := pipeline.NewPipeline(includePositions, contextOutput, contextWords, entityPrompt, relationPrompt, ontologyEnrichmentPrompt, mergePrompt, llm, llmModel, absInput, maxThreads, aiyouAssistantID)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.Messages.ErrorCreatingPipeline, err)
	}

	p.SetProgressCallback(func(info pipeline.ProgressInfo) {
		switch info.CurrentStep {
		case "Starting Pass":
			log.Info("Starting pass %d of %d", info.CurrentPass, info.TotalPasses)
		case "Segmenting":
			log.Info("Segmenting input into %d parts", info.TotalSegments)
		case "Processing Segment":
			log.Debug("Processing segment %d of %d", info.ProcessedSegments, info.TotalSegments)
		}
	})

	onto := model.NewOntology()
	err = p.ExecutePipeline(absInput, output, passes, existingOntology, onto)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.Messages.ErrorExecutingPipeline, err)
	}

	log.Info(i18n.Messages.EnrichProcessCompleted)
	log.Info("File processed: %s, output: %s", absInput, output)
	return nil
}

func generateOutputFilename(input string) string {
	dir := filepath.Dir(input)
	baseName := filepath.Base(input)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))

	return filepath.Join(dir, baseName+".tsv")
}
