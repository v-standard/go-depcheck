# go-depcheck

A static analysis tool for validating package dependencies in Go projects. This tool helps maintain architectural boundaries by detecting and preventing unwanted dependencies between packages. The analyzer checks imports in your Go files against rules defined in a configuration file, ensuring your codebase adheres to the intended architecture.

## Features

The tool provides robust dependency validation through:

- YAML-based dependency rule definitions that are easy to read and maintain
- Package pattern matching using regular expressions for flexible rule creation
- Fine-grained control over which files to analyze or ignore
- Configurable dependency allowances for architectural exceptions
- Seamless integration with Go's standard static analysis framework

## Installation

Install the tool using the standard Go tools:

```bash
go install github.com/v-standard/go-depcheck@latest
```

## Usage

### Configuration File

Create a `depcheck.yml` in your project root directory. This file defines the dependency rules for your project. Here's a basic example:

```yaml
rules:
  - from: "^github.com/your-org/project/core/.*$"
    to: 
      - "^github.com/your-org/project/presentation.*$"
    allowedDependencies:
      - "^github.com/your-org/project/presentation/types.*$"
    ignorePatterns:
      - "mock_.*.go$"
      - "_test.go$"
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

The `depcheck.yml` configuration file uses a straightforward structure to define dependency rules. Each rule consists of several key components:

```yaml
rules:
  # Prevent domain layer from depending on infrastructure
  - from: "^github.com/your-org/project/domain/.*$"    # Source package pattern
    to:                                                # Restricted package patterns
      - "^github.com/your-org/project/infra/.*$"
      - "^github.com/your-org/project/presentation/.*$"
    allowedDependencies:                               # Exceptions to dependency restrictions
      - "^github.com/your-org/project/presentation/types.*$"
    ignorePatterns:                                    # Files to exclude from analysis
      - "mock_.*.go$"      # Generated mock files
      - ".*_mock.go$"      # Alternative mock file naming
      - "_test.go$"        # Test files
```

Let's break down each component:

1. `from`: Specifies the package pattern where this rule applies. This pattern matches against the package that contains the import statement being checked.

2. `to`: Lists package patterns that should be restricted. When an import matches any of these patterns, it will be flagged as a violation unless allowed by other rules.

3. `allowedDependencies`: Defines exceptions to the restricted patterns. If an import matches both a `to` pattern and an `allowedDependencies` pattern, it will be allowed. This is useful for permitting specific cross-layer dependencies that are architecturally acceptable.

4. `ignorePatterns`: Specifies patterns for files that should be completely excluded from analysis. These patterns match against filenames, making it easy to skip generated files, test files, or any other special cases. This is particularly useful for:
    - Generated mock files (which often need to import the mocked dependencies)
    - Test files that may need more relaxed dependency rules
    - Any other generated or special-purpose files that shouldn't be subject to the normal dependency rules

All patterns use Go's regular expression syntax, providing powerful and flexible matching capabilities. Patterns are matched against full package paths for `from`, `to`, and `allowedDependencies`, and against filenames for `ignorePatterns`.

The tool processes these rules in the following order:
1. First, it checks if the file matches any `ignorePatterns` - if so, it's skipped entirely
2. Then, it checks if the importing package matches the `from` pattern
3. If an import matches any `to` pattern, it's checked against `allowedDependencies`
4. If the import isn't in `allowedDependencies`, it's reported as a violation

This hierarchical processing ensures clear and predictable rule evaluation, making it easier to understand and debug dependency violations.