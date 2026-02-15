BINARY=bin/archway

.PHONY: build test lint clean release snapshot install-local

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

install-local:
	go install ./cmd/archway
