# Build variables
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Binary name and paths
BINARY_NAME := aether-webd
CMD_PATH := ./cmd/aether-webd
BIN_DIR := bin

# Linker flags for version injection
LDFLAGS := -X 'main.version=$(VERSION)' \
           -X 'main.commitHash=$(COMMIT)' \
           -X 'main.branch=$(BRANCH)' \
           -X 'main.buildDate=$(DATE)'

.PHONY: build clean test run version

# Build the binary with version info
build:
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Built $(BIN_DIR)/$(BINARY_NAME) $(VERSION)"

# Remove build artifacts
clean:
	rm -rf $(BIN_DIR)

# Run tests
test:
	go test -v ./...

# Build and run
run: build
	$(BIN_DIR)/$(BINARY_NAME)

# Display version info that would be injected
version:
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Branch:     $(BRANCH)"
	@echo "Build Date: $(DATE)"
