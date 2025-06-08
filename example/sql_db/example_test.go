package pgtestkit_test

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tidylogic/pgtestkit"
)

// ExampleConnector는 DBConnector 인터페이스의 예제 구현체입니다.
type ExampleConnector struct {
	db *sql.DB
}

func (e *ExampleConnector) Connect(connString string) (interface{}, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	e.db = db
	return db, nil
}

func (e *ExampleConnector) Close() error {
	if e.db != nil {
		return e.db.Close()
	}
	return nil
}

// Reset 데이터베이스의 모든 테이블을 TRUNCATE하여 초기 상태로 되돌립니다.
func (e *ExampleConnector) Reset() error {
	if e.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 현재 데이터베이스의 모든 테이블 조회
	rows, err := e.db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		AND table_name != 'schema_migrations'`)
	if err != nil {
		return fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	// 각 테이블을 TRUNCATE
	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, fmt.Sprintf(`"%s"`, tableName))
	}

	if len(tables) > 0 {
		truncateSQL := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", strings.Join(tables, ", "))
		if _, err := e.db.Exec(truncateSQL); err != nil {
			return fmt.Errorf("failed to truncate tables: %w", err)
		}
	}

	return rows.Err()
}

func TestExample(t *testing.T) {
	t.Run("Database operations", func(t *testing.T) {
		// 1. 테스트 DB 클라이언트 생성
		dbClient, err := pgtestkit.CreateTestDB(&ExampleConnector{})
		if err != nil {
			t.Fatalf("Failed to create test DB: %v", err)
		}
		defer func() {
			_ = dbClient.Close() // 에러 무시 (테스트 종료 시)
		}()

		// 테스트 헬퍼 생성
		helper := pgtestkit.NewTestHelper(dbClient)
		defer func() {
			_ = helper.Close() // 에러 무시 (테스트 종료 시)
		}()

		// 데이터베이스 클라이언트 가져오기
		db, ok := dbClient.Client.(*sql.DB)
		if !ok {
			t.Fatal("Expected *sql.DB client")
		}

		t.Run("Test database operations", func(t *testing.T) {
			// 테이블 초기화
			helper.MustResetDB(t)

			// 여기에 테스트 코드 작성
			_, err = db.Exec("CREATE TABLE IF NOT EXISTS test (id SERIAL PRIMARY KEY, name TEXT)")
			if err != nil {
				t.Fatalf("Failed to create table: %v", err)
			}

			// 테스트 데이터 삽입
			_, err = db.Exec("INSERT INTO test (name) VALUES ($1)", "test")
			if err != nil {
				t.Fatalf("Failed to insert test data: %v", err)
			}

			// 데이터 조회
			var name string
			err = db.QueryRow("SELECT name FROM test LIMIT 1").Scan(&name)
			if err != nil {
				t.Fatalf("Failed to query data: %v", err)
			}

			if name != "test" {
				t.Errorf("Expected name 'test', got '%s'", name)
			}
		})

		t.Run("Test isolation", func(t *testing.T) {
			// 테이블 초기화
			helper.MustResetDB(t)

			// 다른 테스트에서도 데이터베이스가 깨끗한지 확인
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
			if err != nil {
				t.Fatalf("Failed to count rows: %v", err)
			}

			// 이전 테스트의 데이터가 남아있지 않아야 함
			if count != 0 {
				t.Errorf("Expected 0 rows, got %d", count)
			}
		})
	})
}

func TestMain(m *testing.M) {
	// 모든 테스트에 대해 DB 서버 자동 관리
	os.Exit(pgtestkit.TestMainWrapper(m))
}
