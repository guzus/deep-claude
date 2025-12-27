// Package tmux provides functionality for managing tmux sessions.
package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	// SessionPrefix is the prefix for all continuous-claude tmux sessions.
	SessionPrefix = "cc-"
	// MaxPromptLength is the maximum length of the sanitized prompt in session names.
	MaxPromptLength = 30
)

// Session represents a tmux session.
type Session struct {
	Name      string
	Created   string
	Attached  bool
	WindowsCount int
}

// IsAvailable checks if tmux is installed and available.
func IsAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// GenerateSessionName creates a session name from the current time and prompt.
// Format: cc-{YYMMDD-HHMM}-{sanitized-prompt}
func GenerateSessionName(prompt string) string {
	timestamp := time.Now().Format("060102-1504")
	sanitized := sanitizePrompt(prompt)

	if sanitized == "" {
		return fmt.Sprintf("%s%s", SessionPrefix, timestamp)
	}
	return fmt.Sprintf("%s%s-%s", SessionPrefix, timestamp, sanitized)
}

// sanitizePrompt cleans the prompt for use in a session name.
// - Converts to lowercase
// - Replaces non-alphanumeric chars with hyphens
// - Takes first few words
// - Limits length
func sanitizePrompt(prompt string) string {
	// Convert to lowercase
	prompt = strings.ToLower(prompt)

	// Replace non-alphanumeric characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	prompt = reg.ReplaceAllString(prompt, "-")

	// Trim leading/trailing hyphens
	prompt = strings.Trim(prompt, "-")

	// Split into words and take first few
	words := strings.Split(prompt, "-")
	if len(words) > 4 {
		words = words[:4]
	}
	prompt = strings.Join(words, "-")

	// Limit length
	if len(prompt) > MaxPromptLength {
		prompt = prompt[:MaxPromptLength]
		// Don't end with a hyphen
		prompt = strings.TrimRight(prompt, "-")
	}

	return prompt
}

// CreateSession creates a new detached tmux session running the given command.
func CreateSession(name string, cmd []string, workDir string) error {
	if !IsAvailable() {
		return fmt.Errorf("tmux is required for -d flag. Install with: brew install tmux (macOS) or apt install tmux (Linux)")
	}

	// Check if session already exists
	if SessionExists(name) {
		// Append a random suffix
		name = fmt.Sprintf("%s-%d", name, time.Now().UnixNano()%1000)
	}

	// Build tmux command
	// tmux new-session -d -s <name> -c <workdir> <command>
	args := []string{
		"new-session",
		"-d",
		"-s", name,
		"-c", workDir,
	}
	args = append(args, cmd...)

	command := exec.Command("tmux", args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}

// SessionExists checks if a tmux session with the given name exists.
func SessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

// ListSessions returns all continuous-claude tmux sessions (those starting with cc-).
func ListSessions() ([]Session, error) {
	if !IsAvailable() {
		return nil, fmt.Errorf("tmux is not installed")
	}

	// List all sessions with format: name:created:attached:windows
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}:#{session_created_string}:#{session_attached}:#{session_windows}")
	output, err := cmd.Output()
	if err != nil {
		// No sessions exist
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []Session{}, nil
		}
		return nil, err
	}

	var sessions []Session
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 4)
		if len(parts) < 4 {
			continue
		}

		name := parts[0]
		// Only include cc- prefixed sessions
		if !strings.HasPrefix(name, SessionPrefix) {
			continue
		}

		windowsCount := 0
		fmt.Sscanf(parts[3], "%d", &windowsCount)

		sessions = append(sessions, Session{
			Name:         name,
			Created:      parts[1],
			Attached:     parts[2] == "1",
			WindowsCount: windowsCount,
		})
	}

	return sessions, nil
}

// AttachSession attaches to an existing tmux session.
func AttachSession(name string) error {
	if !IsAvailable() {
		return fmt.Errorf("tmux is not installed")
	}

	if !SessionExists(name) {
		return fmt.Errorf("session '%s' does not exist", name)
	}

	cmd := exec.Command("tmux", "attach-session", "-t", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetSessionLogs captures the pane content from a session.
func GetSessionLogs(name string, lines int) (string, error) {
	if !IsAvailable() {
		return "", fmt.Errorf("tmux is not installed")
	}

	if !SessionExists(name) {
		return "", fmt.Errorf("session '%s' does not exist", name)
	}

	// Capture pane content with history
	// -p prints to stdout, -S specifies start line (negative = history)
	args := []string{"capture-pane", "-t", name, "-p", "-S", fmt.Sprintf("-%d", lines)}
	cmd := exec.Command("tmux", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to capture logs: %w", err)
	}

	return string(output), nil
}

// KillSession kills a tmux session.
func KillSession(name string) error {
	if !IsAvailable() {
		return fmt.Errorf("tmux is not installed")
	}

	if !SessionExists(name) {
		return fmt.Errorf("session '%s' does not exist", name)
	}

	cmd := exec.Command("tmux", "kill-session", "-t", name)
	return cmd.Run()
}
