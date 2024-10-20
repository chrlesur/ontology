package segmenter

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/logger"
	"github.com/pkoukk/tiktoken-go"
)

var (
	ErrInvalidContent = errors.New(i18n.Messages.ErrInvalidContent)
	ErrTokenization   = errors.New(i18n.Messages.ErrTokenization)
)

var log = logger.GetLogger()

// SegmentConfig holds the configuration for segmentation
type SegmentConfig struct {
	MaxTokens   int
	ContextSize int
	Model       string
}

// Segment divides the content into segments of maxTokens
func Segment(content []byte, cfg SegmentConfig) ([][]byte, error) {
	log.Debug(i18n.Messages.LogSegmentationStarted)
	log.Debug(fmt.Sprintf("Segmentation config: MaxTokens=%d, ContextSize=%d, Model=%s", cfg.MaxTokens, cfg.ContextSize, cfg.Model))
	log.Debug(fmt.Sprintf("Content length: %d bytes", len(content)))

	if len(content) == 0 {
		return nil, ErrInvalidContent
	}

	tokenizer, err := getTokenizer(cfg.Model)
	if err != nil {
		return nil, err
	}

	var segments [][]byte
	sentences := splitIntoSentences(content)
	currentSegment := new(bytes.Buffer)
	currentTokenCount := 0

	for _, sentence := range sentences {
		sentenceTokens := CountTokens(sentence, tokenizer)

		if currentTokenCount+sentenceTokens > cfg.MaxTokens {
			if currentSegment.Len() > 0 {
				segments = append(segments, currentSegment.Bytes())
				log.Debug(fmt.Sprintf("Segment created with %d tokens", currentTokenCount))
				currentSegment.Reset()
				currentTokenCount = 0
			}
		}

		currentSegment.Write(sentence)
		currentTokenCount += sentenceTokens

		if currentTokenCount >= cfg.MaxTokens {
			segments = append(segments, currentSegment.Bytes())
			log.Debug(fmt.Sprintf("Segment created with %d tokens", currentTokenCount))
			currentSegment.Reset()
			currentTokenCount = 0
		}
	}

	if currentSegment.Len() > 0 {
		segments = append(segments, currentSegment.Bytes())
		log.Debug(fmt.Sprintf("Final segment created with %d tokens", currentTokenCount))
	}

	log.Info(fmt.Sprintf(i18n.Messages.LogSegmentationCompleted, len(segments)))
	return segments, nil
}

func splitIntoSentences(content []byte) [][]byte {
	var sentences [][]byte
	var currentSentence []byte

	for _, b := range content {
		currentSentence = append(currentSentence, b)
		if b == '.' || b == '!' || b == '?' {
			sentences = append(sentences, currentSentence)
			currentSentence = []byte{}
		}
	}

	if len(currentSentence) > 0 {
		sentences = append(sentences, currentSentence)
	}

	return sentences
}

// CountTokens returns the number of tokens in the content
func CountTokens(content []byte, tokenizer *tiktoken.Tiktoken) int {
	tokens := tokenizer.Encode(string(content), nil, nil)
	count := len(tokens)
	return count
}

// getTokenizer returns a tokenizer for the specified model
func getTokenizer(model string) (*tiktoken.Tiktoken, error) {
	encoding, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.Messages.ErrTokenizerInitialization, err)
	}
	return encoding, nil
}

// CalibrateTokenCount adjusts the token count based on the LLM model
func CalibrateTokenCount(count int, model string) int {
	log.Debug("Calibrating token count for model %s. Original count: %d", model, count)
	// Implement model-specific calibration logic here
	// For now, we'll just return the original count
	log.Debug("Calibrated count: %d", count)
	return count
}

// GetContext returns the context of previous segments
func GetContext(segments [][]byte, currentIndex int, cfg SegmentConfig) string {
	log.Debug(i18n.Messages.LogContextGeneration)
	if currentIndex == 0 {
		return ""
	}

	tokenizer, err := getTokenizer(cfg.Model)
	if err != nil {
		log.Error(i18n.Messages.ErrTokenizerInitialization, err)
		return ""
	}

	var context bytes.Buffer
	tokenCount := 0

	for i := currentIndex - 1; i >= 0; i-- {
		segmentTokens := CountTokens(segments[i], tokenizer)

		if tokenCount+segmentTokens > cfg.ContextSize {
			break
		}

		// Prepend this segment to the context
		temp := make([]byte, len(segments[i]))
		copy(temp, segments[i])
		context.Write(temp)
		context.WriteByte('\n')

		tokenCount += segmentTokens
	}

	log.Debug(fmt.Sprintf("Generated context with %d tokens", tokenCount))
	return context.String()
}
