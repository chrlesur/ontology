package parser

import (
    "os"
    "path/filepath"
    "testing"
)

func TestParsers(t *testing.T) {
    testCases := []struct {
        format string
        content string
    }{
        {".txt", "This is a test text file."},
        {".md", "# This is a Markdown file\n\nWith some content."},
        {".html", "<html><body>This is an HTML file</body></html>"},
    }

    for _, tc := range testCases {
        t.Run(tc.format, func(t *testing.T) {
            tempFile, err := createTempFile(tc.format, tc.content)
            if err != nil {
                t.Fatalf("Failed to create temp file: %v", err)
            }
            defer os.Remove(tempFile)

            parser, err := GetParser(tc.format)
            if err != nil {
                t.Fatalf("Failed to get parser: %v", err)
            }

            content, err := parser.Parse(tempFile)
            if err != nil {
                t.Fatalf("Failed to parse file: %v", err)
            }

            if string(content) != tc.content {
                t.Errorf("Expected content %s, got %s", tc.content, string(content))
            }

            metadata := parser.GetMetadata()
            if len(metadata) == 0 {
                t.Error("Expected non-empty metadata")
            }
        })
    }
}

func TestParseDirectory(t *testing.T) {
    tempDir, err := os.MkdirTemp("", "parser_test")
    if err != nil {
        t.Fatalf("Failed to create temp directory: %v", err)
    }
    defer os.RemoveAll(tempDir)

    files := []struct {
        name string
        content string
    }{
        {"test1.txt", "Text file 1"},
        {"test2.md", "# Markdown file"},
        {"test3.html", "<html><body>HTML file</body></html>"},
    }

    for _, f := range files {
        err := os.WriteFile(filepath.Join(tempDir, f.name), []byte(f.content), 0644)
        if err != nil {
            t.Fatalf("Failed to create test file: %v", err)
        }
    }

    results, err := ParseDirectory(tempDir, false)
    if err != nil {
        t.Fatalf("ParseDirectory failed: %v", err)
    }

    if len(results) != len(files) {
        t.Errorf("Expected %d results, got %d", len(files), len(results))
    }
}

func createTempFile(ext, content string) (string, error) {
    tmpfile, err := os.CreateTemp("", "test*"+ext)
    if err != nil {
        return "", err
    }
    defer tmpfile.Close()

    if _, err := tmpfile.Write([]byte(content)); err != nil {
        return "", err
    }

    return tmpfile.Name(), nil
}