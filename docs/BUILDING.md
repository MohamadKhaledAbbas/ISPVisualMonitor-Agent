# Building from Source

Instructions for building ISP Visual Monitor Agent from source code.

## 📋 Prerequisites

### Required

- **Go 1.21 or later**
  ```bash
  go version  # Should show 1.21+
  ```

- **Protocol Buffers Compiler**
  ```bash
  # Ubuntu/Debian
  sudo apt-get install protobuf-compiler

  # macOS
  brew install protobuf

  # Windows
  # Download from https://github.com/protocolbuffers/protobuf/releases
  ```

- **Go Protocol Buffer Plugins**
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```

### Optional

- **Make** (for using Makefile)
- **golangci-lint** (for linting)
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```

## 🔨 Building

### Quick Build

```bash
# Clone repository
git clone https://github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent.git
cd ISPVisualMonitor-Agent

# Generate protobuf code
make proto

# Build for current platform
make build

# Binary will be in ./bin/ispagent
./bin/ispagent --version
```

### Cross-Compilation

Build for all supported platforms:

```bash
make build-all
```

Binaries will be created in `./bin/`:
- `ispagent-linux-amd64` - Linux x86_64
- `ispagent-linux-arm64` - Linux ARM64 (Raspberry Pi 4, etc.)
- `ispagent-linux-armv7` - Linux ARMv7 (Raspberry Pi 3, etc.)
- `ispagent-windows-amd64.exe` - Windows x86_64
- `ispagent-darwin-amd64` - macOS Intel
- `ispagent-darwin-arm64` - macOS Apple Silicon

### Custom Build

```bash
# Set version
export VERSION=1.0.0

# Build with custom LDFLAGS
go build -ldflags "-X github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/version.Version=$VERSION" \
  -o bin/ispagent \
  ./cmd/ispagent
```

### Static Binary (for containers)

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags "-X github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/version.Version=1.0.0 -w -s" \
  -a -installsuffix cgo \
  -o bin/ispagent-static \
  ./cmd/ispagent
```

## 🧪 Testing

### Run All Tests

```bash
make test
```

Or manually:

```bash
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

### Run Specific Tests

```bash
# Test config package
go test -v ./internal/config

# Test collector
go test -v ./internal/collector/...

# Test with verbose output
go test -v -run TestConfigLoad ./internal/config
```

### Benchmarks

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Benchmark specific package
go test -bench=. ./internal/collector/mikrotik
```

## 🔍 Code Quality

### Linting

```bash
make lint
```

Or manually:

```bash
golangci-lint run ./...

# Auto-fix issues
golangci-lint run --fix ./...
```

### Format Code

```bash
# Format all Go files
go fmt ./...

# Or use gofmt with custom options
gofmt -w -s .
```

### Vet Code

```bash
go vet ./...
```

## 📦 Dependencies

### Update Dependencies

```bash
# Update to latest compatible versions
go get -u ./...

# Tidy up go.mod and go.sum
go mod tidy

# Verify dependencies
go mod verify
```

### Vendor Dependencies

```bash
# Create vendor directory
go mod vendor

# Build using vendor
go build -mod=vendor -o bin/ispagent ./cmd/ispagent
```

## 🐳 Docker Build

### Build Docker Image

```bash
docker build -t ispagent:latest -f deploy/docker/Dockerfile .
```

### Multi-Architecture Build

```bash
docker buildx create --use
docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 \
  -t ispagent:latest \
  -f deploy/docker/Dockerfile \
  --push .
```

## 📦 Creating Releases

### Tag Release

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Build Release Binaries

```bash
# Build all platforms
make build-all

# Create checksums
cd bin
sha256sum * > checksums.txt

# Create tar archives
tar czf ispagent-linux-amd64-v1.0.0.tar.gz ispagent-linux-amd64
tar czf ispagent-linux-arm64-v1.0.0.tar.gz ispagent-linux-arm64
# ... etc
```

## 🛠️ Development Setup

### 1. Clone Repository

```bash
git clone https://github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent.git
cd ISPVisualMonitor-Agent
```

### 2. Install Tools

```bash
# Install development tools
make dev-tools

# Or manually
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 3. Generate Proto Files

```bash
make proto
```

### 4. Build

```bash
make build
```

### 5. Run Tests

```bash
make test
```

## 🔧 Makefile Targets

- `make proto` - Generate protobuf Go code
- `make build` - Build for current platform
- `make build-all` - Build for all platforms
- `make test` - Run tests with coverage
- `make lint` - Run linter
- `make clean` - Remove built binaries and coverage files
- `make install` - Install to /usr/local/bin (requires sudo)

## 🐛 Debugging

### Build with Debug Symbols

```bash
go build -gcflags="all=-N -l" -o bin/ispagent-debug ./cmd/ispagent
```

### Use Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug
dlv debug ./cmd/ispagent -- --config configs/agent.yaml.dev
```

## 📊 Performance Profiling

### CPU Profiling

```bash
go test -cpuprofile=cpu.prof -bench=. ./internal/collector/mikrotik
go tool pprof cpu.prof
```

### Memory Profiling

```bash
go test -memprofile=mem.prof -bench=. ./internal/collector/mikrotik
go tool pprof mem.prof
```

## 🚀 CI/CD Integration

The project includes GitHub Actions workflows:

- `.github/workflows/ci.yml` - Run tests and build on every push
- `.github/workflows/release.yml` - Build and publish releases on tags

### Local CI Simulation

```bash
# Install act (https://github.com/nektos/act)
brew install act  # macOS
# or
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Run CI locally
act push
```

## ❓ Troubleshooting

### "protoc: command not found"

Install Protocol Buffers compiler (see Prerequisites above).

### "protoc-gen-go: program not found"

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Add to `.bashrc` or `.zshrc` for persistence.

### Build Fails with "undefined: X"

Regenerate proto files:
```bash
make proto
go mod tidy
```

### Tests Fail

```bash
# Clean and rebuild
make clean
make proto
go mod tidy
make test
```

## 📚 Additional Resources

- [Go Documentation](https://go.dev/doc/)
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/)
- [Protocol Buffers Guide](https://protobuf.dev/programming-guides/proto3/)

---

**Questions?** Open an issue or check [README.md](../README.md).
