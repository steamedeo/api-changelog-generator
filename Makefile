# Variables
BINARY_NAME=api-changelog.exe
VERSION=$(shell git describe --tags --always --dirty 2>nul || echo dev)
COMMIT=$(shell git rev-parse HEAD 2>nul || echo unknown)
DATE=$(shell powershell -Command "Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ'")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
.DEFAULT_GOAL := help

.PHONY: help build clean install push release

help: ## Show this help message
	@echo Available targets:
	@echo   build    - Build the binary
	@echo   clean    - Remove build artifacts
	@echo   install  - Install the binary to GOPATH/bin
	@echo   push     - Commit and push to main
	@echo   release  - Create and push a release tag

build: ## Build the binary
	@echo Building $(BINARY_NAME)...
	@go build $(LDFLAGS) -o bin/$(BINARY_NAME) .
	@echo Build complete: bin/$(BINARY_NAME)

clean: ## Remove build artifacts
	@echo Cleaning...
	@if exist bin rmdir /s /q bin
	@go clean
	@echo Clean complete

install: ## Install the binary to GOPATH/bin
	@echo Installing $(BINARY_NAME)...
	@go install $(LDFLAGS) .
	@echo Installed

push: ## Commit and push to main
	@powershell -ExecutionPolicy Bypass -File scripts/push.ps1

release: ## Create and push a release tag
	@powershell -ExecutionPolicy Bypass -File scripts/release.ps1
