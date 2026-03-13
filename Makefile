BIN_DIR := bin
TOOLS := agentq agentenv agentfs agentrun agentedit agentdoctor agentrunpod agentfal agentprint

.PHONY: build test fmtcheck check clean

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/agentq ./cmd/agentq
	go build -o $(BIN_DIR)/agentenv ./cmd/agentenv
	go build -o $(BIN_DIR)/agentfs ./cmd/agentfs
	go build -o $(BIN_DIR)/agentrun ./cmd/agentrun
	go build -o $(BIN_DIR)/agentedit ./cmd/agentedit
	go build -o $(BIN_DIR)/agentdoctor ./cmd/agentdoctor
	go build -o $(BIN_DIR)/agentrunpod ./cmd/agentrunpod
	go build -o $(BIN_DIR)/agentfal ./cmd/agentfal
	go build -o $(BIN_DIR)/agentprint ./cmd/agentprint

test:
	go test ./...

fmtcheck:
	test -z "$$(gofmt -l .)"

check: fmtcheck test

clean:
	rm -rf $(BIN_DIR)
