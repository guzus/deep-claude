#!/bin/bash
# Install git hooks for deep-claude development

set -e

HOOKS_DIR="$(git rev-parse --git-dir)/hooks"

echo "Installing git hooks..."

# Pre-commit hook
cat > "$HOOKS_DIR/pre-commit" << 'EOF'
#!/bin/bash
# Pre-commit hook for deep-claude
# Runs linting and tests before allowing commit

set -e

echo "Running pre-commit checks..."

# Check if we have Go files staged
if git diff --cached --name-only | grep -q '\.go$'; then
    echo "Go files detected, running checks..."

    # Run go vet
    echo "  Running go vet..."
    go vet ./...

    # Run golangci-lint if available
    if command -v golangci-lint &> /dev/null; then
        echo "  Running golangci-lint..."
        golangci-lint run --timeout=5m
    else
        echo "  Warning: golangci-lint not found, skipping (install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)"
    fi

    # Run tests
    echo "  Running tests..."
    go test ./...

    echo "All checks passed!"
fi
EOF

chmod +x "$HOOKS_DIR/pre-commit"

echo "Git hooks installed successfully!"
echo ""
echo "Hooks installed:"
echo "  - pre-commit: Runs go vet, golangci-lint, and tests"
echo ""
echo "To skip hooks temporarily, use: git commit --no-verify"
