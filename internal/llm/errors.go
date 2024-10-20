package llm

import (
	"errors"

	"github.com/chrlesur/Ontology/internal/i18n"
)

var (
	ErrUnsupportedModel  = errors.New(i18n.ErrUnsupportedModel)
	ErrAPIKeyMissing     = errors.New(i18n.ErrAPIKeyMissing)
	ErrTranslationFailed = errors.New(i18n.ErrTranslationFailed)
	ErrInvalidLLMType    = errors.New(i18n.ErrInvalidLLMType)
	ErrContextTooLong    = errors.New(i18n.ErrContextTooLong)
)
