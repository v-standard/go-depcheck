package depcheck

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const doc = "depcheck checks package dependency rules defined in YAML"

// DependencyRule は1つの依存関係ルールを表します
type DependencyRule struct {
	From       string   `yaml:"from"`       // 依存元パッケージのパターン
	To         []string `yaml:"to"`         // 依存先パッケージのパターン（複数可）
	Exceptions []string `yaml:"exceptions"` // 例外として許可するパスのパターン
}

// Config はYAMLファイルの構造を表します
type Config struct {
	Rules []DependencyRule `yaml:"rules"`
}

// compiledRule はコンパイル済みの正規表現を保持します
type compiledRule struct {
	from       *regexp.Regexp
	to         []*regexp.Regexp
	exceptions []*regexp.Regexp
}

var Analyzer = &analysis.Analyzer{
	Name: "depcheck",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

// コンパイル済みのルールを保持
var compiledRules []compiledRule

func prepare() {
	// 設定ファイルのパスは環境変数から取得するか、デフォルト値を使用
	configPath := os.Getenv("DEPCHECK_CONFIG")
	if configPath == "" {
		configPath = "rules.yml"
	}

	// 設定ファイルを読み込み
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Warning: Could not read config file: %v\n", err)
		return
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		fmt.Printf("Warning: Could not parse config file: %v\n", err)
		return
	}

	// ルールをコンパイル
	compiledRules = make([]compiledRule, 0, len(config.Rules))
	for _, rule := range config.Rules {
		compiled := compiledRule{
			from:       regexp.MustCompile(rule.From),
			to:         make([]*regexp.Regexp, 0, len(rule.To)),
			exceptions: make([]*regexp.Regexp, 0, len(rule.Exceptions)),
		}

		// 依存先パターンをコンパイル
		for _, toPattern := range rule.To {
			compiled.to = append(compiled.to, regexp.MustCompile(toPattern))
		}

		// 例外パターンをコンパイル
		for _, exceptionPattern := range rule.Exceptions {
			compiled.exceptions = append(compiled.exceptions, regexp.MustCompile(exceptionPattern))
		}

		compiledRules = append(compiledRules, compiled)
	}
}

// hasExceptionComment はインポート文に例外を示すコメントがあるかチェックします
func hasExceptionComment(spec *ast.ImportSpec) bool {
	if spec.Doc == nil {
		return false
	}

	for _, comment := range spec.Doc.List {
		if strings.HasPrefix(comment.Text, "// depcheck:allow") {
			return true
		}
	}
	return false
}

// isException は指定されたインポートパスが例外として許可されているかチェックします
func isException(rule compiledRule, importPath string) bool {
	for _, exceptionPattern := range rule.exceptions {
		if exceptionPattern.MatchString(importPath) {
			return true
		}
	}
	return false
}

func run(pass *analysis.Pass) (any, error) {
	prepare()

	pkgpath := pass.Pkg.Path()

	for _, file := range pass.Files {
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				continue
			}

			// コメントによる例外チェック
			if hasExceptionComment(spec) {
				continue
			}

			// 各ルールについてチェック
			for _, rule := range compiledRules {
				if !rule.from.MatchString(pkgpath) {
					continue
				}

				// 設定ファイルによる例外チェック
				if isException(rule, path) {
					continue
				}

				// 依存関係違反のチェック
				for _, toPattern := range rule.to {
					if toPattern.MatchString(path) {
						pass.Reportf(spec.Pos(), "invalid dependency: %s", path)
					}
				}
			}
		}
	}

	return nil, nil
}
