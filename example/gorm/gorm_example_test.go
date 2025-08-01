package gorm_test

import (
	"bytes"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"testing"
	"time"

	"github.com/tidylogic/pgtestkit"
	gormConnector "github.com/tidylogic/pgtestkit/example/gorm"
	"gorm.io/gorm"
)

// User 테스트용 모델
type User struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	Email     string `gorm:"uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func TestGORMExample(t *testing.T) {
	t.Run("GORM을 사용한 기본 CRUD 테스트", func(t *testing.T) {
		// 1. 테스트 DB 클라이언트 생성
		dbClient, err := pgtestkit.CreateTestDB(&gormConnector.GORMConnector{})
		if err != nil {
			t.Fatalf("테스트 DB 생성 실패: %v", err)
		}
		defer dbClient.Close()

		// 테스트 헬퍼 생성
		helper := pgtestkit.NewTestHelper(dbClient)
		defer helper.Close()

		// GORM 클라이언트 가져오기
		gormDB, ok := dbClient.Client.(*gorm.DB)
		if !ok {
			t.Fatal("GORM DB 클라이언트를 가져오지 못했습니다")
		}

		// 2. 스키마 마이그레이션
		err = gormDB.AutoMigrate(&User{})
		if err != nil {
			t.Fatalf("마이그레이션 실패: %v", err)
		}

		t.Run("사용자 생성 및 조회", func(t *testing.T) {
			// 테이블 초기화
			helper.MustResetDB(t)

			// 3. 테스트 데이터 생성
			user := &User{
				Name:  "테스트 사용자",
				Email: "test@example.com",
			}
			result := gormDB.Create(user)
			if result.Error != nil {
				t.Fatalf("사용자 생성 실패: %v", result.Error)
			}

			// 4. 데이터 조회
			var foundUser User
			if err := gormDB.First(&foundUser, "email = ?", "test@example.com").Error; err != nil {
				t.Fatalf("사용자 조회 실패: %v", err)
			}

			if foundUser.Name != "테스트 사용자" {
				t.Errorf("예상한 사용자 이름: '테스트 사용자', 실제: '%s'", foundUser.Name)
			}
		})

		t.Run("테스트 격리 확인", func(t *testing.T) {
			// 테이블 초기화
			helper.MustResetDB(t)

			// 다른 테스트에서도 데이터베이스가 깨끗한지 확인
			var count int64
			if err := gormDB.Model(&User{}).Count(&count).Error; err != nil {
				t.Fatalf("사용자 수 조회 실패: %v", err)
			}

			// 이전 테스트의 데이터가 남아있지 않아야 함
			if count != 0 {
				t.Errorf("예상된 사용자 수: 0, 실제: %d", count)
			}
		})
	})
}

func TestMain(m *testing.M) {
	buf := &bytes.Buffer{}

	config := embeddedpostgres.DefaultConfig().
		Username(pgtestkit.DefaultUser).
		Password(pgtestkit.DefaultPassword).
		Database(pgtestkit.DefaultDB).
		Version(embeddedpostgres.V13).
		StartTimeout(10 * time.Second).
		Logger(buf).
		Locale(pgtestkit.DefaultLocale)

	// 모든 테스트에 대해 DB 서버 자동 관리
	pgtestkit.TestMainWrapper(m, &config)
}
