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
type SegmentInfo struct {
	Content []byte
	Start   int
	End     int
}

func Segment(content []byte, cfg SegmentConfig) ([]SegmentInfo, []int, error) {
	log.Debug(i18n.Messages.LogSegmentationStarted)
	log.Debug(fmt.Sprintf("Segmentation config: MaxTokens=%d, ContextSize=%d, Model=%s", cfg.MaxTokens, cfg.ContextSize, cfg.Model))
	log.Debug(fmt.Sprintf("Content length: %d bytes", len(content)))

	if len(content) == 0 {
		return nil, nil, ErrInvalidContent
	}

	tokenizer, err := getTokenizer(cfg.Model)
	if err != nil {
		return nil, nil, err
	}

	var segments []SegmentInfo
	var offsets []int
	currentSegment := new(bytes.Buffer)
	currentTokenCount := 0
	currentStart := 0

	sentences := splitIntoSentences(content)
	for _, sentence := range sentences {
		sentenceTokens := CountTokens(sentence, tokenizer)

		if currentTokenCount+sentenceTokens > cfg.MaxTokens {
			if currentSegment.Len() > 0 {
				segmentContent := content[currentStart : currentStart+currentSegment.Len()]
				segments = append(segments, SegmentInfo{
					Content: segmentContent,
					Start:   currentStart,
					End:     currentStart + currentSegment.Len(),
				})
				offsets = append(offsets, currentStart)
				log.Debug(fmt.Sprintf("Segment created. Start: %d, End: %d, Length: %d bytes, Tokens: %d",
					currentStart, currentStart+currentSegment.Len(), len(segmentContent), currentTokenCount))
				currentStart = currentStart + currentSegment.Len()
				currentSegment.Reset()
				currentTokenCount = 0
			}
		}

		currentSegment.Write(sentence)
		currentTokenCount += sentenceTokens
	}

	if currentSegment.Len() > 0 {
		segmentContent := content[currentStart : currentStart+currentSegment.Len()]
		segments = append(segments, SegmentInfo{
			Content: segmentContent,
			Start:   currentStart,
			End:     currentStart + currentSegment.Len(),
		})
		offsets = append(offsets, currentStart)
		log.Debug(fmt.Sprintf("Final segment created. Start: %d, End: %d, Length: %d bytes, Tokens: %d",
			currentStart, currentStart+currentSegment.Len(), len(segmentContent), currentTokenCount))
	}

	log.Info(fmt.Sprintf(i18n.Messages.LogSegmentationCompleted, len(segments)))
	return segments, offsets, nil
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
func GetContext(segments []SegmentInfo, currentIndex int, cfg SegmentConfig) string {
	log.Debug("GetContext called for segment %d/%d", currentIndex+1, len(segments))

	var context bytes.Buffer
	tokenCount := 0
	segmentsUsed := 0

	tokenizer, err := getTokenizer(cfg.Model)
	if err != nil {
		log.Error("Failed to get tokenizer: %v", err)
		return ""
	}

	for i := currentIndex - 1; i >= 0 && tokenCount < cfg.ContextSize; i-- {
		segmentTokens := CountTokens(segments[i].Content, tokenizer)
		if tokenCount+segmentTokens > cfg.ContextSize {
			break
		}
		// Prepend this segment to the context
		temp := make([]byte, len(segments[i].Content)+1)
		copy(temp[1:], segments[i].Content)
		temp[0] = '\n'
		context.Write(temp)
		tokenCount += segmentTokens
		segmentsUsed++
	}

	log.Debug("Generated context for segment %d/%d. Segments used: %d, Token count: %d, Context length: %d bytes",
		currentIndex+1, len(segments), segmentsUsed, tokenCount, context.Len())
	log.Debug("Context preview for segment %d: %s",
		currentIndex+1, truncateString(context.String(), 100))

	return context.String()
}

func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}
