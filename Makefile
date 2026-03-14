BIN_DIR := bin
TOOLS := agentq agentenv agentfs agentrun agentedit agentdoctor agentrunpod agentfal agentprint agentmd

.PHONY: build test fmtcheck check clean

build:
	mkdir -p $(BIN_DIR)
	for tool in $(TOOLS); do \
		go build -o $(BIN_DIR)/$$tool ./cmd/$$tool; \
	done

test:
	go test ./...

fmtcheck:
	test -z "$$(gofmt -l .)"

check: fmtcheck test

clean:
	rm -rf $(BIN_DIR)
