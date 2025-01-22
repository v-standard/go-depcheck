package depcheck

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const doc = "depcheck checks package dependency rules defined in YAML"

// DependencyRule represents a single dependency rule
type DependencyRule struct {
	From       string   `yaml:"from"`       // Source package pattern
	To         []string `yaml:"to"`         // Target package patterns (multiple allowed)
	Exceptions []string `yaml:"exceptions"` // Package patterns allowed as exceptions
}

// Config represents the structure of the YAML configuration file
type Config struct {
	Rules []DependencyRule `yaml:"rules"`
}

// compiledRule holds the compiled regular expressions for rule matching
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

// Variables to hold compiled rules and manage initialization state with mutex
var (
	compiledRules []compiledRule
	initOnce      sync.Once
)

func prepare() error {
	configPath := "depcheck.yml"
	if envPath := os.Getenv("DEPCHECK_CONFIG"); envPath != "" {
		configPath = envPath
	}

	// Search for configuration file
	foundPath, err := findConfigFile(configPath)
	if err != nil {
		return fmt.Errorf("could not find config file: %w", err)
	}

	// Read configuration file
	data, err := os.ReadFile(foundPath)
	if err != nil {
		return fmt.Errorf("warning: Could not read config file: %v\n", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("warning: Could not parse config file: %v\n", err)
	}

	// Compile rules
	compiledRules = make([]compiledRule, 0, len(config.Rules))
	for _, rule := range config.Rules {
		compiled := compiledRule{
			from:       regexp.MustCompile(rule.From),
			to:         make([]*regexp.Regexp, 0, len(rule.To)),
			exceptions: make([]*regexp.Regexp, 0, len(rule.Exceptions)),
		}

		// Compile target patterns
		for _, toPattern := range rule.To {
			compiled.to = append(compiled.to, regexp.MustCompile(toPattern))
		}

		// Compile exception patterns
		for _, exceptionPattern := range rule.Exceptions {
			compiled.exceptions = append(compiled.exceptions, regexp.MustCompile(exceptionPattern))
		}

		compiledRules = append(compiledRules, compiled)
	}

	return nil
}

func findConfigFile(configPath string) (string, error) {
	// 1. Check path specified by environment variable
	if envPath := os.Getenv("DEPCHECK_CONFIG"); envPath != "" {
		return envPath, nil
	}

	// 2. Search upwards from current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		path := filepath.Join(dir, configPath)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("config file %s not found in any parent directory", configPath)
}

// hasExceptionComment checks if an import statement has an exception comment
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

// isException checks if an import path is allowed as an exception
func isException(rule compiledRule, importPath string) bool {
	for _, exceptionPattern := range rule.exceptions {
		if exceptionPattern.MatchString(importPath) {
			return true
		}
	}
	return false
}

func run(pass *analysis.Pass) (any, error) {
	// Execute initialization only once
	var initError error
	initOnce.Do(func() {
		initError = prepare()
	})

	// Return immediately if initialization failed
	if initError != nil {
		return nil, initError
	}

	pkgpath := pass.Pkg.Path()

	for _, file := range pass.Files {
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				continue
			}

			// Check for exception comments
			if hasExceptionComment(spec) {
				continue
			}

			// Check against each rule
			for _, rule := range compiledRules {
				if !rule.from.MatchString(pkgpath) {
					continue
				}

				// Check for configuration-based exceptions
				if isException(rule, path) {
					continue
				}

				// Check for dependency violations
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
