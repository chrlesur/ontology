package ontology

import (
	"fmt"
	"os"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/logger"
	"github.com/chrlesur/Ontology/internal/pipeline"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	inputFile string
	passes    int
	ontology  string
	DebugMode bool
)

var rootCmd = &cobra.Command{
	Use:   "ontology",
	Short: i18n.Messages.RootCmdShortDesc,
	Long:  i18n.Messages.RootCmdLongDesc,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := pipeline.NewPipeline()
		if err != nil {
			return err
		}
		return p.ExecutePipeline(inputFile, passes, ontology)
	},
}

func Execute() error {
	if DebugMode {
		log.Debug(i18n.Messages.DebugFlagUsage)
		logger.GetLogger().SetLevel(logger.DebugLevel)
	}
	err := rootCmd.Execute()
	if err != nil {
		log.Error(i18n.GetMessage("CommandExecutionError"), err)
	}

	return err
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", i18n.GetMessage("ConfigFlagUsage"))
	rootCmd.PersistentFlags().StringVar(&inputFile, "input", "", i18n.GetMessage("InputFlagUsage"))
	rootCmd.PersistentFlags().IntVar(&passes, "passes", 1, i18n.GetMessage("PassesFlagUsage"))
	rootCmd.PersistentFlags().StringVar(&ontology, "ontology", "", i18n.GetMessage("OntologyFlagUsage"))
	rootCmd.PersistentFlags().BoolVar(&config.GetConfig().ExportRDF, "rdf", false, i18n.GetMessage("RDFFlagUsage"))
	rootCmd.PersistentFlags().BoolVar(&config.GetConfig().ExportOWL, "owl", false, i18n.GetMessage("OWLFlagUsage"))
	rootCmd.PersistentFlags().BoolVar(&DebugMode, "debug", false, "Enable debug mode with detailed logging")
	rootCmd.MarkFlagRequired("input")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if DebugMode {
			log.SetLevel(logger.DebugLevel)
			log.Debug("Debug mode enabled in root.go")
		}
	}
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".ontology" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigName(".ontology")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// Initialize logger
	logLevel := viper.GetString("log_level")
	if logLevel == "" {
		logLevel = "info"
	}
	log.SetLevel(logger.ParseLevel(logLevel))
	// Initialize other configurations
	if err := config.GetConfig().Reload(); err != nil {
		fmt.Printf("Error reloading config: %v\n", err)
		os.Exit(1)
	}
}
