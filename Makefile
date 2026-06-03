.PHONY: all build clean test lint run install help

APP_NAME   := wfh
BIN_DIR    := ./build
GO_FLAGS   := -ldflags="-s -w"
GO_PACKAGE := github.com/zinuo-xu/wfh

# Detect OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    BINARY := $(BIN_DIR)/$(APP_NAME)
endif
ifeq ($(UNAME_S),Darwin)
    BINARY := $(BIN_DIR)/$(APP_NAME)
endif
ifeq ($(OS),Windows_NT)
    BINARY := $(BIN_DIR)/$(APP_NAME).exe
endif

all: build

build: ## Build the binary
	@mkdir -p $(BIN_DIR)
	go build $(GO_FLAGS) -o $(BINARY) ./cmd/$(APP_NAME)
	@echo "Built: $(BINARY)"

clean: ## Remove build artifacts
	@rm -rf $(BIN_DIR)
	@echo "Cleaned build directory"

test: ## Run all tests
	go test ./... -v -count=1

test-short: ## Run short tests
	go test ./... -short -count=1

lint: ## Run linters
	go vet ./...
	@which golangci-lint 2>/dev/null && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

run: build ## Build and run wfh today report
	$(BINARY) today

install: build ## Install binary to GOPATH/bin
	@cp $(BINARY) $(GOPATH)/bin/$(APP_NAME)$(suffix)
	@echo "Installed to $(GOPATH)/bin/$(APP_NAME)"

dev: ## Run in development mode
	go run ./cmd/wfh/ status

watch: build ## Build and watch current directory
	$(BINARY) watch .

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
