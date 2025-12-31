## ⚙️ How it works

Using Claude Code to drive iterative development, this script fully automates the PR lifecycle from code changes through to merged commits:

- Claude Code runs in a loop based on your prompt
- All changes are committed to a new branch
- A new pull request is created
- It waits for all required PR checks and code reviews to complete
- Once checks pass and reviews are approved, the PR is merged
- This process repeats until your task is complete
- A `SHARED_TASK_NOTES.md` file maintains continuity by passing context between iterations, enabling seamless handoffs across AI and human developers
- If multiple agents decide that the project is complete, the loop will stop early.

## 🚀 Quick start

### Installation

#### Option 1: One-liner install (recommended)

Install the latest release with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/guzus/deep-claude/main/install.sh | bash
```

This automatically detects your OS and architecture, downloads the appropriate binary, and installs it to `~/.local/bin`.

#### Option 2: Install via Go

If you have Go installed:

```bash
go install github.com/guzus/deep-claude/cmd/dclaude@latest
```

The binary will be installed to your `$GOPATH/bin` directory. Make sure it's in your PATH.

#### Option 3: Build from source

```bash
# Clone the repository
git clone https://github.com/guzus/deep-claude.git
cd deep-claude

# Build and install
make build
sudo mv build/dclaude /usr/local/bin/
```

#### Option 4: Download pre-built binary

Pre-built binaries are available on the [Releases](https://github.com/guzus/deep-claude/releases) page when attached to a release.

**Linux (amd64)**
```bash
curl -fsSL https://github.com/guzus/deep-claude/releases/latest/download/dclaude-linux-amd64 -o dclaude
chmod +x dclaude
sudo mv dclaude /usr/local/bin/
```

**Linux (arm64)**
```bash
curl -fsSL https://github.com/guzus/deep-claude/releases/latest/download/dclaude-linux-arm64 -o dclaude
chmod +x dclaude
sudo mv dclaude /usr/local/bin/
```

**macOS (Apple Silicon)**
```bash
curl -fsSL https://github.com/guzus/deep-claude/releases/latest/download/dclaude-darwin-arm64 -o dclaude
chmod +x dclaude
sudo mv dclaude /usr/local/bin/
```

**macOS (Intel)**
```bash
curl -fsSL https://github.com/guzus/deep-claude/releases/latest/download/dclaude-darwin-amd64 -o dclaude
chmod +x dclaude
sudo mv dclaude /usr/local/bin/
```

**Windows (amd64)**

Download `dclaude-windows-amd64.exe` from the [Releases](https://github.com/guzus/deep-claude/releases/latest) page and add it to your PATH.

#### Verify checksums (optional)

Each release includes `.sha256` checksum files. To verify your download:

```bash
# Download the checksum file
curl -fsSL https://github.com/guzus/deep-claude/releases/latest/download/dclaude-linux-amd64.sha256 -o dclaude.sha256

# Verify (adjust filename for your platform)
sha256sum -c dclaude.sha256
```

#### Uninstall

```bash
rm /usr/local/bin/dclaude
# or if installed via go install:
rm $(go env GOPATH)/bin/dclaude
```

### Prerequisites

Before using `dclaude`, you need:

1. **[Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code)** - Authenticate with `claude auth`
2. **[GitHub CLI](https://cli.github.com)** - Authenticate with `gh auth login`

### Usage

```bash
# Run with your prompt and max runs (owner and repo auto-detected from git remote)
dclaude -p "add unit tests until all code is covered" --max-runs 5

# Or explicitly specify the owner and repo
dclaude -p "add unit tests until all code is covered" --max-runs 5 --owner guzus --repo deep-claude

# Or run with a cost budget instead
dclaude -p "add unit tests until all code is covered" --max-cost 10.00

# Or run for a specific duration (time-boxed bursts)
dclaude -p "add unit tests until all code is covered" --max-duration 2h

# Check version
dclaude version

# Check for updates
dclaude update
```

## 🎯 Flags

- `-p, --prompt`: Task prompt for Claude Code (required)
- `-m, --max-runs`: Maximum number of iterations, use `0` for infinite (required unless --max-cost or --max-duration is provided)
- `--max-cost`: Maximum USD to spend (required unless --max-runs or --max-duration is provided)
- `--max-duration`: Maximum duration to run (e.g., `2h`, `30m`, `1h30m`) (required unless --max-runs or --max-cost is provided)
- `--owner`: GitHub repository owner (auto-detected from git remote if not provided)
- `--repo`: GitHub repository name (auto-detected from git remote if not provided)
- `--merge-strategy`: Merge strategy: `squash`, `merge`, or `rebase` (default: `squash`)
- `--git-branch-prefix`: Prefix for git branch names (default: `deep-claude/`)
- `--notes-file`: Path to shared task notes file (default: `SHARED_TASK_NOTES.md`)
- `--disable-commits`: Disable automatic git commits, PR creation, and merging (useful for testing)
- `--worktree <name>`: Run in a git worktree for parallel execution (creates if needed)
- `--worktree-base-dir <path>`: Base directory for worktrees (default: `../deep-claude-worktrees`)
- `--cleanup-worktree`: Remove worktree after completion
- `--list-worktrees`: List all active git worktrees and exit
- `--dry-run`: Simulate execution without making changes
- `--completion-signal <phrase>`: Phrase that agents output when entire project is complete (default: `DEEP_CLAUDE_PROJECT_COMPLETE`)
- `--completion-threshold <num>`: Number of consecutive completion signals required to stop early (default: `3`)
- `-d, --detach`: Run in a background tmux session (requires tmux)
- `--auto-update`: Automatically install updates when available
- `--disable-updates`: Skip update checks

Any additional flags you provide that are not recognized by `dclaude` will be automatically forwarded to the underlying `claude` command. For example, you can pass `--allowedTools`, `--model`, or any other Claude Code CLI flags.

## 📝 Examples

```bash
# Run 5 iterations (owner and repo auto-detected from git remote)
dclaude -p "improve code quality" -m 5

# Run infinitely until stopped
dclaude -p "add unit tests until all code is covered" -m 0

# Run until $10 budget exhausted
dclaude -p "add documentation" --max-cost 10.00

# Run for 2 hours (time-boxed burst)
dclaude -p "add unit tests" --max-duration 2h

# Run for 30 minutes
dclaude -p "refactor module" --max-duration 30m

# Run for 1 hour and 30 minutes
dclaude -p "add features" --max-duration 1h30m

# Run max 10 iterations or $5, whichever comes first
dclaude -p "refactor code" -m 10 --max-cost 5.00

# Combine duration and cost limits (whichever comes first)
dclaude -p "improve tests" --max-duration 1h --max-cost 5.00

# Use merge commits instead of squash
dclaude -p "add features" -m 5 --merge-strategy merge

# Use rebase strategy
dclaude -p "update dependencies" -m 3 --merge-strategy rebase

# Use custom branch prefix
dclaude -p "refactor code" -m 3 --git-branch-prefix "feature/"

# Use custom notes file
dclaude -p "add features" -m 5 --notes-file "PROJECT_CONTEXT.md"

# Test without creating commits or PRs
dclaude -p "test changes" -m 2 --disable-commits

# Pass additional Claude Code CLI flags (e.g., restrict tools)
dclaude -p "add features" -m 3 --allowedTools "Write,Read"

# Use a different model
dclaude -p "refactor code" -m 5 --model claude-haiku-4-5

# Enable early stopping when agents signal project completion
dclaude -p "add unit tests to all files" -m 50 --completion-threshold 3

# Use custom completion signal
dclaude -p "fix all bugs" -m 20 --completion-signal "ALL_BUGS_FIXED" --completion-threshold 2

# Explicitly specify owner and repo (useful if git remote is not set up or not a GitHub repo)
dclaude -p "add features" -m 5 --owner myuser --repo myproject

# Run in background (detached tmux session)
dclaude -d -p "add documentation" --max-runs 10

# Skip update checks for faster startup
dclaude -p "quick fix" -m 1 --disable-updates

# Auto-install updates when available
dclaude -p "long task" -m 20 --auto-update
```

### Background mode

Run dclaude in a detached tmux session so it continues running after you disconnect:

```bash
# Start in background
dclaude -d -p "add unit tests until all code is covered" --max-runs 10

# Manage sessions
dclaude sessions              # Interactive session picker
dclaude logs dc-*             # View logs from a session
dclaude attach dc-*           # Attach to a session
dclaude kill dc-*             # Kill a session
```

Sessions are named with the format `dc-{YYMMDD-HHMM}-{prompt-summary}` (e.g., `dc-250115-1430-add-unit-tests`). You can use partial names with the management commands.

### Running in parallel

Use git worktrees to run multiple instances simultaneously without conflicts:

```bash
# Terminal 1 (owner and repo auto-detected)
dclaude -p "Add unit tests" -m 5 --worktree tests

# Terminal 2 (simultaneously)
dclaude -p "Add docs" -m 5 --worktree docs
```

Each instance creates its own worktree at `../deep-claude-worktrees/<name>/`, pulls the latest changes, and runs independently. Worktrees persist for reuse.

```bash
# List worktrees
dclaude --list-worktrees

# Clean up after completion
dclaude -p "task" -m 1 --worktree temp --cleanup-worktree
```

## 📊 Example output

Here's what a successful run looks like:

```
🔄 (1/1) Starting iteration...
🌿 (1/1) Creating branch: deep-claude/iteration-1/2025-11-15-be939873
🤖 (1/1) Running Claude Code...
📝 (1/1) Output: Perfect! I've successfully completed this iteration of the testing project. Here's what I accomplished: [...]
💰 (1/1) Cost: $0.042
✅ (1/1) Work completed
🌿 (1/1) Creating branch: deep-claude/iteration-1/2025-11-15-be939873
💬 (1/1) Committing changes...
📦 (1/1) Changes committed on branch: deep-claude/iteration-1/2025-11-15-be939873
📤 (1/1) Pushing branch...
🔨 (1/1) Creating pull request...
🔍 (1/1) PR #893 created, waiting 5 seconds for GitHub to set up...
🔍 (1/1) Checking PR status (iteration 1/180)...
   📊 Found 6 check(s)
   🟢 2    🟡 4    🔴 0
   👁️  Review status: None
⏳ Waiting for: checks to complete
✅ (1/1) All PR checks and reviews passed
🔀 (1/1) Merging PR #893...
📥 (1/1) Pulling latest from main...
🗑️ (1/1) Deleting local branch: deep-claude/iteration-1/2025-11-15-be939873
✅ (1/1) PR #893 merged: Add unit tests for authentication module
🎉 Done with total cost: $0.042
```

## 🛠️ Development

### Building from source

```bash
# Clone the repository
git clone https://github.com/guzus/deep-claude.git
cd deep-claude

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Run linter
make lint

# Build for all platforms
make build-all
```

### Project structure

```
deep-claude/
├── cmd/dclaude/              # Main entry point
├── internal/
│   ├── cli/                  # Cobra CLI commands
│   ├── config/               # Configuration management
│   ├── git/                  # Git operations
│   ├── github/               # GitHub PR management
│   ├── claude/               # Claude Code integration
│   ├── notes/                # Shared notes handling
│   ├── orchestrator/         # Main loop logic
│   ├── ui/                   # Terminal output
│   └── version/              # Update management
├── Makefile                  # Build automation
└── go.mod                    # Go module
```

### Setting up pre-commit hooks

```bash
# Option 1: Using pre-commit framework
pip install pre-commit
pre-commit install

# Option 2: Simple git hooks
./scripts/install-hooks.sh
```

## 📃 License

[MIT](./LICENSE) ©️ [Anand Chowdhary](https://anandchowdhary.com), [guzus](https://github.com/guzus)
