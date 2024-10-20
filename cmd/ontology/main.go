// cmd/ontology/main.go

package main

import (
	"os"

	"github.com/chrlesur/Ontology/internal/cmd/ontology"
	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/logger"
)

var log = logger.GetLogger()

func main() {
	cfg := config.GetConfig()
	if err := cfg.ValidateConfig(); err != nil {
		log.Error(i18n.GetMessage("ConfigurationError"), err)
		os.Exit(1)
	}

	if err := ontology.Execute(); err != nil {
		log.Error(i18n.GetMessage("ExecutionError"), err)
		os.Exit(1)
	}
}
