package ontology

import (
	"fmt"
	"os"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/logger"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	debug   bool
	silent  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ontology",
	Short: i18n.GetMessage("RootCmdShortDesc"),
	Long:  i18n.GetMessage("RootCmdLongDesc"),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", i18n.GetMessage("ConfigFlagUsage"))
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, i18n.GetMessage("DebugFlagUsage"))
	rootCmd.PersistentFlags().BoolVar(&silent, "silent", false, i18n.GetMessage("SilentFlagUsage"))

	rootCmd.Run = rootCmd.HelpFunc()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Set the config file path if specified
		os.Setenv("ONTOLOGY_CONFIG_PATH", cfgFile)
	}

	// This will load the config from file and environment variables
	cfg := config.GetConfig()

	// Validate the config
	if err := cfg.ValidateConfig(); err != nil {
		fmt.Printf("Error in configuration: %v\n", err)
		os.Exit(1)
	}

	log := logger.GetLogger()

	// Initialize logger based on debug and silent flags
	if debug {
		log.SetLevel(logger.DebugLevel)
	} else if silent {
		log.SetLevel(logger.ErrorLevel)
	} else {
		logLevel := logger.ParseLevel(cfg.LogLevel)
		log.SetLevel(logLevel)
	}

	log.Info(i18n.GetMessage("InitializingApplication"))
}
