#!/bin/sh
# palm installer — https://msalah0e.github.io/palm/
# Usage: curl -fsSL https://msalah0e.github.io/palm/install.sh | sh

set -e

REPO="msalah0e/palm"
INSTALL_DIR="${PALM_INSTALL_DIR:-$HOME/.local/bin}"
BINARY="palm"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()  { printf "${CYAN}=>${NC} %s\n" "$1"; }
ok()    { printf "${GREEN}✓${NC} %s\n" "$1"; }
warn()  { printf "${YELLOW}⚠${NC} %s\n" "$1"; }
error() { printf "${RED}✗${NC} %s\n" "$1" >&2; exit 1; }

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        darwin) OS="darwin" ;;
        linux)  OS="linux" ;;
        *)      error "Unsupported OS: $OS" ;;
    esac

    case "$ARCH" in
        x86_64|amd64)  ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *)             error "Unsupported architecture: $ARCH" ;;
    esac

    PLATFORM="${OS}_${ARCH}"
}

# Get latest version from GitHub
get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    else
        error "Neither curl nor wget found"
    fi

    if [ -z "$VERSION" ]; then
        # Fallback: try to build from source
        warn "Could not fetch latest version from GitHub"
        install_from_source
        exit 0
    fi
}

# Download and install binary
install_binary() {
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}_${PLATFORM}.tar.gz"

    info "Downloading palm ${VERSION} for ${PLATFORM}..."

    TMPDIR=$(mktemp -d)
    trap 'rm -rf "$TMPDIR"' EXIT

    if command -v curl >/dev/null 2>&1; then
        HTTP_CODE=$(curl -fsSL -o "${TMPDIR}/${BINARY}.tar.gz" -w "%{http_code}" "$DOWNLOAD_URL" 2>/dev/null || echo "000")
    elif command -v wget >/dev/null 2>&1; then
        wget -q -O "${TMPDIR}/${BINARY}.tar.gz" "$DOWNLOAD_URL" 2>/dev/null && HTTP_CODE="200" || HTTP_CODE="000"
    fi

    if [ "$HTTP_CODE" != "200" ]; then
        warn "No pre-built binary available for ${PLATFORM}"
        install_from_source
        return
    fi

    tar -xzf "${TMPDIR}/${BINARY}.tar.gz" -C "$TMPDIR"

    mkdir -p "$INSTALL_DIR"
    mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    chmod +x "${INSTALL_DIR}/${BINARY}"

    ok "palm ${VERSION} installed to ${INSTALL_DIR}/${BINARY}"
}

# Fallback: install from source using Go
install_from_source() {
    if ! command -v go >/dev/null 2>&1; then
        error "Go is required to install palm from source. Install Go first: https://go.dev/dl/"
    fi

    info "Installing palm from source via go install..."
    go install "github.com/${REPO}@latest"
    ok "palm installed via go install"

    # Find where go installed it
    GOBIN=$(go env GOBIN)
    if [ -z "$GOBIN" ]; then
        GOBIN="$(go env GOPATH)/bin"
    fi

    if [ -f "${GOBIN}/${BINARY}" ]; then
        ok "Binary at ${GOBIN}/${BINARY}"
    fi
    return
}

# Check PATH
check_path() {
    case ":$PATH:" in
        *":${INSTALL_DIR}:"*) ;;
        *)
            echo ""
            warn "${INSTALL_DIR} is not in your PATH"
            echo ""
            echo "  Add this to your shell config (~/.zshrc or ~/.bashrc):"
            echo ""
            echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
            echo ""
            ;;
    esac
}

main() {
    printf "\n${BOLD}🌴 palm installer${NC}\n\n"

    detect_platform
    info "Platform: ${PLATFORM}"

    get_latest_version
    info "Version: ${VERSION}"

    install_binary
    check_path

    echo ""
    printf "${GREEN}${BOLD}Done!${NC} Run ${CYAN}palm --help${NC} to get started.\n\n"
}

main "$@"
