package segmenter

import (
    "testing"

    "github.com/chrlesur/Ontology/internal/config"
    "github.com/stretchr/testify/assert"
)

func TestSegment(t *testing.T) {
    content := []byte("This is a test sentence. This is another test sentence. And a third one.")
    cfg := SegmentConfig{MaxTokens: 10, ContextSize: 5, Model: "gpt-3.5-turbo"}

    segments, err := Segment(content, cfg)
    assert.NoError(t, err)
    assert.Len(t, segments, 3)
}

func TestGetContext(t *testing.T) {
    segments := [][]byte{
        []byte("First segment."),
        []byte("Second segment."),
        []byte("Third segment."),
    }
    cfg := SegmentConfig{MaxTokens: 10, ContextSize: 5, Model: "gpt-3.5-turbo"}

    context := GetContext(segments, 2, cfg)
    assert.Equal(t, "Second segment. First segment.", context)
}

func TestCountTokens(t *testing.T) {
    content := []byte("This is a test sentence.")
    tokenizer, _ := getTokenizer("gpt-3.5-turbo")
    count := CountTokens(content, tokenizer)
    assert.Equal(t, 5, count)
}

func TestMergeSegments(t *testing.T) {
    segments := [][]byte{
        []byte("First segment."),
        []byte("Second segment."),
        []byte("Third segment."),
    }

    merged := MergeSegments(segments)
    assert.Equal(t, "First segment. Second segment. Third segment.", string(merged))
}

func TestCalibrateTokenCount(t *testing.T) {
    count := 100
    model := "gpt-4"
    calibrated := CalibrateTokenCount(count, model)
    assert.Equal(t, count, calibrated) // For now, it should return the same count
}