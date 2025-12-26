#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="continuous-claude"
REPO="guzus/continuous-claude"
RELEASES_URL="https://github.com/${REPO}/releases/latest/download"

echo "🔂 Installing Continuous Claude..."

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    mingw*|msys*|cygwin*)
        echo -e "${RED}❌ Windows detected. Please download the binary manually from:${NC}"
        echo "   https://github.com/${REPO}/releases/latest"
        exit 1
        ;;
    *)
        echo -e "${RED}❌ Unsupported operating system: $OS${NC}"
        exit 1
        ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)
        echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

DOWNLOAD_URL="${RELEASES_URL}/${BINARY_NAME}-${OS}-${ARCH}"

echo "📦 Detected: ${OS}/${ARCH}"

# Create install directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Download the binary
echo "📥 Downloading ${BINARY_NAME} from ${DOWNLOAD_URL}..."
if ! curl -fsSL "$DOWNLOAD_URL" -o "$INSTALL_DIR/$BINARY_NAME"; then
    echo -e "${RED}❌ Failed to download $BINARY_NAME${NC}" >&2
    echo ""
    echo "This could mean:"
    echo "  - No release binaries are available yet"
    echo "  - Network connectivity issues"
    echo ""
    echo "Try building from source instead:"
    echo "  git clone https://github.com/${REPO}.git"
    echo "  cd continuous-claude && make build"
    exit 1
fi

# Make it executable
chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo -e "${GREEN}✅ $BINARY_NAME installed to $INSTALL_DIR/$BINARY_NAME${NC}"

# Check if install directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}⚠️  Warning: $INSTALL_DIR is not in your PATH${NC}"
    echo ""
    echo "To add it to your PATH, add this line to your shell profile:"
    echo ""

    # Detect shell
    if [[ "$SHELL" == *"zsh"* ]]; then
        echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
        echo "  source ~/.zshrc"
    elif [[ "$SHELL" == *"bash"* ]]; then
        echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
        echo "  source ~/.bashrc"
    else
        echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    fi
    echo ""
fi

# Check for dependencies
echo ""
echo "🔍 Checking dependencies..."

missing_deps=()

if ! command -v claude &> /dev/null; then
    missing_deps+=("Claude Code CLI")
fi

if ! command -v gh &> /dev/null; then
    missing_deps+=("GitHub CLI")
fi

if [ ${#missing_deps[@]} -eq 0 ]; then
    echo -e "${GREEN}✅ All dependencies installed${NC}"
else
    echo -e "${YELLOW}⚠️  Missing dependencies:${NC}"
    for dep in "${missing_deps[@]}"; do
        echo "   - $dep"
    done
    echo ""
    echo "Install them with:"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "  brew install gh"
        echo "  # Install Claude Code CLI: https://docs.anthropic.com/en/docs/claude-code"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "  # Install GitHub CLI: https://github.com/cli/cli#installation"
        echo "  # Install Claude Code CLI: https://docs.anthropic.com/en/docs/claude-code"
    fi
fi

echo ""
echo -e "${GREEN}🎉 Installation complete!${NC}"
echo ""
echo "Get started with:"
echo "  $BINARY_NAME -p \"your task\" --max-runs 5"
echo ""
echo "For more information, visit: https://github.com/${REPO}"
