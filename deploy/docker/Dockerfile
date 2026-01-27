# Build stage
FROM golang:1.22-alpine AS builder

# Install git for version info
RUN apk add --no-cache git

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with version info
ARG VERSION=dev
ARG COMMIT=unknown
ARG BRANCH=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X 'main.version=${VERSION}' -X 'main.commitHash=${COMMIT}' -X 'main.branch=${BRANCH}' -X 'main.buildDate=${BUILD_DATE}'" \
    -o aether-webd \
    ./cmd/aether-webd

# Runtime stage
FROM gcr.io/distroless/static:nonroot

# Copy binary from builder
COPY --from=builder /build/aether-webd /usr/local/bin/aether-webd

# Run as non-root user (provided by distroless/nonroot)
USER nonroot:nonroot

# Expose API port
EXPOSE 8680

# Default command - listen on all interfaces for container networking
ENTRYPOINT ["/usr/local/bin/aether-webd"]
CMD ["--listen", "0.0.0.0:8680"]
