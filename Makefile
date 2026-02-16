BINARY = tamr
INSTALL_DIR = $(HOME)/.local/bin
VERSION = 0.3.0

.PHONY: build install clean test completions

build:
	go build -ldflags "-X github.com/msalah0e/tamr/cmd.version=$(VERSION)" -o $(BINARY) .

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/$(BINARY)

clean:
	rm -f $(BINARY)

test:
	go test ./...

completions: build
	mkdir -p completions
	./$(BINARY) completion zsh > completions/tamr.zsh
	./$(BINARY) completion bash > completions/tamr.bash
	./$(BINARY) completion fish > completions/tamr.fish
