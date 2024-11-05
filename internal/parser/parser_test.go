package parser

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chrlesur/Ontology/internal/metadata"
	"github.com/chrlesur/Ontology/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestPDFParser(t *testing.T) {
    content := `%PDF-1.5
    1 0 obj
    <</Type/Catalog/Pages 2 0 R>>
    endobj
    2 0 obj
    <</Type/Pages/Kids[3 0 R]/Count 1>>
    endobj
    3 0 obj
    <</Type/Page/Parent 2 0 R/MediaBox[0 0 612 792]/Resources<<>>/Contents 4 0 R>>
    endobj
    4 0 obj
    <</Length 51>>
    stream
    BT
    /F1 12 Tf
    100 700 Td
    (Test PDF Content) Tj
    ET
    endstream
    endobj
    xref
    0 5
    0000000000 65535 f
    0000000009 00000 n
    0000000052 00000 n
    0000000101 00000 n
    0000000192 00000 n
    trailer
    <</Size 5/Root 1 0 R/Info<</Title(Test PDF)/Author(Test Author)>>>>
    startxref
    291
    %%EOF`

	parser := NewPDFParser()
	result, err := parser.Parse(bytes.NewReader([]byte(content)))

	assert.NoError(t, err)
	assert.Contains(t, string(result), "Test PDF")

	metadata := parser.GetFormatMetadata()
	assert.Equal(t, "PDF", metadata["format"])
	assert.Equal(t, "Test PDF", metadata["title"])
	assert.Equal(t, "Test Author", metadata["author"])
}

func TestDOCXParser(t *testing.T) {
	// This is a simplified DOCX content for testing purposes
    content := `
    <?xml version="1.0" encoding="UTF-8" standalone="yes"?>
    <w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
      <w:body>
        <w:p>
          <w:r>
            <w:t>This is a test DOCX document.</w:t>
          </w:r>
        </w:p>
      </w:body>
    </w:document>`

	parser := NewDOCXParser()
	result, err := parser.Parse(bytes.NewReader([]byte(content)))

	assert.NoError(t, err)
	assert.Contains(t, string(result), "This is a test DOCX document.")

	metadata := parser.GetFormatMetadata()
	assert.Equal(t, "DOCX", metadata["format"])
	assert.Equal(t, "Test DOCX", metadata["title"])
	assert.Equal(t, "Test Author", metadata["creator"])
}

func TestHTMLParser(t *testing.T) {
	content := `
<!DOCTYPE html>
<html>
<head>
    <title>Test HTML</title>
    <meta name="author" content="Test Author">
</head>
<body>
    <h1>This is a test HTML document.</h1>
</body>
</html>`

	parser := NewHTMLParser()
	result, err := parser.Parse(strings.NewReader(content))

	assert.NoError(t, err)
	assert.Contains(t, string(result), "This is a test HTML document.")

	metadata := parser.GetFormatMetadata()
	assert.Equal(t, "HTML", metadata["format"])
	assert.Equal(t, "Test HTML", metadata["title"])
	assert.Equal(t, "Test Author", metadata["author"])
}

func TestTextParser(t *testing.T) {
	content := "This is a test text document.\nIt has multiple lines.\nAnd some words."

	parser := NewTextParser()
	result, err := parser.Parse(strings.NewReader(content))

	assert.NoError(t, err)
	assert.Equal(t, content, string(result))

	metadata := parser.GetFormatMetadata()
	assert.Equal(t, "Text", metadata["format"])
	assert.Equal(t, "3", metadata["lineCount"])
	assert.Equal(t, "13", metadata["wordCount"])
	assert.Equal(t, "69", metadata["charCount"])
    assert.Equal(t, content+"\n", string(result))
}

func TestMarkdownParser(t *testing.T) {
	content := `---
title: Test Markdown
author: Test Author
---

# This is a test Markdown document

It has some content and metadata.`

	parser := NewMarkdownParser()
	result, err := parser.Parse(strings.NewReader(content))

	assert.NoError(t, err)
	assert.Contains(t, string(result), "This is a test Markdown document")

	metadata := parser.GetFormatMetadata()
	assert.Equal(t, "Markdown", metadata["format"])
	assert.Equal(t, "Test Markdown", metadata["title"])
	assert.Equal(t, "Test Author", metadata["author"])
	assert.Equal(t, "3", metadata["lineCount"])  // Excluding front matter
	assert.Equal(t, "11", metadata["wordCount"]) // Excluding front matter
	assert.Equal(t, "1", metadata["headerCount"])
}

func TestParseDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "parser_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test files
	files := map[string]string{
		"test.txt":  "This is a text file.",
		"test.md":   "# This is a Markdown file",
		"test.html": `<!DOCTYPE html><html><body>This is an HTML file</body></html>`,
	}

	for name, content := range files {
		err = ioutil.WriteFile(filepath.Join(tempDir, name), []byte(content), 0644)
		assert.NoError(t, err)
	}

	// Create a metadata generator
	storage := &mockStorage{} // You need to implement a mock storage
	metadataGen := metadata.NewGenerator(storage)

	// Parse the directory
	results, metadataList, err := ParseDirectory(tempDir, false, metadataGen)
	assert.NoError(t, err)
	assert.Len(t, results, 3)
	assert.Len(t, metadataList, 3)

	// Check if all files were parsed
	for _, result := range results {
		assert.Contains(t, []string{
			"This is a text file.",
			"# This is a Markdown file",
			"This is an HTML file",
		}, strings.TrimSpace(string(result)))
	}

	// Check metadata
	for _, meta := range metadataList {
		assert.NotEmpty(t, meta.SourceFile)
		assert.NotEmpty(t, meta.SHA256Hash)
	}
}

// mockStorage is a mock implementation of the storage.Storage interface
type mockStorage struct{}

func (m *mockStorage) Read(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (m *mockStorage) Write(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0644)
}

func (m *mockStorage) List(prefix string) ([]string, error) {
	return nil, nil
}

func (m *mockStorage) Delete(path string) error {
	return os.Remove(path)
}

func (m *mockStorage) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (m *mockStorage) IsDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func (m *mockStorage) GetReader(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (m *mockStorage) Stat(path string) (storage.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &mockFileInfo{info}, nil
}

type mockFileInfo struct {
	os.FileInfo
}

func (mfi *mockFileInfo) ETag() string {
	return "mock-etag"
}
