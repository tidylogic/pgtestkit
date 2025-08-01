package gorm_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/tidylogic/pgtestkit"
	gormConnector "github.com/tidylogic/pgtestkit/example/gorm"
	"gorm.io/gorm"
)

// TestParallelSafety parallel 테스트 안전성을 확인합니다.
func TestParallelSafety(t *testing.T) {
	// 여러 병렬 테스트가 동시에 실행되어도 안전한지 확인
	for i := 0; i < 5; i++ {
		t.Run(fmt.Sprintf("ParallelTest_%d", i), func(t *testing.T) {
			t.Parallel()

			// 테스트 DB 클라이언트 생성
			dbClient, err := pgtestkit.CreateTestDB(&gormConnector.GORMConnector{})
			if err != nil {
				t.Fatalf("테스트 DB 생성 실패: %v", err)
			}
			defer func() {
				if err := dbClient.Close(); err != nil {
					t.Logf("Warning: error closing database client: %v", err)
				}
			}()

			// GORM 클라이언트 가져오기
			gormDB, ok := dbClient.Client.(*gorm.DB)
			if !ok {
				t.Fatal("GORM DB 클라이언트를 가져오지 못했습니다")
			}

			// 스키마 마이그레이션 - 여기서 bigserial 관련 에러가 발생할 수 있음
			err = gormDB.AutoMigrate(&User{})
			if err != nil {
				t.Fatalf("마이그레이션 실패: %v", err)
			}

			// 기본 CRUD 작업
			user := &User{
				Name:  fmt.Sprintf("TestUser_%d", i),
				Email: fmt.Sprintf("test_%d@example.com", i),
			}

			// 생성
			if err := gormDB.Create(user).Error; err != nil {
				t.Fatalf("사용자 생성 실패: %v", err)
			}

			// 조회
			var foundUser User
			if err := gormDB.First(&foundUser, user.ID).Error; err != nil {
				t.Fatalf("사용자 조회 실패: %v", err)
			}

			if foundUser.Name != user.Name {
				t.Errorf("Expected name %s, got %s", user.Name, foundUser.Name)
			}

			t.Logf("Successfully completed parallel test %d", i)
		})
	}
}

// TestConcurrentDatabaseCreation 동시 데이터베이스 생성 테스트
func TestConcurrentDatabaseCreation(t *testing.T) {
	// 동시에 여러 데이터베이스를 생성해도 안전한지 확인
	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("ConcurrentCreate_%d", i), func(t *testing.T) {
			t.Parallel()

			// 약간의 지연을 두어 동시성 테스트
			time.Sleep(time.Duration(i*10) * time.Millisecond)

			dbClient, err := pgtestkit.CreateTestDB(&gormConnector.GORMConnector{})
			if err != nil {
				t.Fatalf("동시 DB 생성 실패: %v", err)
			}
			defer dbClient.Close()

			gormDB, ok := dbClient.Client.(*gorm.DB)
			if !ok {
				t.Fatal("GORM DB 클라이언트를 가져오지 못했습니다")
			}

			// 복잡한 스키마로 테스트
			type ComplexModel struct {
				ID     uint   `gorm:"primaryKey"`
				Name   string `gorm:"size:255;not null"`
				Value  int64  `gorm:"type:bigint"`
				Data   []byte `gorm:"type:bytea"`
				Active bool   `gorm:"default:true"`
			}

			if err := gormDB.AutoMigrate(&ComplexModel{}); err != nil {
				t.Fatalf("복잡한 스키마 마이그레이션 실패: %v", err)
			}

			// 데이터 삽입 테스트
			model := &ComplexModel{
				Name:   fmt.Sprintf("TestModel_%d", i),
				Value:  int64(i * 1000),
				Data:   []byte(fmt.Sprintf("test_data_%d", i)),
				Active: true,
			}

			if err := gormDB.Create(model).Error; err != nil {
				t.Fatalf("복잡한 모델 생성 실패: %v", err)
			}

			t.Logf("Successfully completed concurrent creation test %d", i)
		})
	}
}

// TestLongRunningParallelOperations 장시간 실행되는 병렬 작업 테스트
func TestLongRunningParallelOperations(t *testing.T) {
	for i := 0; i < 2; i++ {
		t.Run(fmt.Sprintf("LongRunning_%d", i), func(t *testing.T) {
			t.Parallel()

			dbClient, err := pgtestkit.CreateTestDB(&gormConnector.GORMConnector{})
			if err != nil {
				t.Fatalf("테스트 DB 생성 실패: %v", err)
			}
			defer dbClient.Close()

			gormDB, ok := dbClient.Client.(*gorm.DB)
			if !ok {
				t.Fatal("GORM DB 클라이언트를 가져오지 못했습니다")
			}

			if err := gormDB.AutoMigrate(&User{}); err != nil {
				t.Fatalf("마이그레이션 실패: %v", err)
			}

			// 대량의 데이터 작업
			for j := 0; j < 10; j++ {
				user := &User{
					Name:  fmt.Sprintf("BulkUser_%d_%d", i, j),
					Email: fmt.Sprintf("bulk_%d_%d@example.com", i, j),
				}

				if err := gormDB.Create(user).Error; err != nil {
					t.Fatalf("대량 데이터 생성 실패 (i=%d, j=%d): %v", i, j, err)
				}

				// 작은 지연
				time.Sleep(10 * time.Millisecond)
			}

			// 데이터 검증
			var count int64
			if err := gormDB.Model(&User{}).Count(&count).Error; err != nil {
				t.Fatalf("사용자 수 카운트 실패: %v", err)
			}

			if count != 10 {
				t.Errorf("Expected 10 users, got %d", count)
			}

			t.Logf("Successfully completed long running test %d with %d records", i, count)
		})
	}
}
