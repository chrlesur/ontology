package segmenter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/chrlesur/Ontology/internal/i18n"
	"github.com/chrlesur/Ontology/internal/logger"
	"github.com/pkoukk/tiktoken-go"
)

var (
	ErrInvalidContent = errors.New(i18n.ErrInvalidContent)
	ErrTokenization   = errors.New(i18n.ErrTokenization)
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
	log.Debug(i18n.LogSegmentationStarted)
	if len(content) == 0 {
		return nil, ErrInvalidContent
	}

	tokenizer, err := getTokenizer(cfg.Model)
	if err != nil {
		return nil, err
	}

	var segments [][]byte
	reader := bytes.NewReader(content)
	buffer := new(bytes.Buffer)
	tokenCount := 0

	for {
		char, _, err := reader.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", i18n.ErrReadingContent, err)
		}

		buffer.WriteRune(char)
		tokenCount = CountTokens(buffer.Bytes(), tokenizer)

		if tokenCount >= cfg.MaxTokens || char == '.' || char == '!' || char == '?' {
			segment := make([]byte, buffer.Len())
			copy(segment, buffer.Bytes())
			segments = append(segments, segment)
			buffer.Reset()
			tokenCount = 0
		}
	}

	if buffer.Len() > 0 {
		segments = append(segments, buffer.Bytes())
	}

	log.Info(i18n.LogSegmentationCompleted, len(segments))
	return segments, nil
}

// GetContext returns the context of previous segments
func GetContext(segments [][]byte, currentIndex int, cfg SegmentConfig) string {
	log.Debug(i18n.LogContextGeneration)
	if currentIndex == 0 {
		return ""
	}

	tokenizer, err := getTokenizer(cfg.Model)
	if err != nil {
		log.Error(i18n.ErrTokenizerInitialization, err)
		return ""
	}

	var context strings.Builder
	tokenCount := 0

	for i := currentIndex - 1; i >= 0; i-- {
		segment := segments[i]
		segmentTokens := CountTokens(segment, tokenizer)

		if tokenCount+segmentTokens > cfg.ContextSize {
			break
		}

		context.Write(segment)
		context.WriteString(" ")
		tokenCount += segmentTokens
	}

	return strings.TrimSpace(context.String())
}

// CountTokens returns the number of tokens in the content
func CountTokens(content []byte, tokenizer *tiktoken.Tiktoken) int {
	tokens := tokenizer.Encode(string(content), nil, nil)
	return len(tokens)
}

// MergeSegments reconstitutes the original document from segments
func MergeSegments(segments [][]byte) []byte {
	log.Debug(i18n.LogMergingSegments)
	return bytes.Join(segments, []byte(" "))
}

// getTokenizer returns a tokenizer for the specified model
func getTokenizer(model string) (*tiktoken.Tiktoken, error) {
	encoding, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.ErrTokenizerInitialization, err)
	}
	return encoding, nil
}

// CalibrateTokenCount adjusts the token count based on the LLM model
func CalibrateTokenCount(count int, model string) int {
	// Implement model-specific calibration logic here
	// For now, we'll just return the original count
	return count
}
