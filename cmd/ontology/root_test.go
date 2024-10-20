package ontology

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	cmd := &cobra.Command{Use: "root"}
	cmd.AddCommand(rootCmd)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error executing root command: %v", err)
	}
}

func TestEnrichCommand(t *testing.T) {
	cmd := &cobra.Command{Use: "root"}
	cmd.AddCommand(enrichCmd)

	// Test with missing required flags
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error due to missing required flags, but got none")
	}

	// Test with required flags
	cmd.SetArgs([]string{"enrich", "--input", "test.txt", "--output", "out.txt"})
	err = cmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error executing enrich command: %v", err)
	}
}
