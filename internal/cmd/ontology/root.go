package ontology

import (
	"fmt"
	"os"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/pipeline"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	inputFile string
	passes    int
	ontology  string
)

var rootCmd = &cobra.Command{
	Use:   "ontology",
	Short: i18n.GetMessage("RootCmdShort"),
	Long:  i18n.GetMessage("RootCmdLong"),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := pipeline.NewPipeline()
		if err != nil {
			return err
		}
		return p.ExecutePipeline(inputFile, passes, ontology)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", i18n.GetMessage("ConfigFlagUsage"))
	rootCmd.PersistentFlags().StringVar(&inputFile, "input", "", i18n.GetMessage("InputFlagUsage"))
	rootCmd.PersistentFlags().IntVar(&passes, "passes", 1, i18n.GetMessage("PassesFlagUsage"))
	rootCmd.PersistentFlags().StringVar(&ontology, "ontology", "", i18n.GetMessage("OntologyFlagUsage"))
	rootCmd.PersistentFlags().BoolVar(&config.GetConfig().ExportRDF, "rdf", false, i18n.GetMessage("RDFFlagUsage"))
	rootCmd.PersistentFlags().BoolVar(&config.GetConfig().ExportOWL, "owl", false, i18n.GetMessage("OWLFlagUsage"))
	rootCmd.MarkFlagRequired("input")
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
	log.SetLevel(log.ParseLevel(logLevel))

	// Initialize other configurations
	config.InitConfig(viper.GetViper())
}
