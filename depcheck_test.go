package depcheck

import (
	"go/ast"
	"golang.org/x/tools/go/analysis/analysistest"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzer(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "depcheck-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// テスト用の設定ファイルを作成
	configPath := filepath.Join(tmpDir, "rules.yml")
	config := []byte(`
rules:
  - from: "example/from/.*$"
    to:
      - "example/to.*$"
    exceptions:
      - "example/to/exception.*$"
`)

	if err := os.WriteFile(configPath, config, 0644); err != nil {
		t.Fatal(err)
	}

	// 環境変数で設定ファイルのパスを指定
	os.Setenv("DEPCHECK_CONFIG", configPath)

	// テストデータのディレクトリを用意
	testdata := analysistest.TestData()

	// テストを実行
	analysistest.Run(t, testdata, Analyzer, "example/...")
}

// TestExceptionComment はコメントによる例外機能をテストします
func TestExceptionComment(t *testing.T) {
	spec := &ast.ImportSpec{
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{
				{Text: "// depcheck:allow テスト用に一時的に許可"},
			},
		},
	}

	if !hasExceptionComment(spec) {
		t.Error("Expected hasExceptionComment to return true for valid exception comment")
	}
}
