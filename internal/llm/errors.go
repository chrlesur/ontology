package llm

import (
	"errors"

	"github.com/chrlesur/Ontology/internal/i18n"
)

var (
	ErrUnsupportedModel  = errors.New(i18n.Messages.ErrUnsupportedModel)
	ErrAPIKeyMissing     = errors.New(i18n.Messages.ErrAPIKeyMissing)
	ErrTranslationFailed = errors.New(i18n.Messages.ErrTranslationFailed)
	ErrInvalidLLMType    = errors.New(i18n.Messages.ErrInvalidLLMType)
	ErrContextTooLong    = errors.New(i18n.Messages.ErrContextTooLong)
)
