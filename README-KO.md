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

Go 언어를 위한 경량화된 테스트용 PostgreSQL 데이터베이스 관리 도구입니다. 테스트 환경에서 쉽게 임베디드 PostgreSQL 서버를 실행하고 관리할 수 있도록 도와줍니다.

## ✨ Features

[View in English](README.md) | [한국어로 보기](README-KO.md)

## ✨ 특징

- **PostgreSQL 전용**: PostgreSQL에 최적화된 테스트 도구
- **Zero Configuration**: 기본 설정으로 즉시 사용 가능
- **빠른 테스트 실행**: 임베디드 서버 재사용으로 빠른 테스트 실행
- **완전한 격리**: 테스트 간 데이터베이스 격리 보장
- **유연한 아키텍처**: 인터페이스 기반 설계로 어떤 ORM이나 드라이버와도 통합 가능
- **유연한 데이터베이스 연결**: DBConnector 인터페이스를 통해 GORM, database/sql 등 다양한 데이터베이스 클라이언트와 호환

## 코드 품질

이 프로젝트는 [golangci-lint](https://golangci-lint.run/)를 사용하여 코드 품질을 유지합니다. 

### 린트 실행

```bash
# 린트 실행
golangci-lint run

# 자동 수정 가능한 문제 수정
golangci-lint run --fix
```

또는 Makefile을 사용할 수 있습니다:

```bash
# 코드 포맷팅 및 린트 실행
make all
```

### 포함된 린터들

- **gofmt/goimports**: 코드 포맷팅
- **govet**: Go 벤더 권장사항 검사
- **staticcheck**: 정적 분석
- **errcheck**: 에러 체크 누락 검사
- **gocritic**: 코드 품질 향상을 위한 제안
- 기타 유용한 린터들

## 주요 기능

- 🚀 테스트 실행 시 자동으로 임베디드 PostgreSQL 서버 시작/중지
- 🧪 테스트 간 자동 데이터베이스 초기화로 테스트 격리 보장
- ⚡️ 고성능 - 서버는 한 번만 시작되고 모든 테스트에서 재사용됨
- 🔄 자동 리소스 정리
- 🛠️ GORM 및 표준 database/sql 인터페이스 모두 지원
- 🔌 커스터마이징 가능한 데이터베이스 설정

## 설치

```bash
go get github.com/tidylogic/pgtestkit
```

### DBConnector 구현하기

`DBConnector` 인터페이스를 구현하여 다양한 데이터베이스 클라이언트를 사용할 수 있습니다. 아래는 인터페이스 정의입니다:

```go
type DBConnector interface {
    // Connect 데이터베이스에 연결하고, ORM 클라이언트와 SQL DB를 반환합니다.
    Connect(connString string) (interface{}, error)
    // Close 데이터베이스 연결을 종료합니다.
    Close() error
    // Reset 데이터베이스를 초기 상태로 되돌립니다.
    Reset() error
}
```

예제 구현은 [example](example) 디렉토리를 참조하세요.

## 빠른 시작

### 기본 사용법

```go
package yourpackage_test

import (
    "database/sql"
    "testing"

    "github.com/tidylogic/pgtestkit"
    "github.com/tidylogic/pgtestkit/example/sql"
)

func TestWithSQL(t *testing.T) {
    // SQL 커넥터로 테스트 DB 생성
    dbClient, err := pgtestkit.CreateTestDB(&sql.SQLConnector{})
    if err != nil {
        t.Fatalf("Failed to create test DB: %v", err)
    }
    defer dbClient.Close()

    // *sql.DB 인스턴스 가져오기
    db := dbClient.Client.(*sql.DB)
    
    // 여기에 테스트 코드 작성...
}
```

### GORM 사용하기

```go
package yourpackage_test

import (
    "testing"

    "github.com/tidylogic/pgtestkit"
    "github.com/tidylogic/pgtestkit/example/gorm"
    "gorm.io/gorm"
)

func TestWithGORM(t *testing.T) {
    // GORM 커넥터로 테스트 DB 생성
    dbClient, err := pgtestkit.CreateTestDB(&gorm.GORMConnector{})
    if err != nil {
        t.Fatalf("Failed to create test DB: %v", err)
    }
    defer dbClient.Close()

    // *gorm.DB 인스턴스 가져오기
    gormDB := dbClient.Client.(*gorm.DB)
    
    // GORM을 사용한 테스트 코드...
}
```

### TestMain과 함께 사용하기

```go
package yourpackage_test

import (
    "os"
    "testing"
    "github.com/tidylogic/pgtestkit"
)

func TestMain(m *testing.M) {
    // 테스트 실행 전후로 DB 서버 자동 관리
    os.Exit(pgtestkit.TestMainWrapper(m))
}
```

### 테스트 헬퍼 사용하기

```go
func TestWithHelper(t *testing.T) {
    dbClient, err := pgtestkit.CreateTestDB(&gorm.GORMConnector{})
    if err != nil {
        t.Fatalf("Failed to create test DB: %v", err)
    }
    defer dbClient.Close()
    
    // 테스트 헬퍼 생성
    helper := pgtestkit.NewTestHelper(dbClient.GormDB)
    defer helper.Close()
    
    t.Run("Test case 1", func(t *testing.T) {
        // 테스트 케이스 시작 전 데이터베이스 초기화
        helper.MustResetDB(t)
        
        // 테스트 코드...
    })
    
    t.Run("Test case 2", func(t *testing.T) {
        // 다른 테스트 케이스 (이전 테스트와 완전히 분리됨)
        helper.MustResetDB(t)
        
        // 다른 테스트 코드...
    })
}
```

## 고급 사용법

### 커스텀 커넥터 구현

`DBConnector` 인터페이스를 구현하여 자신만의 데이터베이스 커넥터를 만들 수 있습니다:

```go
package yourpackage

type YourConnector struct {
    // 커넥터 상태 저장
}

func (c *YourConnector) Connect(connString string) (interface{}, error) {
    // 데이터베이스 연결 로직 구현
    // 반환 타입은 사용하는 ORM/드라이버에 따라 다릅니다
    return yourDBClient, nil
}

func (c *YourConnector) Close() error {
    // 리소스 정리 로직 구현
    return nil
}

func (c *YourConnector) Reset() error {
    // 데이터베이스 초기화 로직 구현
    return nil
}

// 사용 예시
func TestWithCustomConnector(t *testing.T) {
    dbClient, err := pgtestkit.CreateTestDB(&YourConnector{})
    if err != nil {
        t.Fatalf("Failed to create test DB: %v", err)
    }
    defer dbClient.Close()
    
    // 커스텀 클라이언트 사용
    client := dbClient.Client.(YourClientType)
    // ... 테스트 코드 ...
}
```

### 커스텀 데이터베이스 설정

```go
func TestWithCustomConfig(t *testing.T) {
    // 커스텀 설정으로 테스트 DB 생성
    dbClient, err := pgtestkit.CreateTestDB(
        pgtestkit.WithDatabaseName("custom_test_db"),
        pgtestkit.WithPort(5433), // 특정 포트 지정
        pgtestkit.WithGormConfig(&gorm.Config{
            Logger: logger.Default.LogMode(logger.Silent),
        }),
    )
    if err != nil {
        t.Fatalf("Failed to create test DB: %v", err)
    }
    defer dbClient.Close()
    
    // 테스트 코드...
}
```

### 데이터베이스 스키마 마이그레이션

```go
func TestWithMigrations(t *testing.T) {
    dbClient, err := pgtestkit.CreateTestDB()
    if err != nil {
        t.Fatalf("Failed to create test DB: %v", err)
    }
    defer dbClient.Close()
    
    // GORM 마이그레이션 실행
    err = dbClient.GormDB.AutoMigrate(&YourModel{})
    if err != nil {
        t.Fatalf("Failed to migrate: %v", err)
    }
    
    // 테스트 코드...
}
```

## 성능 고려사항

- **서버 재사용**: 테스트 실행 시 PostgreSQL 서버는 한 번만 시작되고 모든 테스트에서 재사용됩니다.
- **데이터베이스 격리**: 각 테스트는 고유한 데이터베이스를 받아 서로 간섭하지 않습니다.
- **병렬 테스트**: `t.Parallel()`과 함께 사용해도 안전합니다.

## 라이선스

MIT 라이선스 하에 배포됩니다. 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

## 기여

버그 리포트나 기능 제안은 이슈 트래커를 이용해 주세요. 풀 리퀘스트도 환영합니다!

## 감사의 말

이 프로젝트는 다음 오픈 소스 프로젝트를 기반으로 합니다:
- [embedded-postgres](https://github.com/fergusstrange/embedded-postgres)
