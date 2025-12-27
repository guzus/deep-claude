package tmux

import (
	"strings"
	"testing"
)

func TestSanitizePrompt(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Add test coverage", "add-test-coverage"},
		{"Fix ALL bugs!!!", "fix-all-bugs"},
		{"Refactor auth", "refactor-auth"},
		{"", ""},
		{"   spaces   everywhere   ", "spaces-everywhere"},
		{"UPPERCASE and lowercase", "uppercase-and-lowercase"},
		{"Special!@#$%^&*()chars", "special-chars"},
		{"123 numbers first", "123-numbers-first"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizePrompt(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizePrompt(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateSessionName(t *testing.T) {
	name := GenerateSessionName("Add test coverage")

	// Should start with cc-
	if !strings.HasPrefix(name, SessionPrefix) {
		t.Errorf("session name should start with %q, got %q", SessionPrefix, name)
	}

	// Should contain the sanitized prompt
	if !strings.Contains(name, "add-test-coverage") {
		t.Errorf("session name should contain sanitized prompt, got %q", name)
	}

	// Should have datetime format (cc-YYMMDD-HHMM-...)
	parts := strings.Split(name, "-")
	if len(parts) < 4 {
		t.Errorf("session name should have at least 4 parts separated by -, got %q", name)
	}
}

func TestGenerateSessionNameEmpty(t *testing.T) {
	name := GenerateSessionName("")

	// Should still generate a valid name with just timestamp
	if !strings.HasPrefix(name, SessionPrefix) {
		t.Errorf("session name should start with %q, got %q", SessionPrefix, name)
	}
}

func TestSanitizePromptMaxLength(t *testing.T) {
	longPrompt := "This is a very long prompt that should be truncated to fit within the maximum allowed length"
	result := sanitizePrompt(longPrompt)

	if len(result) > MaxPromptLength {
		t.Errorf("sanitized prompt should be at most %d chars, got %d", MaxPromptLength, len(result))
	}

	// Should not end with a hyphen
	if strings.HasSuffix(result, "-") {
		t.Errorf("sanitized prompt should not end with hyphen, got %q", result)
	}
}
