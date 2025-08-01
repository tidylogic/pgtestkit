package gorm

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GORMConnector GORM을 사용하는 데이터베이스 커넥터입니다.
type GORMConnector struct {
	gormDB *gorm.DB
}

// Connect 데이터베이스에 연결하고 GORM DB 인스턴스를 반환합니다.
func (g *GORMConnector) Connect(connString string) (interface{}, error) {
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN: connString,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create GORM DB: %w", err)
	}

	g.gormDB = gormDB
	return gormDB, nil
}

// Close 데이터베이스 연결을 종료합니다.
func (g *GORMConnector) Close() error {
	if g.gormDB != nil {
		sqlDB, err := g.gormDB.DB()
		if err != nil {
			return fmt.Errorf("failed to get SQL DB from GORM: %w", err)
		}
		return sqlDB.Close()
	}
	return nil
}

// Reset 데이터베이스를 초기 상태로 되돌립니다.
func (g *GORMConnector) Reset() error {
	if g.gormDB == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 모든 테이블 가져오기
	tableNames := make([]string, 0)
	if err := g.gormDB.Raw(
		"SELECT tablename FROM pg_tables WHERE schemaname = 'public' ").
		Scan(&tableNames).Error; err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}

	// 외래 키 제약 조건 비활성화
	if err := g.gormDB.Exec("SET CONSTRAINTS ALL DEFERRED").Error; err != nil {
		return fmt.Errorf("failed to defer constraints: %w", err)
	}

	// 모든 테이블 TRUNCATE
	for _, tableName := range tableNames {
		if tableName == "schema_migrations" {
			continue // 마이그레이션 테이블은 건너뜀
		}

		// CASCADE 옵션으로 TRUNCATE 실행
		if err := g.gormDB.Exec(fmt.Sprintf("TRUNCATE TABLE \"%s\" CASCADE", tableName)).Error; err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", tableName, err)
		}

		// 시퀀스 재설정 (테이블명_id_seq 형식 가정)
		seqName := fmt.Sprintf("\"%s_id_seq\"", tableName)
		_ = g.gormDB.Exec(fmt.Sprintf(
			"SELECT setval('%s', COALESCE((SELECT MAX(id) FROM \"%s\"), 1), false)",
			seqName, tableName))
		// 시퀀스가 없으면 무시
	}

	// 외래 키 제약 조건 다시 활성화
	if err := g.gormDB.Exec("SET CONSTRAINTS ALL IMMEDIATE").Error; err != nil {
		return fmt.Errorf("failed to re-enable constraints: %w", err)
	}

	return nil
}
