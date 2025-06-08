# GORM Example

[한국어](README-KO.md) | [English](README.md)

This example demonstrates how to perform PostgreSQL database testing with GORM using pgtestkit.

## Before You Begin

Install the required dependencies:

```bash
cd ..
go mod tidy
```

## Running Tests

```bash
# Run all tests
cd ..
go test -v ./...

# Run specific test
cd ..
go test -v -run TestGORMExample
```

## Example Overview

This example demonstrates the following features:

1. Starting and stopping an embedded PostgreSQL server
2. Managing database connections with GORM
3. Resetting the database between tests
4. Basic CRUD operation testing
5. Test isolation verification

## Custom Model

The example uses a `User` model for testing. You can modify or add models as needed.

```go
type User struct {
    ID        uint           `gorm:"primaryKey"`
    Name      string
    Email     string         `gorm:"uniqueIndex"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

## Notes

- The embedded PostgreSQL server starts automatically before running tests.
- Each test runs independently, and the database is automatically reset between tests.
- The embedded server shuts down automatically when tests complete.

[View in Korean](README-KO.md) | [View in English](README.md)
