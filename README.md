<p align="center">
  <img src="https://img.shields.io/github/v/release/msalah0e/palm?style=flat-square&color=2DB682" alt="Release">
  <img src="https://img.shields.io/github/actions/workflow/status/msalah0e/palm/ci.yml?style=flat-square&label=CI" alt="CI">
  <img src="https://img.shields.io/github/license/msalah0e/palm?style=flat-square" alt="License">
  <img src="https://img.shields.io/github/stars/msalah0e/palm?style=flat-square&color=2DB682" alt="Stars">
  <img src="https://img.shields.io/badge/tools-102+-2DB682?style=flat-square" alt="Tools">
  <img src="https://img.shields.io/badge/models-30+-2DB682?style=flat-square" alt="Models">
</p>

<h1 align="center">palm 🌴</h1>
<p align="center"><strong>The AI tool manager and control plane.</strong></p>
<p align="center">
  Discover, install, configure, run, budget, and benchmark 100+ AI CLI tools from one place.<br>
  API key vault. LLM proxy. Model management. Spending controls. Tool pipelines.
</p>

---

## Install

```bash
curl -fsSL https://msalah0e.github.io/palm/install.sh | sh
```

Or with Go:

```bash
go install github.com/msalah0e/palm@latest
```

## Quick Start

```bash
# Discover 102 AI tools across 13 categories
palm discover

# Install tools (parallel by default, auto-detects backend)
palm install ollama aider claude-code

# Store API keys securely in vault
palm keys add ANTHROPIC_API_KEY
palm keys add OPENAI_API_KEY

# Run a tool with vault keys auto-injected
palm run aider

# Pin tools to your project
palm workspace init
palm workspace add aider claude-code
palm workspace install

# Generate context files for AI tools
palm context init

# See everything at a glance
palm matrix
```

## Core Features

### Tool Management
```bash
palm install <tool...>          # Install tools (parallel by default)
palm remove <tool>              # Remove a tool
palm update [tool|--all]        # Update tool(s)
palm list                       # List installed tools
palm search <query>             # Search the registry
palm info <tool>                # Detailed tool info
palm discover                   # Browse curated catalog
palm doctor                     # Health check
```

### Run & Pipe
```bash
palm run aider                  # Run with vault keys injected
palm pipe "echo 'explain quicksort'" "|" "ollama run llama3.3"
```

### API Key Vault
```bash
palm keys add ANTHROPIC_API_KEY # Store in macOS Keychain or encrypted file
palm keys list                  # Show stored keys (masked)
palm keys export                # Print export statements
palm env                        # Shell integration: eval $(palm env)
```

### Workspace & Context
```bash
palm workspace init             # Create .palm.toml in project
palm workspace add aider        # Pin tools to project
palm workspace install          # Install all pinned tools
palm workspace status           # Show what's installed

palm context init               # Generate AI context files
palm context sync               # Sync .palm-context.md to tool files
```

### Models & Providers
```bash
palm models list                # List 30+ models across 6 providers
palm models list -p openai      # Filter by provider
palm models info gpt-4o         # Model details and pricing
palm models pull llama3.3       # Pull local models via Ollama
palm models providers           # Show provider status
```

### Budget & Cost Tracking
```bash
palm budget set --monthly 50    # Set monthly spending limit
palm budget set --daily 10      # Set daily limit
palm budget status              # Current spend vs limit
palm sessions                   # View session history
palm sessions --cost            # Cost breakdown by tool
```

### LLM Proxy
```bash
palm proxy start                # Start local API proxy on :4778
palm proxy start --bg           # Run in background
palm proxy status               # Check if running
palm proxy logs                 # View request logs
palm proxy stop                 # Stop the proxy

# Route API calls through palm proxy
export OPENAI_BASE_URL=http://localhost:4778/openai/v1
export ANTHROPIC_BASE_URL=http://localhost:4778/anthropic/v1
```

### Benchmark
```bash
palm benchmark "explain quicksort" --tools ollama,aider
palm benchmark "write a haiku" --tools ollama,aider --output
```

### Offline Mode
```bash
palm fetch aider ollama         # Pre-download to cache
palm fetch --all                # Cache everything
palm bundle tools.tar.gz        # Create portable archive
palm --offline install aider    # Install from cache
```

## All Commands

```
palm install <tool...>          Install AI tool(s) — parallel by default
palm remove <tool>              Remove an AI tool
palm update [tool|--all]        Update AI tool(s)
palm list                       List installed AI tools
palm search <query>             Search the registry
palm info <tool>                Detailed tool info
palm run <tool> [args...]       Run tool with vault keys injected
palm pipe <cmd> | <cmd>         Chain AI tools together
palm doctor                     Health check (tools + keys + runtimes)
palm keys [add|rm|list|export]  Manage API keys
palm env                        Shell exports for eval $(palm env)
palm workspace [init|add|rm|install|status]  Project tool pinning
palm context [init|show|sync]   AI tool context management
palm models [list|info|pull|providers]  LLM model management
palm budget [set|status|reset]  Spending controls
palm proxy [start|stop|status|logs]  Local LLM API proxy
palm benchmark <prompt>         Compare AI tools
palm sessions                   Session history & costs
palm matrix                     Terminal dashboard
palm discover                   Browse curated catalog
palm fetch [tool...|--all]      Pre-download for offline use
palm bundle <output.tar.gz>     Create portable tool bundle
palm stats                      Usage statistics
palm self-update                Update palm itself
palm completion <shell>         Shell completions (zsh/bash/fish)
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

## LLM Providers

| Provider | Models | Key |
|----------|--------|-----|
| **OpenAI** | GPT-4o, GPT-4.1, o3, o4-mini | `OPENAI_API_KEY` |
| **Anthropic** | Claude Opus 4.6, Sonnet 4.5, Haiku 4.5 | `ANTHROPIC_API_KEY` |
| **Google** | Gemini 2.5 Pro/Flash, Gemini 2.0 Flash | `GOOGLE_API_KEY` |
| **Ollama** | Llama 3.3, Qwen 3, DeepSeek R1, Mistral | (local) |
| **Groq** | Llama 3.3 70B, DeepSeek R1 70B | `GROQ_API_KEY` |
| **Mistral** | Mistral Large, Codestral | `MISTRAL_API_KEY` |

## Install Backends

| Backend | Tools | Examples |
|---------|-------|---------|
| **pip/uv** | Python tools | aider, crewai, deepeval |
| **npm** | Node.js tools | claude-code, codex |
| **go** | Go tools | mods, fabric, opencode |
| **cargo** | Rust tools | qdrant |
| **docker** | Containerized | vllm, localai, chromadb |
| **brew** | macOS packages | ollama, k8sgpt |
| **script** | curl installers | plandex |

## Configuration

### Global: `~/.config/palm/config.toml`

```toml
[parallel]
enabled = true
concurrency = 4

[install]
prefer_uv = true

[hooks]
pre_install = ""
post_install = ""
```

### Project: `.palm.toml`

```toml
[workspace]
name = "my-project"
tools = ["aider", "claude-code"]
keys = ["ANTHROPIC_API_KEY"]

[parallel]
concurrency = 2
```

### Custom tools: `~/.config/palm/plugins/*.toml`

```toml
[[tools]]
name = "my-tool"
display_name = "My Tool"
description = "My custom AI tool"
category = "coding"

[tools.install]
pip = "my-tool"

[tools.install.verify]
command = "my-tool --version"

[tools.keys]
required = ["MY_API_KEY"]
```

## Shell Setup

```bash
# Zsh (add to ~/.zshrc)
eval "$(palm completion zsh)"
eval "$(palm env)"

# Bash (add to ~/.bashrc)
eval "$(palm completion bash)"
eval "$(palm env)"

# Fish
palm completion fish | source
```

## Requirements

- macOS or Linux
- At least one: pip/uv, npm, cargo, go, or docker

## License

MIT
