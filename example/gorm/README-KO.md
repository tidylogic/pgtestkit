# GORM 예제

[한국어](README-KO.md) | [English](README.md)

이 예제는 pgtestkit를 사용하여 GORM으로 PostgreSQL 데이터베이스 테스트를 수행하는 방법을 보여줍니다.

## 시작하기 전에

필요한 의존성을 설치하세요:

```bash
cd ..
go mod tidy
```

## 테스트 실행

```bash
# 모든 테스트 실행
cd ..
go test -v ./...

# 특정 테스트만 실행
cd ..
go test -v -run TestGORMExample
```

## 예제 설명

이 예제는 다음과 같은 기능을 보여줍니다:

1. 임베디드 PostgreSQL 서버 시작 및 중지
2. GORM을 사용한 데이터베이스 연결 관리
3. 테스트 간 데이터베이스 초기화
4. 기본적인 CRUD 작업 테스트
5. 테스트 격리 확인

## 사용자 정의 모델

`User` 모델을 예로 들어 테스트를 수행합니다. 필요에 따라 모델을 수정하거나 추가할 수 있습니다.

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

## 참고 사항

- 테스트 실행 전에 자동으로 임베디드 PostgreSQL 서버가 시작됩니다.
- 각 테스트는 독립적으로 실행되며, 테스트 간에 데이터베이스가 자동으로 초기화됩니다.
- 테스트가 완료되면 임베디드 서버가 자동으로 종료됩니다.
