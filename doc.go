// Package pgtestkit provides an easy way to create and manage test PostgreSQL databases
// for Go tests. It automatically handles the lifecycle of an embedded PostgreSQL
// server and provides utilities for database setup and teardown.
//
// Key Features:
//   - 🚀 Automatic embedded PostgreSQL server management
//   - 🧪 Isolated test databases for each test case
//   - ⚡️ High performance with server reuse
//   - 🔄 Automatic resource cleanup
//   - 🛠️ Interface-based architecture supporting any database driver or ORM
//   - 📦 Built-in support for GORM (see example/gorm)
//
// Basic Usage with database/sql:
//
//	func TestWithSQL(t *testing.T) {
//	    // Create a test database with the default SQL connector
//	    dbClient, err := pgtestkit.CreateTestDB(nil)
//	    if err != nil {
//	        t.Fatalf("Failed to create test DB: %v", err)
//	    }
//	    defer dbClient.Close()
//
//	    // Get the *sql.DB instance
//	    db := dbClient.Client.(*sql.DB)
//	    // ... your test code ...
//	}
//
// Using with GORM (see example/gorm for more details):
//
//	import "github.com/tidylogic/pgtestkit/example/gorm"
//
//	func TestWithGORM(t *testing.T) {
//	    // Create a test database with GORM connector
//	    dbClient, err := pgtestkit.CreateTestDB(&gorm.GORMConnector{})
//	    if err != nil {
//	        t.Fatalf("Failed to create test DB: %v", err)
//	    }
//	    defer dbClient.Close()
//
//	    // Get the *gorm.DB instance
//	    gormDB := dbClient.Client.(*gorm.DB)
//	    // ... your test code ...
//	}
//
// For more examples and advanced usage, see the example directory.
package pgtestkit
