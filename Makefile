.PHONY: all test tidy build run format clean

.PHONY: all
all: tidy test build

.PHONY: test
test:
	go test -failfast ./...

# Цель для запуска тестов с покрытием и генерации простого HTML-отчёта
.PHONY: coverage
coverage:
	mkdir -p coverage/
	go test -coverprofile=coverage/coverage.out -covermode=atomic ./...
	go tool cover -html=coverage/coverage.out -o coverage/report.html
	sed -i 's/black/whitesmoke/g' coverage/report.html

# Цель для запуска тестов с покрытием и генерации HTML-отчёта через ReportGenerator
.PHONY: coverage-rich
coverage-rich:
	mkdir -p coverage/
	go test -coverprofile=coverage/coverage.out -covermode=atomic ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	go tool gocover-cobertura -by-files < coverage/coverage.out > coverage/cobertura.xml
	reportgenerator -reports:coverage/cobertura.xml -targetdir:coverage/report/ -reporttypes:Html

.PHONY: tidy
tidy:
	go mod tidy && \
	go mod vendor

.PHONY: build
build:
	go build -o bin/compiler cmd/main.go

.PHONY: clean
clean:
	go clean -testcache
	rm -f bin/compiler
	rm -f tests-coverage.out tests-coverage.html

.PHONY: lint
lint:
	go vet ./...
	$(shell go env GOPATH)/bin/golangci-lint run

.PHONY: fix
fix:
	go fix ./...
	$(shell go env GOPATH)/bin/golangci-lint run --fix

.PHONY: format
format:
	$(shell go env GOPATH)/bin/golangci-lint fmt