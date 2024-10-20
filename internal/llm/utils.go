package llm

import (
	"github.com/chrlesur/Ontology/internal/tokenizer"
)

func CheckContextLength(model string, context string) error {
	limit, ok := ModelContextLimits[model]
	if !ok {
		return ErrUnsupportedModel
	}

	tokenCount, err := tokenizer.CountTokens(context)
	if err != nil {
		return err
	}

	if tokenCount > limit {
		return ErrContextTooLong
	}

	return nil
}
