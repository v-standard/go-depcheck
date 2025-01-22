package depcheck

import (
	"go/ast"
	"golang.org/x/tools/go/analysis/analysistest"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzer(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "depcheck-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test configuration file
	configPath := filepath.Join(tmpDir, "depcheck.yml")
	config := []byte(`
rules:
  - from: "example/from.*$"
    to:
      - "example/to.*$"
    exceptions:
      - "example/to/exception.*$"
    ignoreTest: true
`)

	if err := os.WriteFile(configPath, config, 0644); err != nil {
		t.Fatal(err)
	}

	// Set configuration file path via environment variable
	os.Setenv("DEPCHECK_CONFIG", configPath)

	// Prepare test data directory
	testdata := analysistest.TestData()

	// Execute the test
	analysistest.Run(t, testdata, Analyzer, "example/...")
}

// TestExceptionComment tests the exception comment functionality
func TestExceptionComment(t *testing.T) {
	spec := &ast.ImportSpec{
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{
				{Text: "// depcheck:allow temporarily allowed for testing"},
			},
		},
	}

	if !hasExceptionComment(spec) {
		t.Error("Expected hasExceptionComment to return true for valid exception comment")
	}
}
