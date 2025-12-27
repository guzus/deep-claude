// Package cli provides the command-line interface for Continuous Claude.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/guzus/continuous-claude/internal/config"
	"github.com/guzus/continuous-claude/internal/git"
	"github.com/guzus/continuous-claude/internal/github"
	"github.com/guzus/continuous-claude/internal/orchestrator"
	"github.com/guzus/continuous-claude/internal/tmux"
	"github.com/guzus/continuous-claude/internal/ui"
	"github.com/guzus/continuous-claude/internal/version"
	"github.com/spf13/cobra"
)

var (
	appVersion   string
	appBuildDate string
	appGitCommit string
)

// Execute runs the CLI.
func Execute(ver, buildDate, gitCommit string) error {
	appVersion = ver
	appBuildDate = buildDate
	appGitCommit = gitCommit

	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "continuous-claude",
	Short: "Continuous Claude - Autonomous AI development with GitHub integration",
	Long: `Continuous Claude orchestrates Claude Code to run iteratively with full GitHub integration.

It automates the entire PR lifecycle for large, multi-iteration AI development tasks -
enabling Claude to autonomously create PRs, monitor CI/CD checks, handle reviews,
and merge changes while maintaining persistent context across runs.

Example:
  continuous-claude -p "Add comprehensive test coverage" --max-runs 5
  continuous-claude -p "Refactor authentication" --max-cost 10.00
  continuous-claude -p "Fix all linting errors" --max-duration 2h`,
	RunE: runMain,
}

var (
	// Required flags
	prompt string

	// Limit flags (at least one required)
	maxRuns     int
	maxCost     float64
	maxDuration string

	// Optional flags
	owner               string
	repo                string
	mergeStrategy       string
	gitBranchPrefix     string
	notesFile           string
	disableCommits      bool
	dryRun              bool
	completionSignal    string
	completionThreshold int
	worktree            string
	worktreeBaseDir     string
	cleanupWorktree     bool
	autoUpdate          bool
	disableUpdates      bool
	detach              bool
)

func init() {
	// Required
	rootCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Task description for Claude (required)")

	// Limits (at least one required)
	rootCmd.Flags().IntVarP(&maxRuns, "max-runs", "m", 0, "Maximum number of iterations (0 = unlimited)")
	rootCmd.Flags().Float64Var(&maxCost, "max-cost", 0, "Maximum cost in USD (0 = unlimited)")
	rootCmd.Flags().StringVar(&maxDuration, "max-duration", "", "Maximum duration (e.g., '2h', '30m', '1h30m')")

	// GitHub/Git options
	rootCmd.Flags().StringVar(&owner, "owner", "", "GitHub repository owner (auto-detected)")
	rootCmd.Flags().StringVar(&repo, "repo", "", "GitHub repository name (auto-detected)")
	rootCmd.Flags().StringVar(&mergeStrategy, "merge-strategy", "squash", "PR merge strategy: squash, merge, rebase")
	rootCmd.Flags().StringVar(&gitBranchPrefix, "git-branch-prefix", "continuous-claude/", "Branch name prefix")
	rootCmd.Flags().StringVar(&notesFile, "notes-file", "SHARED_TASK_NOTES.md", "Path to notes file for context")

	// Execution options
	rootCmd.Flags().BoolVar(&disableCommits, "disable-commits", false, "Run without creating commits/PRs")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate without making changes")
	rootCmd.Flags().StringVar(&completionSignal, "completion-signal", "CONTINUOUS_CLAUDE_PROJECT_COMPLETE", "Signal phrase for early stop")
	rootCmd.Flags().IntVar(&completionThreshold, "completion-threshold", 3, "Consecutive signals needed to stop")

	// Worktree options
	rootCmd.Flags().StringVar(&worktree, "worktree", "", "Name for git worktree (parallel execution)")
	rootCmd.Flags().StringVar(&worktreeBaseDir, "worktree-base-dir", "../continuous-claude-worktrees", "Base directory for worktrees")
	rootCmd.Flags().BoolVar(&cleanupWorktree, "cleanup-worktree", false, "Remove worktree after completion")

	// Update options
	rootCmd.Flags().BoolVar(&autoUpdate, "auto-update", false, "Automatically install updates")
	rootCmd.Flags().BoolVar(&disableUpdates, "disable-updates", false, "Skip update checks")

	// Detach mode
	rootCmd.Flags().BoolVarP(&detach, "detach", "d", false, "Run in background tmux session")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(listWorktreesCmd)
	rootCmd.AddCommand(sessionsCmd)
	rootCmd.AddCommand(attachCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(killCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Continuous Claude %s\n", appVersion)
		fmt.Printf("  Build date: %s\n", appBuildDate)
		fmt.Printf("  Git commit: %s\n", appGitCommit)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for and install updates",
	RunE: func(cmd *cobra.Command, args []string) error {
		printer := ui.NewPrinter(false)

		printer.Info("Checking for updates...")
		latestVersion, hasUpdate, err := version.CheckForUpdates(appVersion)
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}

		if !hasUpdate {
			printer.Success("You are running the latest version (%s)", appVersion)
			return nil
		}

		printer.Info("New version available: %s (current: %s)", latestVersion, appVersion)

		if !printer.Confirm("Would you like to update now?") {
			printer.Info("Update cancelled")
			return nil
		}

		printer.StartSpinner("Downloading update...")
		tmpPath, err := version.DownloadUpdate(latestVersion)
		printer.StopSpinner()

		if err != nil {
			return fmt.Errorf("failed to download update: %w", err)
		}

		printer.StartSpinner("Installing update...")
		if err := version.InstallUpdate(tmpPath); err != nil {
			printer.StopSpinner()
			return fmt.Errorf("failed to install update: %w", err)
		}
		printer.StopSpinner()

		printer.Success("Updated to version %s", latestVersion)
		return nil
	},
}

var listWorktreesCmd = &cobra.Command{
	Use:   "list-worktrees",
	Short: "List active git worktrees",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		gitClient := git.NewClient(cwd)

		worktrees, err := gitClient.WorktreeList()
		if err != nil {
			return err
		}

		if len(worktrees) == 0 {
			fmt.Println("No worktrees found")
			return nil
		}

		fmt.Println("Active worktrees:")
		for _, wt := range worktrees {
			fmt.Printf("  %s\n", wt)
		}
		return nil
	},
}

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List and select active continuous-claude tmux sessions",
	Long: `List active continuous-claude tmux sessions with interactive selection.

Use arrow keys or j/k to navigate, Enter to attach, q to cancel.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tmux.IsAvailable() {
			return fmt.Errorf("tmux is not installed")
		}

		sessions, err := tmux.ListSessions()
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			fmt.Println("No active sessions")
			return nil
		}

		// Interactive picker
		fmt.Println("Select a session to attach:")
		selected, err := tmux.PickSession(sessions)
		if err != nil {
			return err
		}

		if selected == "" {
			fmt.Println("Cancelled")
			return nil
		}

		return tmux.AttachSession(selected)
	},
}

var attachCmd = &cobra.Command{
	Use:   "attach [session-name]",
	Short: "Attach to a tmux session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionName := args[0]

		// Allow partial match - find session starting with the given name
		sessions, err := tmux.ListSessions()
		if err != nil {
			return err
		}

		var match string
		for _, s := range sessions {
			if s.Name == sessionName || strings.HasPrefix(s.Name, sessionName) {
				if match != "" {
					return fmt.Errorf("ambiguous session name '%s' - matches multiple sessions", sessionName)
				}
				match = s.Name
			}
		}

		if match == "" {
			fmt.Println("Session not found. Available sessions:")
			for _, s := range sessions {
				fmt.Printf("  %s\n", s.Name)
			}
			return fmt.Errorf("session '%s' not found", sessionName)
		}

		return tmux.AttachSession(match)
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs [session-name]",
	Short: "View logs from a tmux session (read-only)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionName := args[0]

		// Allow partial match
		sessions, err := tmux.ListSessions()
		if err != nil {
			return err
		}

		var match string
		for _, s := range sessions {
			if s.Name == sessionName || strings.HasPrefix(s.Name, sessionName) {
				if match != "" {
					return fmt.Errorf("ambiguous session name '%s' - matches multiple sessions", sessionName)
				}
				match = s.Name
			}
		}

		if match == "" {
			fmt.Println("Session not found. Available sessions:")
			for _, s := range sessions {
				fmt.Printf("  %s\n", s.Name)
			}
			return fmt.Errorf("session '%s' not found", sessionName)
		}

		// Get last 1000 lines of logs
		logs, err := tmux.GetSessionLogs(match, 1000)
		if err != nil {
			return err
		}

		fmt.Print(logs)
		return nil
	},
}

var killCmd = &cobra.Command{
	Use:   "kill [session-name]",
	Short: "Kill a tmux session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionName := args[0]

		// Allow partial match
		sessions, err := tmux.ListSessions()
		if err != nil {
			return err
		}

		var match string
		for _, s := range sessions {
			if s.Name == sessionName || strings.HasPrefix(s.Name, sessionName) {
				if match != "" {
					return fmt.Errorf("ambiguous session name '%s' - matches multiple sessions", sessionName)
				}
				match = s.Name
			}
		}

		if match == "" {
			fmt.Println("Session not found. Available sessions:")
			for _, s := range sessions {
				fmt.Printf("  %s\n", s.Name)
			}
			return fmt.Errorf("session '%s' not found", sessionName)
		}

		if err := tmux.KillSession(match); err != nil {
			return err
		}

		fmt.Printf("Killed session: %s\n", match)
		return nil
	},
}

func runMain(cmd *cobra.Command, args []string) error {
	// Get working directory
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Parse duration
	duration, err := config.ParseDuration(maxDuration)
	if err != nil {
		return err
	}

	// Build config
	cfg := &config.Config{
		Prompt:              prompt,
		MaxRuns:             maxRuns,
		MaxCost:             maxCost,
		MaxDuration:         duration,
		Owner:               owner,
		Repo:                repo,
		MergeStrategy:       mergeStrategy,
		GitBranchPrefix:     gitBranchPrefix,
		NotesFile:           notesFile,
		DisableCommits:      disableCommits,
		DryRun:              dryRun,
		CompletionSignal:    completionSignal,
		CompletionThreshold: completionThreshold,
		Worktree:            worktree,
		WorktreeBaseDir:     worktreeBaseDir,
		CleanupWorktree:     cleanupWorktree,
		AutoUpdate:          autoUpdate,
		DisableUpdates:      disableUpdates,
		ExtraClaudeArgs:     args, // Pass remaining args to Claude
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Handle detach mode - spawn tmux session and exit
	if detach {
		return runDetached(workDir, cfg)
	}

	printer := ui.NewPrinter(false)
	createdRepo, err := ensureGitHubRepo(printer, workDir)
	if err != nil {
		return err
	}
	if err := ensureInitialCommitAndPush(printer, workDir, createdRepo); err != nil {
		return err
	}

	// Check for updates (unless disabled)
	if !cfg.DisableUpdates {
		checkUpdates(cfg.AutoUpdate)
	}

	// Create and run orchestrator
	orch, err := orchestrator.New(cfg, workDir)
	if err != nil {
		return err
	}

	return orch.Run()
}

func ensureGitHubRepo(printer *ui.Printer, workDir string) (bool, error) {
	gitClient := git.NewClient(workDir)
	if gitClient.IsRepo() {
		return false, nil
	}

	printer.Warning("No git repository detected in %s", workDir)
	if !printer.Confirm("Initialize a git repo and create a GitHub repository with gh?") {
		return false, fmt.Errorf("not in a git repository")
	}

	repoName := repo
	if repoName == "" {
		repoName = filepath.Base(workDir)
		if !printer.Confirm(fmt.Sprintf("Use repository name %q?", repoName)) {
			repoName = printer.Prompt("Repository name")
			if repoName == "" {
				return false, fmt.Errorf("repository name is required")
			}
		}
	}

	private := !printer.Confirm("Create repository as public?")

	ghClient := github.NewClient("", "", workDir)
	if err := ghClient.CheckAuth(); err != nil {
		return false, err
	}
	if err := gitClient.InitRepo(); err != nil {
		return false, err
	}
	if err := ghClient.CreateRepo(repoName, private, owner); err != nil {
		return false, err
	}

	printer.Success("Created GitHub repository: %s", repoName)
	return true, nil
}

func ensureInitialCommitAndPush(printer *ui.Printer, workDir string, skipConfirm bool) error {
	gitClient := git.NewClient(workDir)
	if gitClient.HasCommits() {
		return nil
	}

	if !skipConfirm {
		printer.Info("No commits found. Creating a blank CLAUDE.md, committing, and pushing.")
	}

	claudePath := filepath.Join(workDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err != nil && os.IsNotExist(err) {
		if err := os.WriteFile(claudePath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create CLAUDE.md: %w", err)
		}
	}

	if err := gitClient.StageAll(); err != nil {
		return err
	}
	if err := gitClient.Commit("Initial commit"); err != nil {
		return err
	}

	branch, err := gitClient.CurrentBranch()
	if err != nil {
		return err
	}
	if err := gitClient.Push(branch); err != nil {
		return err
	}

	printer.Success("Pushed initial commit to origin")
	return nil
}

func checkUpdates(autoInstall bool) {
	printer := ui.NewPrinter(false)

	latestVersion, hasUpdate, err := version.CheckForUpdates(appVersion)
	if err != nil {
		// Silently ignore update check errors
		return
	}

	if !hasUpdate {
		return
	}

	if autoInstall {
		printer.Info("Installing update %s...", latestVersion)
		tmpPath, err := version.DownloadUpdate(latestVersion)
		if err != nil {
			printer.Warning("Failed to download update: %v", err)
			return
		}

		if err := version.InstallUpdate(tmpPath); err != nil {
			printer.Warning("Failed to install update: %v", err)
			return
		}

		printer.Success("Updated to %s, please restart", latestVersion)
		os.Exit(0)
	} else {
		printer.Info("New version available: %s (run 'continuous-claude update' to install)", latestVersion)
	}
}

// runDetached spawns a tmux session running continuous-claude and returns immediately.
func runDetached(workDir string, cfg *config.Config) error {
	printer := ui.NewPrinter(false)

	// Check tmux availability
	if !tmux.IsAvailable() {
		return fmt.Errorf("tmux is required for -d flag. Install with: brew install tmux (macOS) or apt install tmux (Linux)")
	}

	// Generate session name
	sessionName := tmux.GenerateSessionName(cfg.Prompt)

	// Build command arguments (same as current, but without -d)
	cmdArgs := buildCommandArgs(cfg)

	// Get the executable path
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Build full command
	fullCmd := append([]string{executable}, cmdArgs...)

	// Create tmux session
	if err := tmux.CreateSession(sessionName, fullCmd, workDir); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}

	printer.Success("Started session: %s", sessionName)
	printer.Info("View logs:   continuous-claude logs %s", sessionName)
	printer.Info("Attach:      continuous-claude attach %s", sessionName)
	printer.Info("Kill:        continuous-claude kill %s", sessionName)

	return nil
}

// buildCommandArgs builds the command line arguments from config (excluding -d flag).
func buildCommandArgs(cfg *config.Config) []string {
	var args []string

	// Required
	args = append(args, "-p", cfg.Prompt)

	// Limits
	if cfg.MaxRuns > 0 {
		args = append(args, "-m", fmt.Sprintf("%d", cfg.MaxRuns))
	}
	if cfg.MaxCost > 0 {
		args = append(args, "--max-cost", fmt.Sprintf("%.2f", cfg.MaxCost))
	}
	if cfg.MaxDuration > 0 {
		args = append(args, "--max-duration", config.FormatDuration(cfg.MaxDuration))
	}

	// GitHub/Git options
	if cfg.Owner != "" {
		args = append(args, "--owner", cfg.Owner)
	}
	if cfg.Repo != "" {
		args = append(args, "--repo", cfg.Repo)
	}
	if cfg.MergeStrategy != "squash" {
		args = append(args, "--merge-strategy", cfg.MergeStrategy)
	}
	if cfg.GitBranchPrefix != "continuous-claude/" {
		args = append(args, "--git-branch-prefix", cfg.GitBranchPrefix)
	}
	if cfg.NotesFile != "SHARED_TASK_NOTES.md" {
		args = append(args, "--notes-file", cfg.NotesFile)
	}

	// Execution options
	if cfg.DisableCommits {
		args = append(args, "--disable-commits")
	}
	if cfg.DryRun {
		args = append(args, "--dry-run")
	}
	if cfg.CompletionSignal != "CONTINUOUS_CLAUDE_PROJECT_COMPLETE" {
		args = append(args, "--completion-signal", cfg.CompletionSignal)
	}
	if cfg.CompletionThreshold != 3 {
		args = append(args, "--completion-threshold", fmt.Sprintf("%d", cfg.CompletionThreshold))
	}

	// Worktree options
	if cfg.Worktree != "" {
		args = append(args, "--worktree", cfg.Worktree)
	}
	if cfg.WorktreeBaseDir != "../continuous-claude-worktrees" {
		args = append(args, "--worktree-base-dir", cfg.WorktreeBaseDir)
	}
	if cfg.CleanupWorktree {
		args = append(args, "--cleanup-worktree")
	}

	// Update options
	if cfg.AutoUpdate {
		args = append(args, "--auto-update")
	}
	if cfg.DisableUpdates {
		args = append(args, "--disable-updates")
	}

	// Extra Claude args
	args = append(args, cfg.ExtraClaudeArgs...)

	return args
}
