package ontology

import (
	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/pipeline"
	"github.com/spf13/cobra"
)

var (
	input            string
	output           string
	format           string
	llm              string
	llmModel         string
	rdf              bool
	owl              bool
	recursive        bool
	existingOntology string
)

// enrichCmd represents the enrich command
var enrichCmd = &cobra.Command{
	Use:   "enrich",
	Short: i18n.Messages.EnrichCmdShortDesc,
	Long:  i18n.Messages.EnrichCmdLongDesc,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info(i18n.Messages.StartingEnrichProcess)

		// Créer une nouvelle instance de Pipeline
		p, err := pipeline.NewPipeline()
		if err != nil {
			log.Error(i18n.Messages.ErrorCreatingPipeline, err)
			return
		}

		// Appeler ExecutePipeline sur l'instance de Pipeline
		err = p.ExecutePipeline(input, passes, existingOntology)
		if err != nil {
			log.Error(i18n.Messages.ErrorExecutingPipeline, err)
			return
		}

		log.Info(i18n.Messages.EnrichProcessCompleted)
	},
}

func init() {
	rootCmd.AddCommand(enrichCmd)

	enrichCmd.Flags().StringVar(&input, "input", "", i18n.Messages.InputFlagUsage)
	enrichCmd.Flags().StringVar(&output, "output", "", i18n.Messages.OutputFlagUsage)
	enrichCmd.Flags().StringVar(&format, "format", "", i18n.Messages.FormatFlagUsage)
	enrichCmd.Flags().StringVar(&llm, "llm", "", i18n.Messages.LLMFlagUsage)
	enrichCmd.Flags().StringVar(&llmModel, "llm-model", "", i18n.Messages.LLMModelFlagUsage)
	enrichCmd.Flags().IntVar(&passes, "passes", 1, i18n.Messages.PassesFlagUsage)
	enrichCmd.Flags().BoolVar(&rdf, "rdf", false, i18n.Messages.RDFFlagUsage)
	enrichCmd.Flags().BoolVar(&owl, "owl", false, i18n.Messages.OWLFlagUsage)
	enrichCmd.Flags().BoolVar(&recursive, "recursive", false, i18n.Messages.RecursiveFlagUsage)
	enrichCmd.Flags().StringVar(&existingOntology, "existing-ontology", "", i18n.Messages.ExistingOntologyFlagUsage) // Ajoutez cette ligne

	enrichCmd.MarkFlagRequired("input")
	enrichCmd.MarkFlagRequired("output")
}
