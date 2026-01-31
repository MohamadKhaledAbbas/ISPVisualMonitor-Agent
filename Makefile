.PHONY: proto build build-all test clean install lint

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/pkg/version.Version=$(VERSION)"

proto:
	@echo "Generating protobuf files..."
	@rm -rf api/proto/agentpb api/proto/licensepb
	@mkdir -p api/proto/agentpb api/proto/licensepb
	PATH=$$PATH:$$HOME/go/bin protoc -I=api/proto \
		--go_out=api/proto/agentpb --go_opt=paths=source_relative \
		--go-grpc_out=api/proto/agentpb --go-grpc_opt=paths=source_relative \
		api/proto/metrics.proto api/proto/sessions.proto api/proto/agent.proto
	PATH=$$PATH:$$HOME/go/bin protoc -I=api/proto \
		--go_out=api/proto/licensepb --go_opt=paths=source_relative \
		--go-grpc_out=api/proto/licensepb --go-grpc_opt=paths=source_relative \
		api/proto/license.proto

build:
	go build $(LDFLAGS) -o bin/ispagent ./cmd/ispagent

build-all:
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/ispagent-linux-amd64 ./cmd/ispagent
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/ispagent-linux-arm64 ./cmd/ispagent
	GOOS=linux GOARCH=arm GOARM=7 go build $(LDFLAGS) -o bin/ispagent-linux-armv7 ./cmd/ispagent
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/ispagent-windows-amd64.exe ./cmd/ispagent
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/ispagent-darwin-amd64 ./cmd/ispagent
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/ispagent-darwin-arm64 ./cmd/ispagent

test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ coverage.out

install: build
	sudo cp bin/ispagent /usr/local/bin/
	sudo mkdir -p /etc/ispagent
	sudo cp configs/agent.yaml.example /etc/ispagent/agent.yaml
