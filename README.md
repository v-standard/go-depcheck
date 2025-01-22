# go-depcheck

A static analysis tool for validating package dependencies in Go projects. This tool helps maintain architectural boundaries by detecting and preventing unwanted dependencies between packages. The analyzer checks imports in your Go files against rules defined in a configuration file, ensuring your codebase adheres to the intended architecture.

## Features

The tool provides robust dependency validation through:

- YAML-based dependency rule definitions that are easy to read and maintain
- Package pattern matching using regular expressions for flexible rule creation
- Rule-specific control over test file analysis
- Configuration-based exceptions for handling necessary architectural deviations
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
    exceptions:
      - "^github.com/your-org/project/presentation/types.*$"
    ignoreTest: true  # Skip checking test files for this rule
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

The `depcheck.yml` configuration file uses a straightforward structure to define dependency rules. Each rule consists of a source pattern ("from"), target patterns ("to"), optional exceptions, and test file handling settings. Here's a comprehensive example:

```yaml
rules:
  # Prevent domain layer from depending on infrastructure, including test files
  - from: "^github.com/your-org/project/domain/.*$"    # Source package pattern
    to:                                                # Restricted package patterns
      - "^github.com/your-org/project/infra/.*$"
      - "^github.com/your-org/project/presentation/.*$"
    exceptions:                                        # Allowed exception patterns
      - "^github.com/your-org/project/presentation/types.*$"
    ignoreTest: false                                 # Include test files in checks

  # Restrict presentation layer dependencies, excluding test files
  - from: "^github.com/your-org/project/presentation/.*$"
    to:
      - "^github.com/your-org/project/infrastructure/.*$"
    exceptions:
      - "^github.com/your-org/project/infrastructure/common.*$"
    ignoreTest: true                                  # Skip test files for this rule
```

The rules use regular expressions for pattern matching, allowing for flexible and powerful dependency control. The analyzer will report violations when it finds imports that match the "to" patterns but aren't covered by the exceptions.

In this configuration:
- The "from" field specifies which packages the rule applies to
- The "to" field lists patterns for restricted imports
- The "exceptions" field defines patterns for allowed exceptions to the rule
- The "ignoreTest" field controls whether the rule applies to test files (files ending with `_test.go`)

The test file handling feature (`ignoreTest`) allows you to:
- Set different dependency rules for test files and production code
- Skip dependency checks for test files where more relaxed rules might be appropriate
- Maintain stricter control over production code dependencies while allowing necessary testing patterns
