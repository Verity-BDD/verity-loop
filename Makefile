BINARY := verity-loop
CMD     := ./cmd/verity-loop
BIN_DIR := bin

.PHONY: build test test-unit test-e2e lint clean

build:
	go build -o $(BIN_DIR)/$(BINARY) $(CMD)

test:
	go test ./...

test-unit:
	go test ./internal/... ./cmd/...

test-e2e:
	go test ./e2e/...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BIN_DIR)
