package tmux

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// PickSession displays an interactive session picker with arrow key navigation.
// Returns the selected session name or empty string if cancelled.
func PickSession(sessions []Session) (string, error) {
	if len(sessions) == 0 {
		return "", nil
	}

	// Get terminal state to restore later
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback to non-interactive if terminal is not available
		return sessions[0].Name, nil
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	selected := 0
	maxItems := len(sessions)

	// Clear and draw initial list
	drawPicker(sessions, selected)

	// Read input
	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return "", err
		}

		// Handle key input
		if n == 1 {
			switch buf[0] {
			case 'q', 3: // q or Ctrl+C
				clearPicker(maxItems)
				return "", nil
			case 13: // Enter
				clearPicker(maxItems)
				return sessions[selected].Name, nil
			case 'j', 'J': // vim-style down
				if selected < maxItems-1 {
					selected++
					drawPicker(sessions, selected)
				}
			case 'k', 'K': // vim-style up
				if selected > 0 {
					selected--
					drawPicker(sessions, selected)
				}
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
			// Arrow keys (escape sequences)
			switch buf[2] {
			case 65: // Up
				if selected > 0 {
					selected--
					drawPicker(sessions, selected)
				}
			case 66: // Down
				if selected < maxItems-1 {
					selected++
					drawPicker(sessions, selected)
				}
			}
		}
	}
}

func drawPicker(sessions []Session, selected int) {
	// Move cursor up to redraw
	fmt.Print("\033[?25l") // Hide cursor

	for i, s := range sessions {
		// Clear line and move to beginning
		fmt.Print("\r\033[K")

		status := "running"
		if s.Attached {
			status = "attached"
		}

		if i == selected {
			// Highlight selected
			fmt.Printf("\033[7m > %s (%s)\033[0m\n", s.Name, status)
		} else {
			fmt.Printf("   %s (%s)\n", s.Name, status)
		}
	}

	// Print controls hint
	fmt.Print("\r\033[K")
	fmt.Print("\033[90m↑/↓ or j/k: navigate | Enter: select | q: cancel\033[0m")

	// Move cursor back up
	fmt.Printf("\033[%dA", len(sessions))
}

func clearPicker(itemCount int) {
	// Clear all lines
	for i := 0; i <= itemCount; i++ {
		fmt.Print("\r\033[K") // Clear line
		if i < itemCount {
			fmt.Print("\033[B") // Move down
		}
	}
	// Move back up
	fmt.Printf("\033[%dA", itemCount)
	fmt.Print("\033[?25h") // Show cursor
}
