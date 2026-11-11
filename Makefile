.PHONY: all test tidy build run format clean coverage coverage-rich lint fix

.PHONY: all
all: tidy test build

.PHONY: test
test:
	go test -failfast ./...

.PHONY: coverage
coverage:
	if not exist coverage mkdir coverage
	go test -coverprofile=coverage/coverage.out -covermode=atomic ./...
	go tool cover -html=coverage/coverage.out -o coverage/report.html
	@echo Отчет готов: coverage/report.html

.PHONY: coverage-rich
coverage-rich:
	if not exist coverage mkdir coverage
	go test -coverprofile=coverage/coverage.out -covermode=atomic ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	go tool gocover-cobertura -by-files < coverage/coverage.out > coverage/cobertura.xml
	reportgenerator -reports:coverage/cobertura.xml -targetdir:coverage/report/ -reporttypes:Html

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: build
build:
	go build -o bin/compiler.exe cmd/main.go

.PHONY: clean
clean:
	go clean -testcache
	if exist bin\compiler.exe del /Q bin\compiler.exe
	if exist tests-coverage.out del /Q tests-coverage.out
	if exist tests-coverage.html del /Q tests-coverage.html

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
	gofmt -w .