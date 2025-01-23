# go-depcheck

[![Go Tests](https://github.com/v-standard/go-depcheck/actions/workflows/go-test.yml/badge.svg)](https://github.com/v-standard/go-depcheck/actions/workflows/go-test.yml)

A static analysis tool for validating package dependencies in Go projects. This tool helps maintain architectural boundaries by detecting and preventing unwanted dependencies between packages. The analyzer checks imports in your Go files against rules defined in a configuration file, ensuring your codebase adheres to the intended architecture.

## Features

The tool provides robust dependency validation through:

- YAML-based dependency rule definitions that are easy to read and maintain
- Package pattern matching using regular expressions for flexible rule creation
- Multi-level file exclusion patterns (global and rule-specific)
- Configurable dependency allowances for architectural exceptions
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
  - "mock_.*.go$"
  - ".*_mock.go$"
  - "_test.go$"

# Dependency rules
rules:
  - from: "^github.com/your-org/project/core/.*$"
    to: 
      - "^github.com/your-org/project/presentation.*$"
    allowedDependencies:
      - "^github.com/your-org/project/presentation/types.*$"
    ignorePatterns:  # Rule-specific patterns to ignore
      - "special_case_.*.go$"
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

The `depcheck.yml` configuration file uses a straightforward structure that allows for both global and rule-specific settings. Let's look at each component in detail:

### Global Ignore Patterns

At the root level, you can define patterns for files that should be ignored across all rules:

```yaml
ignorePatterns:
  - "mock_.*.go$"      # Generated mock files
  - ".*_mock.go$"      # Alternative mock file naming
  - "_test.go$"        # Test files
```

These patterns are applied first, before any rule-specific checks. Files matching these patterns will be completely excluded from dependency analysis.

### Dependency Rules

Each rule in the `rules` section consists of four components:

```yaml
rules:
  - from: "^github.com/your-org/project/domain/.*$"    # Source package pattern
    to:                                                # Restricted package patterns
      - "^github.com/your-org/project/infra/.*$"
      - "^github.com/your-org/project/presentation/.*$"
    allowedDependencies:                               # Exceptions to dependency restrictions
      - "^github.com/your-org/project/presentation/types.*$"
    ignorePatterns:                                    # Rule-specific files to ignore
      - "special_case_.*.go$"
```

1. `from`: Defines the source package pattern where this rule applies. The pattern is matched against the package containing the import statement being checked.

2. `to`: Lists package patterns that represent restricted dependencies. When an import matches any of these patterns, it will be flagged as a violation unless explicitly allowed.

3. `allowedDependencies`: Specifies exceptions to the restricted patterns. If an import matches both a `to` pattern and an `allowedDependencies` pattern, it will be allowed.

4. `ignorePatterns`: Defines additional file patterns to ignore, specific to this rule. These patterns are checked after the global ignore patterns.

All patterns use Go's regular expression syntax. Package patterns (`from`, `to`, and `allowedDependencies`) are matched against full package paths, while ignore patterns are matched against filenames.

### Pattern Matching Priority

The tool evaluates patterns in the following order:

1. Global ignore patterns
2. Rule-specific ignore patterns
3. Package matching (`from` pattern)
4. Allowed dependencies
5. Restricted dependencies (`to` patterns)

This hierarchical processing ensures clear and predictable rule evaluation. Files that match any ignore pattern (global or rule-specific) are skipped entirely, making it easy to exclude generated code, test files, or other special cases from dependency checking.

### Example Configuration

Here's a more complete example showing how to structure rules for a typical layered architecture:

```yaml
ignorePatterns:
  - "mock_.*.go$"
  - ".*_mock.go$"
  - "_test.go$"
  - "generated_.*.go$"

rules:
  # Domain layer dependencies
  - from: "^github.com/your-org/project/domain/.*$"
    to:
      - "^github.com/your-org/project/infra/.*$"
      - "^github.com/your-org/project/presentation/.*$"
    allowedDependencies:
      - "^github.com/your-org/project/presentation/types.*$"

  # Presentation layer dependencies
  - from: "^github.com/your-org/project/presentation/.*$"
    to:
      - "^github.com/your-org/project/infra/.*$"
    allowedDependencies:
      - "^github.com/your-org/project/infra/common.*$"
    ignorePatterns:
      - "legacy_.*.go$"
```

This configuration demonstrates how to:
- Exclude common generated and test files globally
- Enforce architectural boundaries between layers
- Allow specific cross-layer dependencies where necessary
- Handle special cases with rule-specific ignore patterns

The tool helps ensure your codebase maintains its intended architecture while providing the flexibility to handle real-world development needs through its layered exception handling and ignore patterns.