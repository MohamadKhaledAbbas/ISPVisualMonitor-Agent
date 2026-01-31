# Multi-stage build for minimal image size
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make protobuf protobuf-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Install protoc plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Copy source code
COPY . .

# Generate protobuf code
RUN make proto

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/version.Version=${VERSION:-dev}" \
    -a -installsuffix cgo \
    -o ispagent \
    ./cmd/ispagent

# Final stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates for TLS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -S ispagent && adduser -S ispagent -G ispagent

# Create directories
RUN mkdir -p /etc/ispagent /var/log/ispagent && \
    chown -R ispagent:ispagent /etc/ispagent /var/log/ispagent

# Copy binary from builder
COPY --from=builder /build/ispagent /usr/local/bin/ispagent
COPY --from=builder /build/configs/agent.yaml.example /etc/ispagent/agent.yaml.example

# Set user
USER ispagent

# Set working directory
WORKDIR /home/ispagent

# Expose gRPC port (if needed for reverse scenario)
# EXPOSE 50051

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD /usr/local/bin/ispagent --version || exit 1

# Run agent
ENTRYPOINT ["/usr/local/bin/ispagent"]
CMD ["--config", "/etc/ispagent/agent.yaml"]
