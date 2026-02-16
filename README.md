# tamr 🌴

**The AI tool manager.** Discover, install, configure, and run 100+ AI CLI tools from one place.

## Install

```bash
curl -fsSL https://msalah0e.github.io/tamr/install.sh | sh
```

Or with Go:

```bash
go install github.com/msalah0e/tamr@latest
```

Or build from source:

```bash
git clone https://github.com/msalah0e/tamr.git
cd tamr
make install  # builds and copies to ~/.local/bin/
```

## Quick Start

```bash
# Browse the full catalog
tamr discover

# Search for tools
tamr search agent

# Install tools (auto-detects backend: pip/uv, npm, cargo, go, docker)
tamr install ollama
tamr install aider claude-code    # parallel install by default

# See what's installed
tamr list

# Run a tool with API keys auto-injected from vault
tamr run aider

# Update tools
tamr update --all

# Health check — tools, API keys, runtimes
tamr doctor
```

## API Key Management

Store API keys securely (macOS Keychain or encrypted file vault):

```bash
tamr keys add ANTHROPIC_API_KEY
tamr keys add OPENAI_API_KEY
tamr keys list              # shows masked values
tamr keys rm OPENAI_API_KEY

# Export to shell
eval $(tamr keys export)
```

## Offline Mode

Pre-download tools for airgapped environments:

```bash
tamr fetch aider ollama       # cache specific tools
tamr fetch --all              # cache everything
tamr bundle tools.tar.gz      # create portable archive
tamr install aider --offline  # install from cache
```

## Registry

102 tools across 13 categories:

| Category | Count | Examples |
|----------|-------|---------|
| **Coding** | 14 | claude-code, aider, codex, gemini-cli, opencode |
| **LLM & Inference** | 12 | ollama, llm, llamafile, lm-studio, exo |
| **Agents** | 14 | fabric, crewai, goose, mods, claude-squad |
| **Chat** | 5 | aichat, chatgpt-cli, elia, oterm |
| **Dev Tools** | 8 | cursor, windsurf, tabby, pr-agent |
| **Media** | 6 | whisper, bark, comfyui, stable-diffusion |
| **Infrastructure** | 7 | vllm, litellm, mlflow, modal, k8sgpt |
| **Data** | 7 | chromadb, qdrant, weaviate, lancedb |
| **Testing** | 7 | promptfoo, deepeval, ragas, inspect-ai |
| **Security** | 6 | garak, llm-guard, guardrails-ai, pyrit |
| **Observability** | 6 | langfuse, phoenix, helicone, logfire |
| **Search & RAG** | 5 | perplexica, khoj, anything-llm, memgpt |
| **Writing** | 5 | marker, docling, instructor, outlines |

Tools are defined as TOML files in [`registry/`](registry/) and embedded at compile time.

## Commands

```
tamr install <tool...>       Install AI tool(s) — parallel by default
tamr remove <tool>           Remove an AI tool
tamr update [tool|--all]     Update AI tool(s)
tamr list                    List installed AI tools
tamr search <query>          Search the registry
tamr info <tool>             Detailed tool info
tamr run <tool> [args...]    Run tool with vault keys injected
tamr doctor                  Health check (tools + keys + runtimes)
tamr keys [add|rm|list|export]  Manage API keys
tamr discover                Browse curated catalog
tamr fetch [tool...|--all]   Pre-download for offline use
tamr bundle <output.tar.gz>  Create portable tool bundle
tamr stats                   Usage statistics
tamr self-update             Update tamr itself
tamr completion <shell>      Shell completions (zsh/bash/fish)
```

## Install Backends

tamr auto-detects the best install method for each tool:

| Backend | Tools | Examples |
|---------|-------|---------|
| **pip/uv** | Python tools | aider, crewai, deepeval |
| **npm** | Node.js tools | claude-code, codex |
| **go** | Go tools | mods, fabric, opencode |
| **cargo** | Rust tools | qdrant |
| **docker** | Containerized tools | vllm, localai, chromadb |
| **brew** | macOS packages | ollama, k8sgpt |
| **script** | curl installers | plandex |

## Configuration

### Global config: `~/.config/tamr/config.toml`

```toml
[ui]
emoji = true
color = true

[install]
prefer_uv = true   # use uv over pip when available

[parallel]
enabled = true      # parallel multi-tool install
concurrency = 4     # max simultaneous installs

[hooks]
pre_install = ""    # shell command before install
post_install = ""   # shell command after install

[stats]
enabled = false     # local usage tracking
```

### Project config: `.tamr.toml`

Drop a `.tamr.toml` in any project to override global settings:

```toml
[hooks]
post_install = "echo 'Tool installed for this project'"

[install]
prefer_uv = false
```

tamr walks up from the current directory to find `.tamr.toml`.

## Extending

Add custom tools via external plugins at `~/.config/tamr/plugins/*.toml`:

```toml
[[tools]]
name = "my-tool"
display_name = "My Tool"
description = "My custom AI tool"
category = "coding"
tags = ["ai", "custom"]

[tools.install]
pip = "my-tool"
docker = "myorg/my-tool:latest"

[tools.install.verify]
command = "my-tool --version"

[tools.keys]
required = ["MY_API_KEY"]
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

- macOS or Linux
- At least one: pip/uv, npm, cargo, go, or docker

## License

MIT
