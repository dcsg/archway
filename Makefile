BINARY=bin/archway

.PHONY: build test lint vet clean release snapshot install-local help

build:
	go build -o $(BINARY) ./cmd/archway

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/

release:
	goreleaser release

snapshot:
	goreleaser release --snapshot --clean

vet:
	go vet ./...

install-local:
	go install ./cmd/archway

help:
	@echo "build          Build binary to bin/archway"
	@echo "test           Run all tests"
	@echo "lint           Run golangci-lint"
	@echo "vet            Run go vet"
	@echo "clean          Remove build artifacts"
	@echo "release        Create a release with GoReleaser"
	@echo "snapshot       Create a snapshot release (no publish)"
	@echo "install-local  Install to GOPATH/bin"
