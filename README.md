<div align="center">

# 🚀 pgtestkit

[한국어](README-KO.md) | [English](README.md)

[![Go Reference](https://pkg.go.dev/badge/github.com/yourusername/pgtestkit.svg)](https://pkg.go.dev/github.com/yourusername/pgtestkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/pgtestkit)](https://goreportcard.com/report/github.com/yourusername/pgtestkit)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/yourusername/pgtestkit/blob/main/LICENSE)
[![GitHub release](https://img.shields.io/github/release/yourusername/pgtestkit.svg)](https://github.com/yourusername/pgtestkit/releases)
[![Build Status](https://github.com/yourusername/pgtestkit/actions/workflows/test.yml/badge.svg)](https://github.com/yourusername/pgtestkit/actions)
[![codecov](https://codecov.io/gh/yourusername/pgtestkit/branch/main/graph/badge.svg)](https://codecov.io/gh/yourusername/pgtestkit)

</div>

A lightweight PostgreSQL database management tool for Go testing. It helps you easily run and manage an embedded PostgreSQL server in your test environment.

## ✨ Features

- **PostgreSQL Specialized**: Optimized testing tool specifically for PostgreSQL
- **Zero Configuration**: Ready to use with default settings
- **Fast Test Execution**: Reuses embedded server for faster test runs
- **Complete Isolation**: Ensures database isolation between tests
- **Flexible Architecture**: Interface-based design works with any ORM or driver
- **GORM Support**: Built-in GORM connector (see example/gorm)
- **database/sql Support**: Fully compatible with the standard library

## Code Quality

This project uses [golangci-lint](https://golangci-lint.run/) to maintain code quality.

### Running Linters

```bash
# Run linters
golangci-lint run

# Fix auto-fixable issues
golangci-lint run --fix
```

Or use the Makefile:

```bash
# Format code and run linters
make all
```

### Included Linters

- **gofmt/goimports**: Code formatting
- **govet**: Go vet checks
- **staticcheck**: Static analysis
- **errcheck**: Error check verification
- **gocritic**: Code quality suggestions
- And other useful linters

## Key Features

- 🚀 Automatically starts/stops embedded PostgreSQL server during tests
- 🧪 Ensures test isolation with automatic database reset between tests
- ⚡️ High performance - server starts only once and is reused across tests
- 🔄 Automatic resource cleanup
- 🛠️ Supports both GORM and standard database/sql interfaces
- 🔌 Customizable database settings

## Installation

```bash
go get github.com/tidylogic/pgtestkit
```

### Using with GORM

To use with GORM, reference the `example/gorm` package:

```bash
go get github.com/tidylogic/pgtestkit/example/gorm@latest
```

See the [example/gorm](example/gorm) directory for detailed examples.

## Quick Start

### Basic Usage (database/sql)

```go
package yourpackage_test

import (
    "database/sql"
    "testing"

    "github.com/tidylogic/pgtestkit"
)

func TestWithSQL(t *testing.T) {
    // Create test DB with default SQL connector
    dbClient, err := pgtestkit.CreateTestDB(nil)
    if err != nil {
        t.Fatalf("Failed to create test DB: %v", err)
    }
    defer dbClient.Close()

    // Get *sql.DB instance
    db := dbClient.Client.(*sql.DB)
    
    // Write your test code here...
}
```

### Using with GORM

```go
package yourpackage_test

import (
    "testing"

    "github.com/tidylogic/pgtestkit"
    "github.com/tidylogic/pgtestkit/example/gorm"
    "gorm.io/gorm"
)

func TestWithGORM(t *testing.T) {
    // Create test DB with GORM connector
    dbClient, err := pgtestkit.CreateTestDB(&gorm.GORMConnector{})
    if err != nil {
        t.Fatalf("Failed to create test DB: %v", err)
    }
    defer dbClient.Close()

    // Get *gorm.DB instance
    gormDB := dbClient.Client.(*gorm.DB)
    
    // Write your test code using GORM...
}
```

### Using with TestMain

```go
package yourpackage_test

import (
    "os"
    "testing"
    "github.com/tidylogic/pgtestkit"
)

func TestMain(m *testing.M) {
    // Automatically manages DB server before/after tests
    os.Exit(pgtestkit.TestMainWrapper(m))
}
```

### Using Test Helpers

```go
func TestWithHelper(t *testing.T) {
    dbClient, err := pgtestkit.CreateTestDB()
    if err != nil {
        t.Fatalf("Failed to create test DB: %v", err)
    }
    defer dbClient.Close()
    
    // Create test helper
    helper := pgtestkit.NewTestHelper(dbClient.GormDB)
    defer helper.Close()
    
    t.Run("Test case 1", func(t *testing.T) {
        // Reset database before test case
        helper.MustResetDB(t)
        
        // Test code...
    })
    
    t.Run("Test case 2", func(t *testing.T) {
        // Another test case (completely isolated from previous test)
        helper.MustResetDB(t)
        
        // Another test code...
    })
}
```

## Advanced Usage

### Implementing Custom Connector

You can create your own database connector by implementing the `DBConnector` interface:

```go
package yourpackage

type YourConnector struct {
    // Store connector state
}

func (c *YourConnector) Connect(connString string) (interface{}, error) {
    // Implement database connection logic
    // Return type depends on your ORM/driver
    return yourDBClient, nil
}
```

[View in Korean](README-KO.md) | [View in English](README.md)
