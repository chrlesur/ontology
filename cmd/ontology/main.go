// cmd/ontology/main.go

package main

import (
	"fmt"
	"os"

	"github.com/chrlesur/Ontology/cmd/ontology"
	"github.com/chrlesur/Ontology/internal/config"
)

func main() {
	cfg := config.GetConfig()
	if err := cfg.ValidateConfig(); err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		os.Exit(1)
	}
	ontology.Execute()
}
