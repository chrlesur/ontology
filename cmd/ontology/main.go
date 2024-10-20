package main

import (
	"fmt"
	"os"

	"github.com/chrlesur/Ontology/internal/cmd/ontology"
)

func main() {
	if err := ontology.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
