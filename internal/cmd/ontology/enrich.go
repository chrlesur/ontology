package ontology

import (
	"github.com/chrlesur/Ontology/internal/i18n"
    "github.com/chrlesur/Ontology/internal/pipeline"
	"github.com/spf13/cobra"
)

var (
	input     string
	output    string
	format    string
	llm       string
	llmModel  string
	rdf       bool
	owl       bool
	recursive bool
)

// enrichCmd represents the enrich command
var enrichCmd = &cobra.Command{
	Use:   "enrich",
	Short: i18n.EnrichCmdShortDesc,
	Long:  i18n.EnrichCmdLongDesc,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info(i18n.StartingEnrichProcess)

		// TODO: Implement actual pipeline execution
		err := pipeline.ExecutePipeline()
		if err != nil {
			log.Error(i18n.ErrorExecutingPipeline, err)
			return
		}

		log.Info(i18n.EnrichProcessCompleted)
	},
}

func init() {
	rootCmd.AddCommand(enrichCmd)

	enrichCmd.Flags().StringVar(&input, "input", "", i18n.InputFlagUsage)
	enrichCmd.Flags().StringVar(&output, "output", "", i18n.OutputFlagUsage)
	enrichCmd.Flags().StringVar(&format, "format", "", i18n.FormatFlagUsage)
	enrichCmd.Flags().StringVar(&llm, "llm", "", i18n.LLMFlagUsage)
	enrichCmd.Flags().StringVar(&llmModel, "llm-model", "", i18n.LLMModelFlagUsage)
	enrichCmd.Flags().IntVar(&passes, "passes", 1, i18n.PassesFlagUsage)
	enrichCmd.Flags().BoolVar(&rdf, "rdf", false, i18n.RDFFlagUsage)
	enrichCmd.Flags().BoolVar(&owl, "owl", false, i18n.OWLFlagUsage)
	enrichCmd.Flags().BoolVar(&recursive, "recursive", false, i18n.RecursiveFlagUsage)

	enrichCmd.MarkFlagRequired("input")
	enrichCmd.MarkFlagRequired("output")
}
