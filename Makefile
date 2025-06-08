GOCMD=go
GOFMT=$(GOCMD) fmt
GOLINT=golangci-lint

# 기본 타겟
.PHONY: help
help:
	@echo "다음 명령어들을 사용할 수 있습니다:"
	@echo "  make lint       - 코드 품질 검사"
	@echo "  make test       - 테스트 실행"
	@echo "  make tidy       - go.mod 정리"
	@echo "  make all        - 모든 검사 및 테스트 실행"

# 코드 포맷팅
.PHONY: fmt
fmt:
	@$(GOFMT) ./...

# 코드 품질 검사
.PHONY: lint
lint:
	@$(GOLINT) run --timeout 2m

# 테스트 실행
.PHONY: test
test:
	@$(GOCMD) test -v -coverprofile=coverage.out -covermode=atomic ./...

# example 테스트 실행
.PHONY: test-example
test-example:
	@$(GOCMD) test -v ./example/...

# 커버리지 리포트 생성
.PHONY: coverage
coverage: test
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html

# go.mod 정리
.PHONY: tidy
tidy:
	@$(GOCMD) mod tidy
	@$(GOCMD) mod verify

# 모든 검사 및 테스트 실행
all: fmt lint test

# 도커를 사용한 빌드 (선택사항)
.PHONY: docker-build
docker-build:
	@docker build -t pgtestkit .

# 도커를 사용한 테스트 (선택사항)
.PHONY: docker-tests
docker-test:
	@docker run --rm -v ${PWD}:/app -w /app golangci/golangci-lint golangci-lint run -v
