package depcheck

import (
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
ignorePatterns:
  - '.*_mock.go$'
rules:
  - from: 'example/from.*$'
    to:
      - 'example/to.*$'
    allowedDependencies:
      - 'example/to/exception.*$'
    ignorePatterns:
      - '.*_test\.go$'
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
