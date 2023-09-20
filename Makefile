all: lint tests binary ## Runs lint, tests, and builds the binary

PHONY: help all tests coverage lint golint clean binary go-run vendor unit-tests
GOOS=linux

APP_NAME=loadbalancerproviderhaproxy

help: Makefile ## Print help
	@grep -h "##" $(MAKEFILE_LIST) | grep -v grep | sed -e 's/:.*##/#/' | column -c 2 -t -s#

tests: | unit-tests

unit-tests: ## Runs unit tests
	@echo --- Running unit tests...
	@date -Iseconds
	@go test -race -cover -failfast -tags testtools -p 1 -v ./...

coverage: ## Generates coverage report
	@echo --- Generating coverage report...
	@date -Iseconds
	@go test -race -coverprofile=coverage.out -covermode=atomic -tags testtools -p 1 ./...
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out

lint: golint ## Runs linting

golint:
	@echo --- Running golint...
	@date -Iseconds
	@golangci-lint run

clean: ## Clean up all the things
	@echo --- Cleaning...
	@date -Iseconds
	@rm -rf ./bin/
	@rm -rf coverage.out
	@go clean -testcache

binary: | vendor ## Builds the binary
	@echo --- Building binary...
	@date -Iseconds
	@go build -o bin/${APP_NAME} main.go

vendor: ## Vendors dependencies
	@echo --- Downloading dependencies...
	@date -Iseconds
	@go mod tidy
	@go mod download

go-run: ## Runs the app
	@echo --- Running binary...
	@date -Iseconds
	@go run main.go process --debug
