BINARY = palm
INSTALL_DIR = $(HOME)/.local/bin
VERSION = 1.5.0

.PHONY: build install clean test test-e2e completions build-all

build:
	go build -ldflags "-X github.com/msalah0e/palm/cmd.version=$(VERSION)" -o $(BINARY) .

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/$(BINARY)

clean:
	rm -f $(BINARY) palm-*

test:
	go test ./...

test-e2e:
	go test -tags e2e -count=1 -v -timeout 120s .

completions: build
	mkdir -p completions
	./$(BINARY) completion zsh > completions/palm.zsh
	./$(BINARY) completion bash > completions/palm.bash
	./$(BINARY) completion fish > completions/palm.fish
	./$(BINARY) completion powershell > completions/palm.ps1

build-all:
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X github.com/msalah0e/palm/cmd.version=$(VERSION)" -o palm-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X github.com/msalah0e/palm/cmd.version=$(VERSION)" -o palm-darwin-amd64 .
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X github.com/msalah0e/palm/cmd.version=$(VERSION)" -o palm-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X github.com/msalah0e/palm/cmd.version=$(VERSION)" -o palm-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X github.com/msalah0e/palm/cmd.version=$(VERSION)" -o palm-windows-amd64.exe .
