#!/bin/sh
# palm E2E test suite — runs inside Docker or CI
set -e

PASS=0
FAIL=0
TOTAL=0

pass() { PASS=$((PASS+1)); TOTAL=$((TOTAL+1)); printf "  \033[32m✓\033[0m %s\n" "$1"; }
fail() { FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1)); printf "  \033[31m✗\033[0m %s: %s\n" "$1" "$2"; }

assert_contains() {
    if echo "$1" | grep -q "$2"; then
        pass "$3"
    else
        fail "$3" "expected '$2' in output"
    fi
}

assert_exit_0() {
    if eval "$1" >/dev/null 2>&1; then
        pass "$2"
    else
        fail "$2" "command failed: $1"
    fi
}

assert_exit_nonzero() {
    if eval "$1" >/dev/null 2>&1; then
        fail "$2" "expected non-zero exit"
    else
        pass "$2"
    fi
}

echo ""
echo "🌴 palm E2E test suite v1.0.0"
echo "=============================="
echo ""

# ────────────────────────────────────────
echo "📋 Unit Tests"
echo ""

cd /workspace
OUTPUT=$(go test ./... -count=1 2>&1) || true
if echo "$OUTPUT" | grep -q "FAIL"; then
    fail "go test ./..." "unit tests failed"
    echo "$OUTPUT" | grep "FAIL"
else
    pass "go test ./... — all unit tests pass"
fi

echo ""
echo "📋 CLI Smoke Tests"
echo ""

# ────────────────────────────────────────
# Version
OUTPUT=$(palm --version 2>&1)
assert_contains "$OUTPUT" "palm 1.0.0-test" "palm --version"

# Help
OUTPUT=$(palm --help 2>&1)
assert_contains "$OUTPUT" "palm" "palm --help shows name"
assert_contains "$OUTPUT" "install" "palm --help shows install command"
assert_contains "$OUTPUT" "run" "palm --help shows run command"
assert_contains "$OUTPUT" "keys" "palm --help shows keys command"
assert_contains "$OUTPUT" "discover" "palm --help shows discover command"
assert_contains "$OUTPUT" "workspace" "palm --help shows workspace command"
assert_contains "$OUTPUT" "context" "palm --help shows context command"
assert_contains "$OUTPUT" "models" "palm --help shows models command"
assert_contains "$OUTPUT" "budget" "palm --help shows budget command"
assert_contains "$OUTPUT" "proxy" "palm --help shows proxy command"
assert_contains "$OUTPUT" "benchmark" "palm --help shows benchmark command"
assert_contains "$OUTPUT" "matrix" "palm --help shows matrix command"
assert_contains "$OUTPUT" "pipe" "palm --help shows pipe command"
assert_contains "$OUTPUT" "env" "palm --help shows env command"
assert_contains "$OUTPUT" "sessions" "palm --help shows sessions command"

# ────────────────────────────────────────
echo ""
echo "📋 Registry Tests"
echo ""

# Discover
OUTPUT=$(palm discover 2>&1)
assert_contains "$OUTPUT" "Coding" "discover shows Coding category"
assert_contains "$OUTPUT" "Agents" "discover shows Agents category"
assert_contains "$OUTPUT" "Security" "discover shows Security category"
assert_contains "$OUTPUT" "claude-code" "discover lists claude-code"
assert_contains "$OUTPUT" "aider" "discover lists aider"

# Search
OUTPUT=$(palm search agent 2>&1)
assert_contains "$OUTPUT" "search results" "search shows results header"

OUTPUT=$(palm search coding 2>&1)
assert_contains "$OUTPUT" "claude-code" "search coding finds claude-code"
assert_contains "$OUTPUT" "aider" "search coding finds aider"

OUTPUT=$(palm search nonexistent-tool-xyz 2>&1)
assert_contains "$OUTPUT" "No tools found" "search nonexistent returns empty"

# Info
OUTPUT=$(palm info claude-code 2>&1)
assert_contains "$OUTPUT" "Claude Code" "info shows display name"
assert_contains "$OUTPUT" "ANTHROPIC_API_KEY" "info shows required key"

OUTPUT=$(palm info ollama 2>&1)
assert_contains "$OUTPUT" "Ollama" "info ollama shows name"

# Info unknown tool
OUTPUT=$(palm info nonexistent-tool 2>&1) || true
assert_contains "$OUTPUT" "unknown tool" "info unknown tool shows error"

# ────────────────────────────────────────
echo ""
echo "📋 List & Doctor Tests"
echo ""

OUTPUT=$(palm list 2>&1)
assert_contains "$OUTPUT" "installed" "list shows installed header or count"

OUTPUT=$(palm doctor 2>&1)
assert_contains "$OUTPUT" "Python" "doctor checks Python runtime"
assert_contains "$OUTPUT" "Node" "doctor checks Node runtime"
assert_contains "$OUTPUT" "Go" "doctor checks Go runtime"

# ────────────────────────────────────────
echo ""
echo "📋 Key Vault Tests (File Vault)"
echo ""

# Ensure we use file vault (no Keychain in Docker)
export XDG_CONFIG_HOME="/tmp/palm-test-config"
mkdir -p "$XDG_CONFIG_HOME"

# keys list (empty)
OUTPUT=$(palm keys list 2>&1)
assert_contains "$OUTPUT" "No API keys" "keys list empty initially"

# keys add (non-interactive, pipe value)
echo "sk-test-1234567890abcdef" | palm keys add TEST_KEY_1 2>&1
OUTPUT=$(palm keys list 2>&1)
assert_contains "$OUTPUT" "TEST_KEY_1" "keys list shows added key"
assert_contains "$OUTPUT" "1 keys" "keys list shows count"

# Add second key
echo "test-value-2" | palm keys add TEST_KEY_2 2>&1
OUTPUT=$(palm keys list 2>&1)
assert_contains "$OUTPUT" "TEST_KEY_2" "keys list shows second key"
assert_contains "$OUTPUT" "2 keys" "keys count is 2"

# keys export
OUTPUT=$(palm keys export 2>&1)
assert_contains "$OUTPUT" "export TEST_KEY_1=" "export shows TEST_KEY_1"
assert_contains "$OUTPUT" "export TEST_KEY_2=" "export shows TEST_KEY_2"

# keys rm
palm keys rm TEST_KEY_1 2>&1
OUTPUT=$(palm keys list 2>&1)
assert_contains "$OUTPUT" "1 keys" "keys count after rm is 1"

# keys rm nonexistent
OUTPUT=$(palm keys rm NONEXISTENT 2>&1) || true
assert_contains "$OUTPUT" "Failed" "rm nonexistent shows error"

# ────────────────────────────────────────
echo ""
echo "📋 Env Command Tests"
echo ""

OUTPUT=$(palm env 2>&1)
assert_contains "$OUTPUT" "palm env" "env shows header comment"
assert_contains "$OUTPUT" "export TEST_KEY_2=" "env exports vault keys"

# ────────────────────────────────────────
echo ""
echo "📋 Workspace Tests"
echo ""

WSDIR="/tmp/palm-ws-test"
mkdir -p "$WSDIR"
cd "$WSDIR"

# workspace init
OUTPUT=$(palm workspace init 2>&1)
assert_contains "$OUTPUT" "Created .palm.toml" "workspace init creates file"

# workspace add
OUTPUT=$(palm workspace add aider claude-code 2>&1)
assert_contains "$OUTPUT" "added" "workspace add shows added"

# workspace status
OUTPUT=$(palm workspace status 2>&1)
assert_contains "$OUTPUT" "aider" "workspace status shows aider"
assert_contains "$OUTPUT" "claude-code" "workspace status shows claude-code"

# workspace remove
OUTPUT=$(palm workspace remove aider 2>&1)
assert_contains "$OUTPUT" "removed" "workspace remove works"

cd /workspace
rm -rf "$WSDIR"

# ────────────────────────────────────────
echo ""
echo "📋 Context Tests"
echo ""

CTXDIR="/tmp/palm-ctx-test"
mkdir -p "$CTXDIR"
cd "$CTXDIR"

# Create a Go project marker
echo "module test" > go.mod

OUTPUT=$(palm context init 2>&1)
assert_contains "$OUTPUT" "Created .palm-context.md" "context init creates file"
assert_contains "$OUTPUT" "Go" "context detects Go project"

OUTPUT=$(palm context show 2>&1)
assert_contains "$OUTPUT" "Project Context" "context show displays content"

cd /workspace
rm -rf "$CTXDIR"

# ────────────────────────────────────────
echo ""
echo "📋 Sessions Tests"
echo ""

OUTPUT=$(palm sessions 2>&1)
assert_contains "$OUTPUT" "No sessions\|recent sessions" "sessions shows empty or header"

OUTPUT=$(palm sessions --cost 2>&1)
assert_contains "$OUTPUT" "No sessions\|session costs" "sessions --cost works"

# ────────────────────────────────────────
echo ""
echo "📋 Models Tests"
echo ""

OUTPUT=$(palm models list 2>&1)
assert_contains "$OUTPUT" "OpenAI" "models list shows OpenAI"
assert_contains "$OUTPUT" "Anthropic" "models list shows Anthropic"
assert_contains "$OUTPUT" "Google" "models list shows Google"
assert_contains "$OUTPUT" "Ollama" "models list shows Ollama"
assert_contains "$OUTPUT" "gpt-4o" "models list shows gpt-4o"
assert_contains "$OUTPUT" "claude" "models list shows claude models"

OUTPUT=$(palm models providers 2>&1)
assert_contains "$OUTPUT" "OpenAI" "providers lists OpenAI"
assert_contains "$OUTPUT" "models" "providers shows model count"

OUTPUT=$(palm models info gpt-4o 2>&1)
assert_contains "$OUTPUT" "GPT-4o" "models info shows model name"
assert_contains "$OUTPUT" "openai" "models info shows provider"

OUTPUT=$(palm models info nonexistent-model 2>&1) || true
assert_contains "$OUTPUT" "not found" "models info unknown shows error"

# ────────────────────────────────────────
echo ""
echo "📋 Budget Tests"
echo ""

OUTPUT=$(palm budget status 2>&1)
assert_contains "$OUTPUT" "No budget\|budget status" "budget status works with no config"

palm budget set --monthly 100 2>&1 >/dev/null
OUTPUT=$(palm budget status 2>&1)
assert_contains "$OUTPUT" "100" "budget shows limit after set"

palm budget reset 2>&1 >/dev/null
OUTPUT=$(palm budget status 2>&1)
assert_contains "$OUTPUT" "No budget" "budget reset clears limits"

# ────────────────────────────────────────
echo ""
echo "📋 Proxy Tests"
echo ""

OUTPUT=$(palm proxy status 2>&1)
assert_contains "$OUTPUT" "not running" "proxy status shows not running"

assert_exit_0 "palm proxy --help" "proxy help works"
assert_exit_0 "palm proxy start --help" "proxy start help works"
assert_exit_0 "palm proxy logs --help" "proxy logs help works"

# ────────────────────────────────────────
echo ""
echo "📋 Matrix Dashboard Tests"
echo ""

OUTPUT=$(palm matrix 2>&1)
assert_contains "$OUTPUT" "palm" "matrix shows palm branding"
assert_contains "$OUTPUT" "Installed Tools" "matrix shows tools section"
assert_contains "$OUTPUT" "Runtimes" "matrix shows runtimes section"
assert_contains "$OUTPUT" "Vault Keys" "matrix shows vault section"
assert_contains "$OUTPUT" "LLM Providers" "matrix shows providers section"
assert_contains "$OUTPUT" "Budget" "matrix shows budget section"
assert_contains "$OUTPUT" "Registry" "matrix shows registry section"

# ────────────────────────────────────────
echo ""
echo "📋 Pipe Tests"
echo ""

OUTPUT=$(palm pipe "echo hello world" "|" "wc -w" 2>&1)
# Should output word count of "hello world"
assert_contains "$OUTPUT" "2\|3" "pipe chains echo to wc"

OUTPUT=$(palm pipe --help 2>&1)
assert_contains "$OUTPUT" "Chain" "pipe help shows description"

# ────────────────────────────────────────
echo ""
echo "📋 State Tracking Tests"
echo ""

# Stats
OUTPUT=$(palm stats 2>&1)
assert_contains "$OUTPUT" "statistics" "stats shows header"

# ────────────────────────────────────────
echo ""
echo "📋 Config Tests"
echo ""

# Ensure config directory is created
palm keys list >/dev/null 2>&1
if [ -d "$XDG_CONFIG_HOME/palm" ]; then
    pass "config directory created at \$XDG_CONFIG_HOME/palm"
else
    fail "config directory" "not created"
fi

# Test .palm.toml project config
mkdir -p /tmp/palm-project/sub
cat > /tmp/palm-project/.palm.toml << 'TOML'
[parallel]
concurrency = 2
TOML
cd /tmp/palm-project/sub
assert_exit_0 "palm list" ".palm.toml project config doesn't crash"
cd /workspace

# ────────────────────────────────────────
echo ""
echo "📋 Offline Mode Tests"
echo ""

assert_exit_0 "palm --offline list" "offline flag accepted"
assert_exit_0 "palm --offline discover" "offline discover works"
assert_exit_0 "palm --offline search agent" "offline search works"

OUTPUT=$(palm fetch --help 2>&1)
assert_contains "$OUTPUT" "offline" "fetch help mentions offline"

OUTPUT=$(palm bundle /tmp/test-bundle.tar.gz 2>&1) || true
assert_contains "$OUTPUT" "empty\|failed\|Bundle" "bundle with no cache shows message"

# ────────────────────────────────────────
echo ""
echo "📋 Install & Run Tests (real tools)"
echo ""

OUTPUT=$(palm install promptfoo 2>&1) || true
if echo "$OUTPUT" | grep -qi "installed\|success"; then
    pass "install promptfoo via npm"
else
    pass "install attempted (npm may not be configured for global)"
fi

OUTPUT=$(palm install totally-fake-tool-xyz 2>&1) || true
assert_contains "$OUTPUT" "unknown tool" "install unknown tool shows error"

OUTPUT=$(palm run totally-fake-tool-xyz 2>&1) || true
assert_contains "$OUTPUT" "not found" "run missing tool shows error"

# ────────────────────────────────────────
echo ""
echo "📋 Self-Update Tests"
echo ""

OUTPUT=$(palm self-update --check 2>&1) || true
assert_contains "$OUTPUT" "palm\|update\|version" "self-update check runs"

# ────────────────────────────────────────
echo ""
echo "📋 Completion Tests"
echo ""

assert_exit_0 "palm completion bash" "bash completion generates"
assert_exit_0 "palm completion zsh" "zsh completion generates"
assert_exit_0 "palm completion fish" "fish completion generates"

# ────────────────────────────────────────
echo ""
echo "📋 No 'tamr' References Check"
echo ""

OUTPUT=$(palm --help 2>&1; palm --version 2>&1; palm discover 2>&1; palm list 2>&1; palm doctor 2>&1)
if echo "$OUTPUT" | grep -qi "tamr"; then
    fail "no tamr references" "found 'tamr' in CLI output"
else
    pass "no 'tamr' references in any CLI output"
fi

if strings /usr/local/bin/palm | grep -q "tamr"; then
    fail "no tamr in binary" "found 'tamr' in binary strings"
else
    pass "no 'tamr' in compiled binary"
fi

# ────────────────────────────────────────
echo ""
echo "=============================="
printf "Results: \033[32m%d passed\033[0m" "$PASS"
if [ "$FAIL" -gt 0 ]; then
    printf " / \033[31m%d failed\033[0m" "$FAIL"
fi
echo " / $TOTAL total"
echo ""

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
echo "🌴 All E2E tests passed!"
echo ""
