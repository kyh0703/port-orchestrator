.PHONY: build test run

build:
	go build -o bin/orchestrator ./cmd/orchestrator

test:
	go test ./...

run:
	go run ./cmd/orchestrator
