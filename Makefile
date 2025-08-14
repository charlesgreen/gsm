# Makefile for Google Secret Manager Emulator

# Variables
BINARY_NAME=gsm-server
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=cmd/server/main.go
DOCKER_IMAGE=gsm-emulator
DOCKER_REGISTRY=charlesgreen
GO_FILES=$(shell find . -name '*.go' -type f -not -path './vendor/*' -not -path './.git/*')

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOFMT=gofmt
GOVET=$(GOCMD) vet
GOLINT=golangci-lint

# Build flags
LDFLAGS=-ldflags "-s -w"

# Default target
.PHONY: all
all: clean build test

# Build binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_PATH)

# Run without building
.PHONY: dev
dev:
	@echo "Running in development mode..."
	$(GOCMD) run $(MAIN_PATH)

# Test commands
.PHONY: test
test:
	@echo "Running all tests..."
	$(GOTEST) -v ./...

.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v ./tests/unit/...

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v ./tests/integration/...

.PHONY: test-parity
test-parity:
	@echo "Running production parity tests..."
	$(GOTEST) -v ./tests/integration/production_parity_test.go

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -cover -coverprofile=coverage.out ./...
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-race
test-race:
	@echo "Running tests with race detector..."
	$(GOTEST) -race -v ./...

# Validate production parity with script
.PHONY: validate-parity
validate-parity:
	@echo "Validating production parity..."
	@chmod +x scripts/validate_parity.sh
	./scripts/validate_parity.sh

# Linting and formatting
.PHONY: lint
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		$(GOLINT) run; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		$(GOLINT) run; \
	fi

.PHONY: lint-fix
lint-fix:
	@echo "Running linters with auto-fix..."
	@if command -v golangci-lint > /dev/null; then \
		$(GOLINT) run --fix; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		$(GOLINT) run --fix; \
	fi

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w $(GO_FILES)
	@echo "Code formatted"

.PHONY: fmt-check
fmt-check:
	@echo "Checking code formatting..."
	@if [ -n "$$($(GOFMT) -l $(GO_FILES))" ]; then \
		echo "The following files need formatting:"; \
		$(GOFMT) -l $(GO_FILES); \
		exit 1; \
	else \
		echo "All files are properly formatted"; \
	fi

.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Dependencies downloaded"

.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	$(GOMOD) tidy
	@echo "Dependencies updated"

.PHONY: deps-verify
deps-verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify
	@echo "Dependencies verified"

# Docker commands
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

.PHONY: docker-build-push
docker-build-push: docker-build
	@echo "Tagging and pushing Docker image..."
	docker tag $(DOCKER_IMAGE) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest
	@echo "Docker image pushed: $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest"

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8085:8085 $(DOCKER_IMAGE)

.PHONY: docker-compose-up
docker-compose-up:
	@echo "Starting services with docker-compose..."
	docker-compose up -d
	@echo "Services started. Emulator available at http://localhost:8085"

.PHONY: docker-compose-down
docker-compose-down:
	@echo "Stopping services..."
	docker-compose down
	@echo "Services stopped"

.PHONY: docker-compose-logs
docker-compose-logs:
	@echo "Showing docker-compose logs..."
	docker-compose logs -f

# Clean commands
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

.PHONY: clean-docker
clean-docker:
	@echo "Cleaning Docker resources..."
	@docker rmi $(DOCKER_IMAGE) 2>/dev/null || true
	@docker-compose down --volumes --remove-orphans 2>/dev/null || true
	@echo "Docker resources cleaned"

# Installation
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BINARY_PATH) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) installed to $(GOPATH)/bin/"

.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(GOPATH)/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) uninstalled"

# CI/CD helpers
.PHONY: ci
ci: deps fmt-check vet lint test

.PHONY: pre-commit
pre-commit: fmt lint test

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all              - Clean, build, and test"
	@echo "  build            - Build the binary"
	@echo "  run              - Build and run the application"
	@echo "  dev              - Run in development mode (without building)"
	@echo ""
	@echo "Testing:"
	@echo "  test             - Run all tests"
	@echo "  test-unit        - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-parity      - Run production parity tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  test-race        - Run tests with race detector"
	@echo "  validate-parity  - Run parity validation script"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint             - Run golangci-lint"
	@echo "  lint-fix         - Run golangci-lint with auto-fix"
	@echo "  fmt              - Format code"
	@echo "  fmt-check        - Check code formatting"
	@echo "  vet              - Run go vet"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps             - Download dependencies"
	@echo "  deps-update      - Update and tidy dependencies"
	@echo "  deps-verify      - Verify dependencies"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-build-push- Build and push Docker image"
	@echo "  docker-run       - Run Docker container"
	@echo "  docker-compose-up   - Start with docker-compose"
	@echo "  docker-compose-down - Stop docker-compose services"
	@echo "  docker-compose-logs - Show docker-compose logs"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean            - Clean build artifacts"
	@echo "  clean-docker     - Clean Docker resources"
	@echo "  install          - Install binary to GOPATH/bin"
	@echo "  uninstall        - Remove binary from GOPATH/bin"
	@echo ""
	@echo "CI/CD:"
	@echo "  ci               - Run CI pipeline (deps, format check, vet, lint, test)"
	@echo "  pre-commit       - Run pre-commit checks (format, lint, test)"

.DEFAULT_GOAL := help