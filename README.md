# palm 🌴

**The AI tool manager.** Discover, install, configure, and run 100+ AI CLI tools from one place.

## Install

```bash
curl -fsSL https://msalah0e.github.io/palm/install.sh | sh
```

Or with Go:

```bash
go install github.com/msalah0e/palm@latest
```

Or build from source:

```bash
git clone https://github.com/msalah0e/palm.git
cd palm
make install  # builds and copies to ~/.local/bin/
```

## Quick Start

```bash
# Browse the full catalog
palm discover

# Search for tools
palm search agent

# Install tools (auto-detects backend: pip/uv, npm, cargo, go, docker)
palm install ollama
palm install aider claude-code    # parallel install by default

# See what's installed
palm list

# Run a tool with API keys auto-injected from vault
palm run aider

# Update tools
palm update --all

# Health check — tools, API keys, runtimes
palm doctor
```

## API Key Management

Store API keys securely (macOS Keychain or encrypted file vault):

```bash
palm keys add ANTHROPIC_API_KEY
palm keys add OPENAI_API_KEY
palm keys list              # shows masked values
palm keys rm OPENAI_API_KEY

# Export to shell
eval $(palm keys export)
```

## Offline Mode

Pre-download tools for airgapped environments:

```bash
palm fetch aider ollama       # cache specific tools
palm fetch --all              # cache everything
palm bundle tools.tar.gz      # create portable archive
palm install aider --offline  # install from cache
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
palm install <tool...>       Install AI tool(s) — parallel by default
palm remove <tool>           Remove an AI tool
palm update [tool|--all]     Update AI tool(s)
palm list                    List installed AI tools
palm search <query>          Search the registry
palm info <tool>             Detailed tool info
palm run <tool> [args...]    Run tool with vault keys injected
palm doctor                  Health check (tools + keys + runtimes)
palm keys [add|rm|list|export]  Manage API keys
palm discover                Browse curated catalog
palm fetch [tool...|--all]   Pre-download for offline use
palm bundle <output.tar.gz>  Create portable tool bundle
palm stats                   Usage statistics
palm self-update             Update palm itself
palm completion <shell>      Shell completions (zsh/bash/fish)
```

## Install Backends

palm auto-detects the best install method for each tool:

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

### Global config: `~/.config/palm/config.toml`

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

### Project config: `.palm.toml`

Drop a `.palm.toml` in any project to override global settings:

```toml
[hooks]
post_install = "echo 'Tool installed for this project'"

[install]
prefer_uv = false
```

palm walks up from the current directory to find `.palm.toml`.

## Extending

Add custom tools via external plugins at `~/.config/palm/plugins/*.toml`:

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
palm completion zsh > "${fpath[1]}/_palm"

# bash
palm completion bash > /etc/bash_completion.d/palm

# fish
palm completion fish > ~/.config/fish/completions/palm.fish
```

## Requirements

- macOS or Linux
- At least one: pip/uv, npm, cargo, go, or docker

## License

MIT
