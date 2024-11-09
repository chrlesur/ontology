// pipeline/segmentation_test.go

package pipeline

import (
	"fmt"
	"testing"
	"time"

	"github.com/chrlesur/Ontology/internal/config"
	"github.com/chrlesur/Ontology/internal/logger"
	"github.com/stretchr/testify/assert"
)

// newTestPipeline crée une instance de Pipeline pour les tests
func newTestPipeline() *Pipeline {
	cfg := &config.Config{
		// Ajoutez ici les configurations nécessaires pour le test
		MaxTokens:    1000,
		ContextSize:  2000,
		DefaultLLM:   "test-llm",
		DefaultModel: "test-model",
	}

	return &Pipeline{
		logger: logger.GetLogger(),
		config: cfg,
		// Ajoutez d'autres champs nécessaires pour le test
	}
}

func TestMergeResultsWithDB(t *testing.T) {
	p := newTestPipeline()

	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)
	previousResult := fmt.Sprintf("Entity1\tType1\tDescription1\t%s\t%s\tSource1\n", nowStr, nowStr)
	newResults := []string{
		fmt.Sprintf("Entity2\tType2\tUpdated Description2\t%s\t%s\tSource2\n", nowStr, nowStr),
		fmt.Sprintf("Entity3\tType3\tDescription3\t%s\t%s\tSource3\n", nowStr, nowStr),
		fmt.Sprintf("Relation1\tRelType\tEntity1\tEntity2\tDescription\t0.50\tforward\t%s\t%s\n", nowStr, nowStr),
	}

	t.Logf("Previous result: %s", previousResult)
	t.Logf("New results: %v", newResults)

	mergedResult, err := p.mergeResultsWithDB(previousResult, newResults)

	assert.NoError(t, err)
	t.Logf("Merged result: %s", mergedResult)

	expectedResult := fmt.Sprintf("Entity1\tType1\tDescription1\t%s\t%s\tSource1\n"+
		"Entity2\tType2\tUpdated Description2\t%s\t%s\tSource2\n"+
		"Entity3\tType3\tDescription3\t%s\t%s\tSource3\n"+
		"Relation1\tRelType\tEntity1\tEntity2\tDescription\t0.50\tforward\t%s\t%s\n",
		nowStr, nowStr, nowStr, nowStr, nowStr, nowStr, nowStr, nowStr)

	assert.Equal(t, expectedResult, mergedResult)
}
