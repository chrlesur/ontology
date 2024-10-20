package ontology

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/chrlesur/Ontology/internal/config"
    "github.com/chrlesur/Ontology/internal/logger"
    "github.com/chrlesur/Ontology/internal/i18n"
)

var (
    cfgFile string
    debug   bool
    silent  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
    Use:   "ontology",
    Short: i18n.RootCmdShortDesc,
    Long:  i18n.RootCmdLongDesc,
    // Uncomment the following line if your bare application
    // has an action associated with it:
    // Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
    err := rootCmd.Execute()
    if err != nil {
        logger.Error(i18n.ErrorExecutingRootCmd, err)
        os.Exit(1)
    }
}

func init() {
    cobra.OnInitialize(initConfig)

    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", i18n.ConfigFlagUsage)
    rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, i18n.DebugFlagUsage)
    rootCmd.PersistentFlags().BoolVar(&silent, "silent", false, i18n.SilentFlagUsage)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
    if cfgFile != "" {
        // Use config file from the flag.
        config.LoadConfig(cfgFile)
    }

    // Initialize logger based on debug and silent flags
    if debug {
        logger.SetLevel(logger.DebugLevel)
    } else if silent {
        logger.SetLevel(logger.ErrorLevel)
    } else {
        logger.SetLevel(logger.InfoLevel)
    }

    logger.Info(i18n.InitializingApplication)
}