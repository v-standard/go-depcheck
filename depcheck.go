package depcheck

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const doc = "depcheck checks package dependency rules defined in YAML"

// Config represents the structure of the YAML configuration file
type Config struct {
	IgnorePatterns []string         `yaml:"ignorePatterns"` // Global patterns to ignore
	Rules          []DependencyRule `yaml:"rules"`
}

// DependencyRule represents a single dependency rule
type DependencyRule struct {
	From                string   `yaml:"from"`                // Source package pattern
	To                  []string `yaml:"to"`                  // Target package patterns (multiple allowed)
	AllowedDependencies []string `yaml:"allowedDependencies"` // Patterns for allowed dependencies
	IgnorePatterns      []string `yaml:"ignorePatterns"`      // Patterns for files to exclude from analysis

}

// compiledRule holds the compiled regular expressions for rule matching
type compiledRule struct {
	from                *regexp.Regexp
	to                  []*regexp.Regexp
	allowedDependencies []*regexp.Regexp
	ignorePatterns      []*regexp.Regexp
}

var Analyzer = &analysis.Analyzer{
	Name:     "depcheck",
	Doc:      doc,
	Run:      run,
	Requires: []*analysis.Analyzer{},
}

// Variables to hold compiled rules and manage initialization state with mutex
var (
	compiledRules          []compiledRule
	compiledIgnorePatterns []*regexp.Regexp
	prepareOnce            = sync.OnceValue(prepare)
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

	// Compile global ignore patterns
	compiledIgnorePatterns = make([]*regexp.Regexp, 0, len(config.IgnorePatterns))
	for _, pattern := range config.IgnorePatterns {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid ignore pattern %q: %v", pattern, err)
		}
		compiledIgnorePatterns = append(compiledIgnorePatterns, compiled)
	}

	// Compile rules
	compiledRules = make([]compiledRule, 0, len(config.Rules))
	for _, rule := range config.Rules {
		compiled := compiledRule{
			from:                regexp.MustCompile(rule.From),
			to:                  make([]*regexp.Regexp, 0, len(rule.To)),
			allowedDependencies: make([]*regexp.Regexp, 0, len(rule.AllowedDependencies)),
			ignorePatterns:      make([]*regexp.Regexp, 0, len(rule.IgnorePatterns)),
		}

		// Compile target patterns
		for _, toPattern := range rule.To {
			compiled.to = append(compiled.to, regexp.MustCompile(toPattern))
		}

		// Compile allowed dependency patterns
		for _, allowedPattern := range rule.AllowedDependencies {
			compiled.allowedDependencies = append(compiled.allowedDependencies, regexp.MustCompile(allowedPattern))
		}

		// Compile ignore patterns
		for _, ignorePattern := range rule.IgnorePatterns {
			compiled.ignorePatterns = append(compiled.ignorePatterns, regexp.MustCompile(ignorePattern))
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
	if strings.HasPrefix(spec.Comment.Text(), "depcheck:allow") {
		return true
	}
	return false
}

// shouldIgnore checks if a file should be ignored based on ignore patterns
func shouldIgnoreFile(filename string, globalPatterns []*regexp.Regexp, rulePatterns []*regexp.Regexp) bool {
	for _, pattern := range globalPatterns {
		if pattern.MatchString(filename) {
			return true
		}
	}

	for _, pattern := range rulePatterns {
		if pattern.MatchString(filename) {
			return true
		}
	}

	return false
}

// isAllowed checks if an import path is allowed as an exception
func isAllowed(rule compiledRule, importPath string) bool {
	for _, allowedPattern := range rule.allowedDependencies {
		if allowedPattern.MatchString(importPath) {
			return true
		}
	}
	return false
}

func run(pass *analysis.Pass) (any, error) {
	// Execute initialization only once
	err := prepareOnce()
	if err != nil {
		return nil, err
	}

	pkgpath := pass.Pkg.Path()

	for _, file := range pass.Files {
		pos := pass.Fset.Position(file.Pos())
		filename := filepath.Base(pos.Filename)

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

				// Skip files matching ignore patterns
				if shouldIgnoreFile(filename, compiledIgnorePatterns, rule.ignorePatterns) {
					continue
				}

				// Check for allowed dependencies
				if isAllowed(rule, path) {
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
