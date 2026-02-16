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

# Install tools (auto-detects backend: brew, pip/uv, npm, cargo, go)
tamr install ollama
tamr install aider claude-code    # install multiple at once

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
tamr install <tool...>       Install AI tool(s)
tamr remove <tool>           Remove an AI tool
tamr update [tool|--all]     Update AI tool(s)
tamr list                    List installed AI tools
tamr search <query>          Search the registry
tamr info <tool>             Detailed tool info
tamr run <tool> [args...]    Run tool with vault keys injected
tamr doctor                  Health check (tools + keys + runtimes)
tamr keys [add|rm|list|export]  Manage API keys
tamr discover                Browse curated catalog
tamr stats                   Usage statistics
tamr self-update             Update tamr itself
tamr completion <shell>      Shell completions (zsh/bash/fish)
```

## Configuration

Optional config at `~/.config/tamr/config.toml`:

```toml
[ui]
emoji = true
color = true

[install]
prefer_uv = true   # use uv over pip when available

[stats]
enabled = false     # local usage tracking

[hooks]
pre_install = ""    # shell command to run before install
post_install = ""   # shell command to run after install
```

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
- One of: Homebrew, pip/uv, npm, cargo, or go (for installing tools)

## License

MIT
