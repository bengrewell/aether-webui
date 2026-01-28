# Build variables
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Binary name and paths
BINARY_NAME := aether-webd
CMD_PATH := ./cmd/aether-webd
BIN_DIR := bin

# Frontend paths
FRONTEND_DIR := web/frontend
FRONTEND_DIST := $(FRONTEND_DIR)/dist
EMBED_DIR := internal/frontend/dist

# Docker settings
DOCKER_IMAGE := ghcr.io/bengrewell/aether-webd
DOCKER_TAG := $(VERSION)

# Linker flags for version injection
LDFLAGS := -X 'main.version=$(VERSION)' \
           -X 'main.commitHash=$(COMMIT)' \
           -X 'main.branch=$(BRANCH)' \
           -X 'main.buildDate=$(DATE)'

.PHONY: build clean test coverage coverage-html run version docker-build docker-push frontend embed-frontend all ensure-frontend init-submodules

# Initialize git submodules if needed
init-submodules:
	@if [ ! -f $(FRONTEND_DIR)/package.json ]; then \
		echo "Initializing git submodules..."; \
		git submodule update --init --recursive; \
	fi

# Build frontend (requires npm)
frontend: init-submodules
	cd $(FRONTEND_DIR) && npm install && npm run build

# Copy frontend dist to embed location
embed-frontend: frontend
	rm -rf $(EMBED_DIR)
	cp -r $(FRONTEND_DIST) $(EMBED_DIR)

# Ensure frontend is embedded (build only if missing)
ensure-frontend:
	@if [ ! -f $(EMBED_DIR)/index.html ]; then \
		echo "Frontend not found in $(EMBED_DIR), building..."; \
		$(MAKE) embed-frontend; \
	fi

# Build the binary with version info (ensures frontend is embedded)
build: ensure-frontend
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Built $(BIN_DIR)/$(BINARY_NAME) $(VERSION)"

# Build both frontend and backend with embedding (forces frontend rebuild)
all: embed-frontend build
	@echo "Built $(BIN_DIR)/$(BINARY_NAME) $(VERSION) with embedded frontend"

# Remove build artifacts
clean:
	rm -rf $(BIN_DIR) coverage.out coverage.html $(FRONTEND_DIST) $(EMBED_DIR)

# Run tests with coverage
test:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# View coverage report in terminal
coverage: test
	go tool cover -func=coverage.out

# Generate HTML coverage report
coverage-html: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build and run
run: build
	$(BIN_DIR)/$(BINARY_NAME)

# Display version info that would be injected
version:
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Branch:     $(BRANCH)"
	@echo "Build Date: $(DATE)"

# Build Docker image
docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BRANCH=$(BRANCH) \
		--build-arg BUILD_DATE=$(DATE) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		-f deploy/docker/Dockerfile .
	@echo "Built $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Push Docker image
docker-push: docker-build
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest
