APP := filecairn
CMD := ./cmd/filecairn
BIN_DIR := ./bin
OUT := $(BIN_DIR)/$(APP)

.PHONY: build install test fmt run clean

build:
	mkdir -p $(BIN_DIR)
	go build -o $(OUT) $(CMD)

install:
	./scripts/install.sh

test:
	go test ./...

fmt:
	go fmt ./...

run:
	go run $(CMD) help

clean:
	rm -rf $(BIN_DIR)
