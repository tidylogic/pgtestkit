package gorm_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidylogic/pgtestkit"
	gormConnector "github.com/tidylogic/pgtestkit/example/gorm"
	"gorm.io/gorm"
)

// ParallelUser 테스트용 병렬 사용자 모델
type ParallelUser struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	Email     string `gorm:"uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func TestGORMParallel(t *testing.T) {
	t.Run("병렬 테스트에서의 데이터베이스 격리 테스트", func(t *testing.T) {
		t.Parallel() // 병렬 테스트 활성화

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
		err = gormDB.AutoMigrate(&ParallelUser{})
		if err != nil {
			t.Fatalf("마이그레이션 실패: %v", err)
		}

		// 테스트 데이터 생성
		user := &ParallelUser{
			Name:  t.Name(),
			Email: t.Name() + "@example.com",
		}

		// 3. 데이터베이스에 사용자 생성
		result := gormDB.Create(user)
		assert.NoError(t, result.Error, "사용자 생성에 실패했습니다")

		// 4. 생성된 사용자 조회
		var foundUser ParallelUser
		err = gormDB.Where("email = ?", user.Email).First(&foundUser).Error
		assert.NoError(t, err, "사용자 조회에 실패했습니다")
		assert.Equal(t, user.Name, foundUser.Name, "사용자 이름이 일치하지 않습니다")

		t.Logf("테스트 %s에서 사용자 생성 및 조회 성공: %s", t.Name(), user.Email)
	})

	t.Run("병렬 테스트 간 데이터베이스 격리 확인", func(t *testing.T) {
		t.Parallel() // 병렬 테스트 활성화

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
		err = gormDB.AutoMigrate(&ParallelUser{})
		if err != nil {
			t.Fatalf("마이그레이션 실패: %v", err)
		}

		// 3. 현재 테스트의 데이터만 조회 (다른 병렬 테스트와 격리되어야 함)
		var count int64
		err = gormDB.Model(&ParallelUser{}).Count(&count).Error
		assert.NoError(t, err, "사용자 수 조회에 실패했습니다")
		assert.Equal(t, int64(0), count, "새로운 테스트는 비어있는 데이터베이스로 시작해야 합니다")

		t.Logf("테스트 %s에서 데이터베이스 격리 확인: 사용자 수 = %d", t.Name(), count)
	})
}

func TestConcurrentDatabaseOperations(t *testing.T) {
	// 테스트 DB 클라이언트 생성
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

	// 스키마 마이그레이션
	err = gormDB.AutoMigrate(&ParallelUser{})
	if err != nil {
		t.Fatalf("마이그레이션 실패: %v", err)
	}

	// 동시에 실행할 고루틴 수
	concurrency := 5
	var wg sync.WaitGroup
	results := make(chan error, concurrency)

	// 동시성 테스트 실행
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 각 고루틴은 별도의 트랜잭션을 사용
			err := gormDB.Transaction(func(tx *gorm.DB) error {
				// 고유한 이메일 생성
				email := t.Name() + "_" + string(rune('a'+id)) + "@example.com"

				// 사용자 생성
				user := &ParallelUser{
					Name:  t.Name() + "_user" + string(rune('A'+id)),
					Email: email,
				}

				if err := tx.Create(user).Error; err != nil {
					return err
				}

				// 생성된 사용자 조회
				var foundUser ParallelUser
				if err := tx.Where("email = ?", email).First(&foundUser).Error; err != nil {
					return err
				}

				if foundUser.Name != user.Name {
					t.Errorf("예상한 이름: %s, 실제: %s", user.Name, foundUser.Name)
				}

				return nil
			})

			results <- err
		}(i)
	}

	// 모든 고루틴이 완료될 때까지 대기
	go func() {
		wg.Wait()
		close(results)
	}()

	// 결과 확인
	for err := range results {
		assert.NoError(t, err, "동시성 테스트 중 오류 발생")
	}

	// 최종적으로 모든 사용자가 생성되었는지 확인
	var count int64
	if err := gormDB.Model(&ParallelUser{}).Count(&count).Error; err != nil {
		t.Fatalf("사용자 수 조회 실패: %v", err)
	}

	assert.Equal(t, int64(concurrency), count, "생성된 사용자 수가 일치하지 않습니다")
	t.Logf("동시성 테스트 완료: 총 %d명의 사용자가 생성됨", count)
}
