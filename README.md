# go-depcheck

[![Go Tests](https://github.com/v-standard/go-depcheck/actions/workflows/go-test.yml/badge.svg)](https://github.com/v-standard/go-depcheck/actions/workflows/go-test.yml)

A static analysis tool for validating package dependencies in Go projects. This tool helps maintain architectural boundaries by detecting and preventing unwanted dependencies between packages. The analyzer checks imports in your Go files against rules defined in a configuration file, ensuring your codebase adheres to the intended architecture.

## Features

The tool provides robust dependency validation through:

- Define forbidden dependency patterns using regular expressions
- Package pattern matching for flexible rule creation
- Multiple ways to exclude dependencies (global patterns, rule-specific patterns, and inline comments)
- Configurable exceptions to forbidden dependencies
- Seamless integration with Go's standard static analysis framework

## Installation

Install the tool using the standard Go tools:

```bash
go install github.com/v-standard/go-depcheck/cmd/depcheck@latest
```

## Usage

### Configuration File

Create a `depcheck.yml` in your project root directory. The configuration file consists of two main sections: global ignore patterns and dependency rules. Here's a basic example:

```yaml
# Global patterns to ignore across all rules
ignorePatterns:
  - 'mock_.*\.go$'
  - '.*_mock\.go$'
  - '_test\.go$'

# Define forbidden dependencies
rules:
  # Prevent core layer from depending on presentation layer
  - from: '^github.com/your-org/project/core/.*$'    # Source package
    to:                                              # Forbidden dependencies
      - '^github.com/your-org/project/presentation.*$'
    allowedDependencies:                             # Exceptions to forbidden dependencies
      - '^github.com/your-org/project/presentation/types.*$'
    ignorePatterns:                                  # Rule-specific patterns to ignore
      - 'special_case_.*\.go$'
```

### Project Integration

Create a `tools.go` file in your project to integrate the tool into your development workflow:

```go
//go:build tools
package tools

import (
    _ "github.com/v-standard/go-depcheck"
)
```

### Execution

You can run the dependency check using Go's built-in vet tool:

```bash
go vet -vettool=$(which depcheck) ./...
```

## Configuration Details

The `depcheck.yml` configuration file allows for both global and rule-specific settings. Let's look at each component in detail:

### Excluding Dependencies

There are three ways to exclude dependencies from checks:

1. **Global Ignore Patterns**: Files that should be ignored across all rules
```yaml
ignorePatterns:
  - 'mock_.*\.go$'
  - '_test\.go$'
```

2. **Rule-Specific Ignore Patterns**: Files to ignore for specific rules
```yaml
rules:
  - from: '...'
    ignorePatterns:
      - 'special_case_.*\.go$'
```

3. **Inline Comments**: Individual imports can be excluded using comments
```go
import "github.com/your-org/project/presentation" // depcheck:allow
```

### Dependency Rules

Each rule in the `rules` section defines forbidden dependencies and their exceptions:

```yaml
rules:
  - from: '^github.com/your-org/project/domain/.*$'    # Source package pattern
    to:                                                # Forbidden package patterns
      - '^github.com/your-org/project/infra/.*$'
      - '^github.com/your-org/project/presentation/.*$'
    allowedDependencies:                               # Exceptions to forbidden dependencies
      - '^github.com/your-org/project/presentation/types.*$'
    ignorePatterns:                                    # Rule-specific files to ignore
      - 'special_case_.*\.go$'
```

1. `from`: Defines the source package pattern where this rule applies
2. `to`: Lists package patterns that are forbidden to import
3. `allowedDependencies`: Specifies exceptions to the forbidden patterns
4. `ignorePatterns`: Defines additional file patterns to ignore for this specific rule

All patterns use Go's regular expression syntax. Package patterns (`from`, `to`, and `allowedDependencies`) are matched against full package paths, while ignore patterns are matched against filenames.

### Pattern Matching Priority

The tool evaluates patterns in the following order:

1. Global ignore patterns
2. Rule-specific ignore patterns
3. Package matching (`from` pattern)
4. Allowed dependencies
5. Forbidden dependencies (`to` patterns)

This hierarchical processing ensures clear and predictable rule evaluation. Files that match any ignore pattern (global or rule-specific) are skipped entirely, making it easy to exclude generated code, test files, or other special cases from dependency checking.

### Example Configuration

Here's a more complete example showing how to structure rules for a typical layered architecture:

```yaml
ignorePatterns:
  - 'mock_.*\.go$'
  - '.*_mock\.go$'
  - '_test\.go$'
  - 'generated_.*\.go$'

rules:
  # Prevent domain layer from depending on infrastructure or presentation
  - from: '^github.com/your-org/project/domain/.*$'
    to:
      - '^github.com/your-org/project/infra/.*$'
      - '^github.com/your-org/project/presentation/.*$'
    allowedDependencies:
      - '^github.com/your-org/project/presentation/types.*$'

  # Prevent presentation layer from depending on infrastructure
  - from: '^github.com/your-org/project/presentation/.*$'
    to:
      - '^github.com/your-org/project/infra/.*$'
    allowedDependencies:
      - '^github.com/your-org/project/infra/common.*$'
    ignorePatterns:
      - 'legacy_.*\.go$'
```

This configuration demonstrates how to:
- Exclude common generated and test files globally
- Define forbidden dependencies between architectural layers
- Allow specific exceptions where cross-layer dependencies are necessary
- Handle special cases with rule-specific ignore patterns

The tool helps ensure your codebase maintains its intended architecture while providing the flexibility to handle real-world development needs through its layered exception handling and ignore patterns.