package pipeline

import (
	"os"
	"testing"
	"time"

	"github.com/chrlesur/Ontology/internal/logger"
	"github.com/chrlesur/Ontology/internal/storage"
	"github.com/stretchr/testify/assert"
)

// MockStorage implémente l'interface Storage pour les tests
type MockStorage struct {
	IsDirectoryFunc func(string) (bool, error)
	StatFunc        func(string) (storage.FileInfo, error)
	ReadFunc        func(string) ([]byte, error)
	WriteFunc       func(string, []byte) error
	ListFunc        func(string) ([]string, error)
	DeleteFunc      func(string) error
	ExistsFunc      func(string) (bool, error)
}

type MockFileInfo struct {
	NameValue    string
	SizeValue    int64
	ModeValue    os.FileMode
	ModTimeValue time.Time
	IsDirValue   bool
}

type FileInfo interface {
	Name() string
	Size() int64
	Mode() os.FileMode
	ModTime() time.Time
	IsDir() bool
	Sys() interface{}
}

func (m *MockStorage) IsDirectory(path string) (bool, error) {
	if m.IsDirectoryFunc != nil {
		return m.IsDirectoryFunc(path)
	}
	return false, nil
}

func (m *MockStorage) Stat(path string) (storage.FileInfo, error) {
	if m.StatFunc != nil {
		return m.StatFunc(path)
	}
	return &MockFileInfo{}, nil
}

func (m *MockStorage) Read(path string) ([]byte, error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(path)
	}
	return []byte("Mock content"), nil
}

func (m *MockStorage) Write(path string, data []byte) error {
	if m.WriteFunc != nil {
		return m.WriteFunc(path, data)
	}
	return nil
}

func (m *MockStorage) List(prefix string) ([]string, error) {
	if m.ListFunc != nil {
		return m.ListFunc(prefix)
	}
	return []string{}, nil
}

func (m *MockStorage) Delete(path string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(path)
	}
	return nil
}

func (m *MockStorage) Exists(path string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(path)
	}
	return true, nil
}

func (mfi *MockFileInfo) Name() string       { return mfi.NameValue }
func (mfi *MockFileInfo) Size() int64        { return mfi.SizeValue }
func (mfi *MockFileInfo) Mode() os.FileMode  { return mfi.ModeValue }
func (mfi *MockFileInfo) ModTime() time.Time { return mfi.ModTimeValue }
func (mfi *MockFileInfo) IsDir() bool        { return mfi.IsDirValue }
func (mfi *MockFileInfo) Sys() interface{}   { return nil }

func TestNewPipeline(t *testing.T) {
	p, err := NewPipeline(true, false, 30, "", "", "", "", "claude", "claude-3-5-sonnet-20240620")
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.True(t, p.includePositions)
	assert.False(t, p.contextOutput)
	assert.Equal(t, 30, p.contextWords)
}

func TestCreatePositionIndex(t *testing.T) {
	p := &Pipeline{
		logger: logger.GetLogger(),
	}
	content := []byte("This is a test. This is another test with some_underscore.")
	index := p.createPositionIndex(content)

	assert.Equal(t, []int{0}, index["this"])
	assert.Equal(t, []int{1, 5}, index["is"])
	assert.Equal(t, []int{3, 7}, index["test"])
	assert.Equal(t, []int{9}, index["another"])
	assert.Equal(t, []int{11}, index["some_underscore"])
	assert.Equal(t, []int{11}, index["some underscore"])
}

func TestFindPositions(t *testing.T) {
	p := &Pipeline{
		logger: logger.GetLogger(),
	}
	content := "This is a test. This is another test with some_underscore."
	index := p.createPositionIndex([]byte(content))

	tests := []struct {
		word     string
		expected []int
	}{
		{"test", []int{3, 7}},
		{"This", []int{0, 4}},
		{"is", []int{1, 5}},
		{"some_underscore", []int{11}},
		{"nonexistent", []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			positions := p.findPositions(tt.word, index, content)
			assert.Equal(t, tt.expected, positions)
		})
	}
}

func TestNormalizeWord(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello", "hello"},
		{"World!", "world"},
		{"l'exemple", "l'exemple"},
		{"UPPER_CASE", "upper case"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, normalizeWord(tt.input))
	}
}

func TestGenerateArticleVariants(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"example", []string{"example", "l'example", "d'example", "l example", "d example"}},
		{"underscore_word", []string{
			"underscore_word", "l'underscore_word", "d'underscore_word", "l underscore_word", "d underscore_word",
			"underscore word", "l'underscore word", "d'underscore word", "l underscore word", "d underscore word",
		}},
		{"l'apple", []string{"l'apple"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			variants := generateArticleVariants(tt.input)
			assert.ElementsMatch(t, tt.expected, variants)
		})
	}
}

func TestTruncateString(t *testing.T) {
	assert.Equal(t, "Hello...", truncateString("Hello, World!", 5))
	assert.Equal(t, "Hello, World!", truncateString("Hello, World!", 20))
}

func TestMergeOverlappingPositions(t *testing.T) {
	positions := []PositionRange{
		{Start: 0, End: 5, Element: "A"},
		{Start: 3, End: 8, Element: "B"},
		{Start: 10, End: 15, Element: "C"},
	}
	merged := mergeOverlappingPositions(positions)
	assert.Equal(t, 2, len(merged))
	assert.Equal(t, PositionRange{Start: 0, End: 8, Element: "A"}, merged[0])
}

func TestGenerateContextJSON(t *testing.T) {
	content := []byte("This is a test sentence for context generation.")
	positions := []int{3, 7}
	positionRanges := []PositionRange{
		{Start: 3, End: 3, Element: "test"},
		{Start: 7, End: 7, Element: "sentence"},
	}

	json, err := GenerateContextJSON(content, positions, 2, positionRanges)
	assert.NoError(t, err)
	assert.Contains(t, json, "test")
	assert.Contains(t, json, "sentence")
}

