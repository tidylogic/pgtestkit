package pgtestkit

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/phayes/freeport"
	"go.uber.org/zap"
)

const (
	// 데이터베이스 기본 설정
	DefaultUser     = "postgres"
	DefaultPassword = "postgres"
	DefaultDB       = "postgres"
	DefaultLocale   = "en_US.UTF-8"
	TestDBPrefix    = "testdb_"
)

var (
	// 임베디드 PostgreSQL 서버 관련 변수
	serverOnce     sync.Once
	server         *embeddedpostgres.EmbeddedPostgres
	baseDBClient   *sql.DB
	port           uint32
	cacheDirectory string
	serverStarted  bool
	serverStopped  bool
	serverMutex    sync.Mutex
)

// DBConnector 데이터베이스 연결을 관리하는 인터페이스입니다.
type DBConnector interface {
	// Connect 데이터베이스에 연결하고, ORM 클라이언트와 SQL DB를 반환합니다.
	Connect(connString string) (interface{}, error)

	// Close 데이터베이스 연결을 종료합니다.
	Close() error

	// Reset 데이터베이스를 초기 상태로 되돌립니다.
	Reset() error
}

// DBClient 데이터베이스 클라이언트를 나타냅니다.
type DBClient struct {
	Client           interface{} // 모든 종류의 DB 클라이언트를 저장
	DBName           string
	ConnectionString string
	connector        DBConnector
}

// StartEmbeddedPostgres 임베디드 PostgreSQL 서버를 시작합니다.
// 이 함수는 스레드 안전하며, 여러 번 호출되어도 서버는 한 번만 시작됩니다.
func StartEmbeddedPostgres(dbConfig *embeddedpostgres.Config) error {
	var startErr error

	serverOnce.Do(func() {
		serverMutex.Lock()
		defer serverMutex.Unlock()

		logger := getLogger()
		logger.Info("Starting embedded PostgreSQL server")

		// 이미 서버가 중지된 경우
		if serverStopped {
			err := fmt.Errorf("server has been stopped and cannot be restarted")
			logError("Failed to start server", err)
			startErr = err
			return
		}

		pg, p, err := startPostgresServer(dbConfig)
		if err != nil {
			logError("Failed to start PostgreSQL server", err)
			startErr = fmt.Errorf("failed to start postgres server: %w", err)
			return
		}

		server = pg
		port = p
		logger.Info("PostgreSQL server started", zap.Uint32("port", port))

		// 기본 데이터베이스에 연결
		db, err := sql.Open("pgx", getConnectionString(DefaultDB))
		if err != nil {
			logError("Failed to connect to database", err,
				zap.String("database", DefaultDB),
				zap.Uint32("port", port))
			if stopErr := pg.Stop(); stopErr != nil {
				logError("Failed to stop PostgreSQL server after connection error", stopErr)
			}
			startErr = fmt.Errorf("failed to connect to database: %w", err)
			return
		}

		if err := db.Ping(); err != nil {
			logError("Failed to ping database", err)
			if closeErr := db.Close(); closeErr != nil {
				logError("Failed to close database connection", closeErr)
			}
			if stopErr := pg.Stop(); stopErr != nil {
				logError("Failed to stop PostgreSQL server after ping error", stopErr)
			}
			startErr = fmt.Errorf("failed to ping database: %w", err)
			return
		}

		baseDBClient = db
		serverStarted = true
		logger.Info("Successfully connected to PostgreSQL server")

		// SIGINT, SIGTERM 시그널 처리
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			sig := <-c
			logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))
			if err := StopPostgres(); err != nil {
				logError("Error during shutdown", err)
			}
			os.Exit(0)
		}()
	})

	return startErr
}

// Close 데이터베이스 연결을 안전하게 종료하고 테스트 데이터베이스를 삭제합니다.
// 이 메서드는 여러 번 호출해도 안전합니다.
//
// 일반적으로 테스트 함수 시작 시 defer와 함께 사용합니다:
//
//	func TestSomething(t *testing.T) {
//	    dbClient, err := CreateTestDB()
//	    if err != nil {
//	        t.Fatalf("Failed to create test DB: %v", err)
//	    }
//	    defer func() {
//	        if err := dbClient.Close(); err != nil {
//	            t.Logf("Warning: error closing database client: %v", err)
//	        }
//	    }()
//	    // ... 테스트 코드 ...
//	}
func (c *DBClient) Close() error {
	if c == nil {
		return nil
	}

	logger := getLogger().With(zap.String("database", c.DBName))
	logger.Info("Closing database client and cleaning up resources")

	var errs []error

	// 커넥터를 통한 정리
	if c.connector != nil {
		logger.Debug("Closing database connector")
		if err := c.connector.Close(); err != nil {
			err = fmt.Errorf("connector close error: %w", err)
			logger.Error("Failed to close database connector", zap.Error(err))
			errs = append(errs, err)
		} else {
			logger.Debug("Successfully closed database connector")
		}
	}

	// 데이터베이스 삭제
	if c.DBName != "" {
		logger.Debug("Dropping test database", zap.String("database", c.DBName))
		if err := dropDatabase(c.DBName); err != nil {
			err = fmt.Errorf("failed to drop database %s: %w", c.DBName, err)
			logger.Error("Failed to drop test database", zap.Error(err))
			errs = append(errs, err)
		} else {
			logger.Info("Successfully dropped test database", zap.String("database", c.DBName))
		}
	}

	if len(errs) > 0 {
		err := fmt.Errorf("%d error(s) occurred while closing DB client: %+v", len(errs), errs)
		logger.Error("Errors occurred during database client close", zap.Errors("errors", errs))
		return err
	}

	logger.Info("Successfully closed database client and cleaned up all resources")
	return nil
}

// getConnectionString 데이터베이스 연결 문자열을 생성합니다.
func getConnectionString(dbName string) string {
	return fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable&client_encoding=UTF8",
		DefaultUser, DefaultPassword, port, dbName)
}

// startPostgresServer PostgreSQL 서버를 시작합니다.
func startPostgresServer(dbConfig *embeddedpostgres.Config) (*embeddedpostgres.EmbeddedPostgres, uint32, error) {
	freePort, err := freeport.GetFreePort()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get free port: %w", err)
	}

	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user home directory: %w", err)
	}

	cacheDirectory = filepath.Join(userHome, ".embedded-postgres-go", fmt.Sprintf(DefaultDB+"_%d", freePort))

	// 캐시 디렉토리 생성
	if err := os.MkdirAll(cacheDirectory, 0o755); err != nil {
		return nil, 0, fmt.Errorf("failed to create cache directory: %w", err)
	}

	var config embeddedpostgres.Config
	if dbConfig == nil {
		config = embeddedpostgres.DefaultConfig().
			Username(DefaultUser).
			Password(DefaultPassword).
			Database(DefaultDB).
			Version(embeddedpostgres.V15).
			Port(uint32(freePort)).
			RuntimePath(cacheDirectory).
			DataPath(filepath.Join(cacheDirectory, "data")).
			BinariesPath(cacheDirectory).
			Locale(DefaultLocale)
	} else {
		config = *dbConfig
		config = config.Port(uint32(freePort))
		config = config.RuntimePath(cacheDirectory)
		config = config.DataPath(filepath.Join(cacheDirectory, "data"))
		config = config.BinariesPath(cacheDirectory)
	}

	pg := embeddedpostgres.NewDatabase(config)
	if err := pg.Start(); err != nil {
		return nil, 0, fmt.Errorf("failed to start embedded postgres: %w", err)
	}

	return pg, uint32(freePort), nil
}

// StopPostgres 임베디드 PostgreSQL 서버를 중지합니다.
// 이 함수는 스레드 안전하며, 여러 번 호출되어도 안전합니다.
func StopPostgres() error {
	serverMutex.Lock()
	defer serverMutex.Unlock()

	logger := getLogger()
	logger.Info("Stopping PostgreSQL server")

	if !serverStarted || serverStopped {
		logger.Debug("PostgreSQL server is not running or already stopped")
		return nil
	}

	var errs []error

	// 기본 데이터베이스 연결 종료
	if baseDBClient != nil {
		logger.Debug("Closing base database connection")
		if err := baseDBClient.Close(); err != nil {
			err := fmt.Errorf("failed to close base database connection: %w", err)
			logger.Error("Failed to close base database connection", zap.Error(err))
			errs = append(errs, err)
		} else {
			logger.Debug("Successfully closed base database connection")
		}
	}

	// 서버 중지
	if server != nil {
		logger.Debug("Stopping PostgreSQL server process")
		if err := server.Stop(); err != nil {
			err := fmt.Errorf("failed to stop embedded postgres: %w", err)
			logger.Error("Failed to stop PostgreSQL server", zap.Error(err))
			errs = append(errs, err)
		} else {
			serverStopped = true
			logger.Info("Successfully stopped PostgreSQL server")
		}
	}

	serverStopped = true
	serverStarted = false

	// 캐시 디렉토리 정리
	if cacheDirectory != "" {
		logger.Debug("Removing cache directory", zap.String("path", cacheDirectory))
		if err := os.RemoveAll(cacheDirectory); err != nil {
			err := fmt.Errorf("failed to remove cache directory %s: %w", cacheDirectory, err)
			logger.Error("Failed to remove cache directory",
				zap.String("path", cacheDirectory),
				zap.Error(err))
			errs = append(errs, err)
		} else {
			logger.Info("Successfully removed cache directory",
				zap.String("path", cacheDirectory))
		}
	}

	if len(errs) > 0 {
		err := fmt.Errorf("%d error(s) occurred while stopping PostgreSQL: %+v", len(errs), errs)
		logger.Error("Errors occurred while stopping PostgreSQL",
			zap.Errors("errors", errs))
		return err
	}

	logger.Info("Successfully stopped PostgreSQL server and cleaned up all resources")
	return nil
}

// TestMainWrapper 테스트 메인 함수를 래핑하여 테스트 환경을 설정합니다.
// 이 함수는 테스트 패키지의 TestMain 함수에서 호출되어야 합니다.
//
// 사용 예시:
//
//	func TestMain(m *testing.M) {
//		os.Exit(testing.TestMainWrapper(m, nil))
//	}
func TestMainWrapper(m *testing.M, dbConfig *embeddedpostgres.Config) int {
	logger := getLogger()
	logger.Info("Starting test execution")

	// 서버 시작
	logger.Info("Starting embedded PostgreSQL server for tests")
	if err := StartEmbeddedPostgres(dbConfig); err != nil {
		logError("Failed to start embedded PostgreSQL server", err)
		return 1
	}

	// 테스트 실행
	logger.Info("Running tests...")
	code := m.Run()

	// 모든 테스트가 완료되면 서버 중지
	logger.Info("Tests completed, stopping PostgreSQL server")
	if err := StopPostgres(); err != nil {
		logError("Failed to stop embedded PostgreSQL server", err)
		return 1
	}

	logger.Info("Test execution completed", zap.Int("exit_code", code))
	return code
}

// createDatabase 새 데이터베이스를 생성합니다.
func createDatabase(dbName string) error {
	logger := getLogger().With(zap.String("database", dbName))
	logger.Debug("Creating database")

	if baseDBClient == nil {
		err := fmt.Errorf("base database client is not initialized")
		logError("Cannot create database: base connection not available", err)
		return err
	}

	// 이미 존재하는 데이터베이스인지 확인
	var exists bool
	err := baseDBClient.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		err = fmt.Errorf("failed to check if database exists: %w", err)
		logError("Failed to check database existence", err)
		return err
	}

	if exists {
		logger.Debug("Database already exists, skipping creation")
		return nil // 이미 존재하면 생성하지 않음
	}

	// 데이터베이스 생성
	logger.Info("Creating new database")
	// 안전하게 식별자를 이스케이프
	escapedDBName := `"` + strings.ReplaceAll(dbName, `"`, `""`) + `"`
	createQuery := fmt.Sprintf("CREATE DATABASE %s", escapedDBName)
	_, err = baseDBClient.Exec(createQuery)
	if err != nil {
		return fmt.Errorf("failed to create database %s: %w", dbName, err)
	}

	return nil
}

// dropDatabase는 데이터베이스를 삭제합니다.
// 이 함수는 내부적으로 사용되며, 외부에서는 DBClient.Close()를 통해 호출되어야 합니다.
func dropDatabase(dbName string) error {
	if dbName == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	logger := getLogger().With(zap.String("database", dbName))
	logger.Info("Starting database drop process")

	if baseDBClient == nil {
		err := fmt.Errorf("base database client is not initialized")
		logger.Error("Cannot drop database: base client is nil", zap.Error(err))
		return err
	}

	// 데이터베이스 존재 여부 확인
	var exists bool
	err := baseDBClient.QueryRow(
		`SELECT 1 FROM pg_database WHERE datname = $1`, dbName).Scan(&exists)

	if err != nil && err != sql.ErrNoRows {
		err = fmt.Errorf("failed to check if database exists: %w", err)
		logger.Error("Database existence check failed", zap.Error(err))
		return err
	}

	if !exists {
		logger.Debug("Database does not exist, nothing to drop")
		return nil
	}

	// 다른 세션이 연결되어 있는 경우 강제 종료
	logger.Debug("Terminating active connections to database")
	_, err = baseDBClient.Exec(
		`SELECT pg_terminate_backend(pid) 
		 FROM pg_stat_activity 
		 WHERE datname = $1 
		 AND pid <> pg_backend_pid()`, dbName)

	if err != nil {
		logger.Warn("Failed to terminate some database connections",
			zap.Error(err),
			zap.String("database", dbName))
		// 계속 진행 (일부 연결이 종료되지 않아도 삭제 시도)
	}

	// 데이터베이스 삭제
	logger.Info("Dropping database")
	// 안전하게 식별자를 이스케이프
	escapedDBName := `"` + strings.ReplaceAll(dbName, `"`, `""`) + `"`
	dropQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s", escapedDBName)
	_, err = baseDBClient.Exec(dropQuery)

	if err != nil {
		err = fmt.Errorf("failed to drop database %s: %w", dbName, err)
		logger.Error("Database drop failed",
			zap.String("database", dbName),
			zap.Error(err))
		return err
	}

	logger.Info("Successfully dropped database")
	return nil
}

// generateTestDBName 테스트용 데이터베이스 이름을 생성합니다.
func generateTestDBName() string {
	return fmt.Sprintf("%s%d_%d", TestDBPrefix, os.Getpid(), time.Now().UnixNano()%10000)
}

// CreateTestDB 지정된 커넥터를 사용하여 테스트 데이터베이스를 생성합니다.
// 사용자는 반드시 DBConnector 인터페이스를 구현한 커넥터를 전달해야 합니다.
func CreateTestDB(connector DBConnector) (*DBClient, error) {
	logger := getLogger()
	logger.Info("Creating test database")

	if connector == nil {
		err := fmt.Errorf("DBConnector must not be nil")
		logError("Invalid argument", err)
		return nil, err
	}

	serverMutex.Lock()
	defer serverMutex.Unlock()

	if !serverStarted || serverStopped {
		err := fmt.Errorf("database server is not running")
		logError("Cannot create test database", err)
		return nil, err
	}

	dbName := generateTestDBName()
	logger = logger.With(zap.String("database", dbName))
	logger.Debug("Generated test database name")

	// 데이터베이스 생성
	logger.Info("Creating test database")
	if err := createDatabase(dbName); err != nil {
		err = fmt.Errorf("failed to create test database: %w", err)
		logError("Failed to create test database", err)
		return nil, err
	}

	// 데이터베이스 연결 문자열 생성
	connString := getConnectionString(dbName)
	logger.Debug("Connecting to test database")

	// 커넥터를 사용하여 데이터베이스에 연결
	client, err := connector.Connect(connString)
	if err != nil {
		// 생성된 데이터베이스 정리
		logger.Error("Failed to connect to test database, cleaning up", zap.Error(err))
		if dropErr := dropDatabase(dbName); dropErr != nil {
			logError("Failed to clean up test database after connection error", dropErr)
		}
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Reset 호출로 데이터베이스 초기화
	logger.Debug("Resetting test database")
	if err := connector.Reset(); err != nil {
		logger.Error("Failed to reset test database, cleaning up", zap.Error(err))
		if closeErr := connector.Close(); closeErr != nil {
			logError("Failed to close connector after reset error", closeErr)
		}
		if dropErr := dropDatabase(dbName); dropErr != nil {
			logError("Failed to clean up test database after reset error", dropErr)
		}
		return nil, fmt.Errorf("failed to reset test database: %w", err)
	}

	logger.Info("Successfully created and initialized test database")
	return &DBClient{
		Client:           client,
		DBName:           dbName,
		ConnectionString: connString,
		connector:        connector,
	}, nil
}

// TestHelper 테스트에 유용한 헬퍼 함수들을 제공합니다.
type TestHelper struct {
	dbClient *DBClient
}

// NewTestHelper 새로운 TestHelper 인스턴스를 생성합니다.
func NewTestHelper(dbClient *DBClient) *TestHelper {
	return &TestHelper{
		dbClient: dbClient,
	}
}

// ResetDB 데이터베이스를 초기 상태로 되돌립니다.
func (h *TestHelper) ResetDB() error {
	logger := getLogger().With(zap.String("database", h.dbClient.DBName))
	logger.Debug("Resetting database to initial state")

	if h.dbClient.connector == nil {
		err := fmt.Errorf("database connector is nil")
		logError("Cannot reset database", err)
		return err
	}

	if err := h.dbClient.connector.Reset(); err != nil {
		logError("Failed to reset database", err)
		return fmt.Errorf("failed to reset database: %w", err)
	}

	logger.Info("Successfully reset database")
	return nil
}

// MustResetDB 데이터베이스를 초기 상태로 되돌리고, 실패하면 테스트를 즉시 중단합니다.
func (h *TestHelper) MustResetDB(t testing.TB) {
	t.Helper()
	logger := getLogger().With(zap.String("database", h.dbClient.DBName))
	logger.Debug("Resetting database (must)")

	if err := h.ResetDB(); err != nil {
		logError("Fatal: failed to reset database", err)
		t.Fatalf("Failed to reset database: %v", err)
	}
}

// Close 테스트 헬퍼에서 사용한 리소스를 정리합니다.
func (h *TestHelper) Close() error {
	logger := getLogger()
	if h.dbClient != nil {
		logger = logger.With(zap.String("database", h.dbClient.DBName))
	}
	logger.Debug("Closing test helper")

	if h.dbClient.connector != nil {
		logger.Debug("Closing database connector")
		if err := h.dbClient.connector.Close(); err != nil {
			logError("Failed to close database connector", err)
			return fmt.Errorf("failed to close database connector: %w", err)
		}
		logger.Info("Successfully closed test helper")
	}
	return nil
}
