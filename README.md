# tamr 🌴

A tastier package manager for the AI era.

tamr wraps Homebrew with a unified interface for discovering, installing, and managing AI CLI tools — all from one place.

## Install

```bash
go install github.com/msalah0e/tamr@latest
```

Or build from source:

```bash
git clone https://github.com/msalah0e/tamr.git
cd tamr
make install  # builds and copies to ~/.local/bin/
```

## Usage

### Homebrew (fully compatible)

tamr passes through all brew commands — use it as a drop-in replacement:

```bash
tamr install ripgrep
tamr upgrade
tamr search jq
tamr doctor
```

### AI Tools

```bash
# Discover what's available
tamr ai discover

# Search the registry (47 tools across 6 categories)
tamr ai search agent

# Install a tool (auto-detects backend: brew, pip/uv, npm, cargo, go)
tamr ai install ollama
tamr ai install aider        # installs via pip/uv
tamr ai install claude-code  # installs via npm

# See what's installed
tamr ai list

# Detailed info about a tool
tamr ai info claude-code

# Health check — tools, API keys, runtimes
tamr ai doctor

# Remove a tool
tamr ai remove aider
```

### API Key Management

Store API keys in macOS Keychain:

```bash
tamr ai keys add ANTHROPIC_API_KEY
tamr ai keys add OPENAI_API_KEY
tamr ai keys list              # shows masked values
tamr ai keys rm OPENAI_API_KEY

# Export to shell (eval-able)
eval $(tamr ai keys export)
```

## Registry

47 tools across 6 categories:

| Category | Tools | Examples |
|----------|-------|---------|
| **Coding** | 8 | claude-code, aider, copilot-cli, codex |
| **LLM** | 8 | ollama, llm, llamafile, localai |
| **Agents** | 8 | fabric, crewai, goose, shell-gpt |
| **Media** | 6 | whisper, bark, comfyui |
| **Infra** | 6 | vllm, litellm, mlflow, modal |
| **Data** | 7 | chromadb, qdrant, lancedb, datasette |

Tools are defined as TOML files in [`registry/`](registry/) and embedded at compile time.

## Configuration

Optional config at `~/.config/tamr/config.toml`:

```toml
[ui]
emoji = true
color = true
rebrand = true     # replace "Homebrew" with "Tamr" in output

[install]
prefer_uv = true   # use uv over pip when available

[stats]
enabled = false     # local usage tracking
```

## Shell Completions

```bash
# zsh
tamr completion zsh > "${fpath[1]}/_tamr"

# bash
tamr completion bash > /etc/bash_completion.d/tamr

# fish
tamr completion fish > ~/.config/fish/completions/tamr.fish
```

## Requirements

- macOS (uses Keychain for API key storage)
- [Homebrew](https://brew.sh)

## License

MIT
