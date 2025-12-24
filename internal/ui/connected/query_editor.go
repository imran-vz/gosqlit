package connected

import (
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/imran-vz/gosqlit/internal/debug"
)

// QueryEditor is multi-line SQL editor
type QueryEditor struct {
	lines     []string
	cursorRow int
	cursorCol int
	scroll    int
	width     int
	height    int
}

// NewQueryEditor creates editor
func NewQueryEditor() *QueryEditor {
	return &QueryEditor{
		lines:     []string{""},
		cursorRow: 0,
		cursorCol: 0,
	}
}

// Update handles messages
func (qe *QueryEditor) Update(msg tea.Msg) (*QueryEditor, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		debug.Logf("QueryEditor received key: '%s' | cursor: (%d,%d) | lines: %d",
			keyMsg.String(), qe.cursorRow, qe.cursorCol, len(qe.lines))

		switch keyMsg.String() {
		case "up":
			if qe.cursorRow > 0 {
				qe.cursorRow--
				// Clamp cursor column
				if qe.cursorCol > len(qe.lines[qe.cursorRow]) {
					qe.cursorCol = len(qe.lines[qe.cursorRow])
				}
				debug.Logf("Cursor moved up to (%d,%d)", qe.cursorRow, qe.cursorCol)
			}
		case "down":
			if qe.cursorRow < len(qe.lines)-1 {
				qe.cursorRow++
				// Clamp cursor column
				if qe.cursorCol > len(qe.lines[qe.cursorRow]) {
					qe.cursorCol = len(qe.lines[qe.cursorRow])
				}
				debug.Logf("Cursor moved down to (%d,%d)", qe.cursorRow, qe.cursorCol)
			}
		case "left":
			if qe.cursorCol > 0 {
				qe.cursorCol--
				debug.Logf("Cursor moved left to (%d,%d)", qe.cursorRow, qe.cursorCol)
			} else if qe.cursorRow > 0 {
				qe.cursorRow--
				qe.cursorCol = len(qe.lines[qe.cursorRow])
				debug.Logf("Cursor moved to end of prev line (%d,%d)", qe.cursorRow, qe.cursorCol)
			}
		case "right":
			if qe.cursorCol < len(qe.lines[qe.cursorRow]) {
				qe.cursorCol++
				debug.Logf("Cursor moved right to (%d,%d)", qe.cursorRow, qe.cursorCol)
			} else if qe.cursorRow < len(qe.lines)-1 {
				qe.cursorRow++
				qe.cursorCol = 0
				debug.Logf("Cursor moved to start of next line (%d,%d)", qe.cursorRow, qe.cursorCol)
			}
		case "home":
			qe.cursorCol = 0
			debug.Logf("Cursor moved to home (%d,%d)", qe.cursorRow, qe.cursorCol)
		case "end":
			qe.cursorCol = len(qe.lines[qe.cursorRow])
			debug.Logf("Cursor moved to end (%d,%d)", qe.cursorRow, qe.cursorCol)
		case "enter":
			// Split line at cursor
			currentLine := qe.lines[qe.cursorRow]
			before := currentLine[:qe.cursorCol]
			after := currentLine[qe.cursorCol:]

			qe.lines[qe.cursorRow] = before
			qe.lines = append(qe.lines[:qe.cursorRow+1], append([]string{after}, qe.lines[qe.cursorRow+1:]...)...)

			qe.cursorRow++
			qe.cursorCol = 0
			debug.Logf("Line split at cursor | new total lines: %d | cursor: (%d,%d)",
				len(qe.lines), qe.cursorRow, qe.cursorCol)
		case "backspace":
			// Check for modifier key combinations
			if keyMsg.String() == "ctrl+backspace" || keyMsg.String() == "cmd+backspace" {
				// Clear current line from cursor to beginning or entire line
				currentLine := qe.lines[qe.cursorRow]
				after := currentLine[qe.cursorCol:]

				// If at start of line and Ctrl/Cmd+Backspace pressed, delete previous line
				if qe.cursorCol == 0 && qe.cursorRow > 0 {
					// Merge with previous line
					prevLine := qe.lines[qe.cursorRow-1]
					qe.lines[qe.cursorRow-1] = prevLine + after
					qe.lines = append(qe.lines[:qe.cursorRow], qe.lines[qe.cursorRow+1:]...)
					qe.cursorRow--
					qe.cursorCol = len(prevLine)
					debug.Logf("Line cleared and merged: new total lines: %d | cursor: (%d,%d)", len(qe.lines), qe.cursorRow, qe.cursorCol)
				} else {
					// Clear from beginning of line to cursor
					qe.lines[qe.cursorRow] = after
					qe.cursorCol = 0
					debug.Logf("Line cleared from start to cursor: line length: %d | cursor: (%d,%d)", len(qe.lines[qe.cursorRow]), qe.cursorRow, qe.cursorCol)
				}
			} else {
				// Regular backspace
				if qe.cursorCol > 0 {
					// Delete character before cursor
					line := qe.lines[qe.cursorRow]
					qe.lines[qe.cursorRow] = line[:qe.cursorCol-1] + line[qe.cursorCol:]
					qe.cursorCol--
					debug.Logf("Character deleted | new line length: %d | cursor: (%d,%d)",
						len(qe.lines[qe.cursorRow]), qe.cursorRow, qe.cursorCol)
				} else if qe.cursorRow > 0 {
					// Merge with previous line
					prevLine := qe.lines[qe.cursorRow-1]
					currentLine := qe.lines[qe.cursorRow]
					qe.lines[qe.cursorRow-1] = prevLine + currentLine
					qe.lines = append(qe.lines[:qe.cursorRow], qe.lines[qe.cursorRow+1:]...)
					qe.cursorRow--
					qe.cursorCol = len(prevLine)
					debug.Logf("Lines merged | new total lines: %d | cursor: (%d,%d) | merged line length: %d",
						len(qe.lines), qe.cursorRow, qe.cursorCol, len(prevLine+currentLine))
				}
			}
		case "alt+backspace", "ctrl+backspace":
			// Handle line clearing - Alt+Backspace on macOS, Ctrl+Backspace on Linux
			currentLine := qe.lines[qe.cursorRow]
			after := currentLine[qe.cursorCol:]

			if qe.cursorCol == 0 && qe.cursorRow > 0 {
				// If at start, merge with previous line
				prevLine := qe.lines[qe.cursorRow-1]
				qe.lines[qe.cursorRow-1] = prevLine + after
				qe.lines = append(qe.lines[:qe.cursorRow], qe.lines[qe.cursorRow+1:]...)
				qe.cursorRow--
				qe.cursorCol = len(prevLine)
				debug.Logf("Line clear: Lines merged | total: %d | cursor: (%d,%d)", len(qe.lines), qe.cursorRow, qe.cursorCol)
			} else {
				// Clear to beginning of current line
				qe.lines[qe.cursorRow] = after
				qe.cursorCol = 0
				debug.Logf("Line clear: Line cleared | line length: %d | cursor: (%d,%d)", len(qe.lines[qe.cursorRow]), qe.cursorRow, qe.cursorCol)
			}
		case "ctrl+k":
			// Ctrl+K - alternative to Ctrl+Backspace for clearing to end of line
			currentLine := qe.lines[qe.cursorRow]
			before := currentLine[:qe.cursorCol]

			// Clear from cursor to end of line
			qe.lines[qe.cursorRow] = before
			debug.Logf("Ctrl+K: Line cleared to end | line length: %d | cursor: (%d,%d)", len(qe.lines[qe.cursorRow]), qe.cursorRow, qe.cursorCol)
		case "ctrl+cmd+backspace", "ctrl+meta+backspace":
			// Handle Cmd+Backspace (may be detected differently on different systems)
			qe.lines[qe.cursorRow] = ""
			if len(qe.lines) == 1 {
				qe.cursorCol = 0
			} else if qe.cursorRow > 0 {
				// If not first line, merge with previous
				prevLine := qe.lines[qe.cursorRow-1]
				qe.lines = append(qe.lines[:qe.cursorRow], qe.lines[qe.cursorRow+1:]...)
				qe.cursorRow--
				qe.cursorCol = len(prevLine)
			} else {
				// If first line, delete it
				qe.lines = append(qe.lines[:qe.cursorRow], qe.lines[qe.cursorRow+1:]...)
				qe.cursorCol = 0
			}
			debug.Logf("Cmd+Backspace: Line fully cleared/merged | lines: %d | cursor: (%d,%d)", len(qe.lines), qe.cursorRow, qe.cursorCol)
		case "tab":
			// Let parent handle tab for pane switching
			debug.Logf("Tab received - delegating to parent for pane switching")
			return qe, nil
		case "alt+enter":
			debug.Logf("Alt+Enter received - delegating to parent for query execution")
			// Don't consume - simply return without handling
			return qe, nil
		case "ctrl+v", "ctrl+cmd+v", "ctrl+meta+v":
			// Handle paste functionality
			go func() {
				clipboard := getClipboardContent()
				debug.Logf(" attempting paste from clipboard length: %d", len(clipboard))
				if clipboard != "" {
					// Clean clipboard content - remove brackets and normalize
					cleaned := cleanClipboardContent(clipboard)
					debug.Logf("Cleaned clipboard content: %.100s", cleaned)

					// Insert the cleaned content at cursor position
					currentLine := qe.lines[qe.cursorRow]
					before := currentLine[:qe.cursorCol]
					after := currentLine[qe.cursorCol:]

					// Split newlines into multiple lines
					pasteLines := strings.Split(cleaned, "\n")
					if len(pasteLines) == 1 {
						// Single line paste
						qe.lines[qe.cursorRow] = before + pasteLines[0] + after
						qe.cursorCol += len(pasteLines[0])
						debug.Logf("Single line pasted | line length: %d | cursor: (%d,%d)", len(qe.lines[qe.cursorRow]), qe.cursorRow, qe.cursorCol)
					} else {
						// Multi-line paste - replace current line and add new lines
						qe.lines[qe.cursorRow] = before + pasteLines[0]
						newLines := append(pasteLines[1:], after)
						qe.lines = append(qe.lines[:qe.cursorRow+1], append(newLines, qe.lines[qe.cursorRow+1:]...)...)
						qe.cursorRow += len(pasteLines) - 1
						qe.cursorCol = len(pasteLines[len(pasteLines)-1])
						debug.Logf("Multi-line pasted | new total lines: %d | cursor: (%d,%d)", len(qe.lines), qe.cursorRow, qe.cursorCol)
					}
				}
			}()
		default:
			// Insert character
			if len(keyMsg.String()) == 1 {
				line := qe.lines[qe.cursorRow]
				qe.lines[qe.cursorRow] = line[:qe.cursorCol] + keyMsg.String() + line[qe.cursorCol:]
				qe.cursorCol++
				debug.Logf("Character inserted: '%s' | line length: %d | cursor: (%d,%d)",
					keyMsg.String(), len(qe.lines[qe.cursorRow]), qe.cursorRow, qe.cursorCol)
			}
		}

		// Adjust scroll
		if qe.cursorRow < qe.scroll {
			qe.scroll = qe.cursorRow
			debug.Logf("Scroll adjusted up: %d", qe.scroll)
		} else if qe.cursorRow >= qe.scroll+qe.height-4 {
			qe.scroll = qe.cursorRow - qe.height + 5
			debug.Logf("Scroll adjusted down: %d", qe.scroll)
		}
	}

	return qe, nil
}

// View renders the editor
func (qe *QueryEditor) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true).
		Padding(0, 1)

	lineNumStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(4).
		Align(lipgloss.Right)

	title := titleStyle.Render("SQL Editor")

	var lines []string

	// Render visible lines
	end := qe.scroll + qe.height - 4
	if end > len(qe.lines) {
		end = len(qe.lines)
	}

	for i := qe.scroll; i < end; i++ {
		lineNum := lineNumStyle.Render(string(rune(i + 1)))
		lineContent := qe.lines[i]

		// Show cursor
		if i == qe.cursorRow {
			// Insert cursor character
			if qe.cursorCol <= len(lineContent) {
				lineContent = lineContent[:qe.cursorCol] + "█" + lineContent[qe.cursorCol:]
			}
		}

		lines = append(lines, lineNum+" "+lineContent)
	}

	content := strings.Join(lines, "\n")

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("\nAlt+Enter: Execute  Tab: Switch pane  F5/Ctrl+R: Refresh schemas\n" +
			"Ctrl+K: Clear to end  Ctrl+V: Paste  Alt+Backspace: Clear line")

	return title + "\n\n" + content + helpText
}

// SetContent sets editor content
func (qe *QueryEditor) SetContent(content string) {
	qe.lines = strings.Split(content, "\n")
	if len(qe.lines) == 0 {
		qe.lines = []string{""}
	}
	qe.cursorRow = 0
	qe.cursorCol = 0
	qe.scroll = 0
}

// GetContent returns editor content
func (qe *QueryEditor) GetContent() string {
	return strings.Join(qe.lines, "\n")
}

// SetDimensions sets width and height
func (qe *QueryEditor) SetDimensions(width, height int) {
	qe.width = width
	qe.height = height
}

// getClipboardContent retrieves clipboard content based on the OS
func getClipboardContent() string {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("pbpaste")
	case "linux":
		// Try xclip first
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			// Fallback to xsel
			cmd = exec.Command("xsel", "--clipboard", "--output")
		} else {
			return ""
		}
	case "windows":
		cmd = exec.Command("powershell", "-command", "Get-Clipboard")
	default:
		return ""
	}

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return string(output)
}

// cleanClipboardContent removes brackets and cleans clipboard content
func cleanClipboardContent(content string) string {
	// Remove square brackets []
	content = regexp.MustCompile(`^\[|\]$`).ReplaceAllString(content, "")

	// Remove parentheses ()
	content = regexp.MustCompile(`^\(|\)$`).ReplaceAllString(content, "")

	// Remove curly braces {}
	content = regexp.MustCompile(`^\{|\}$`).ReplaceAllString(content, "")

	// Remove quotes around entire content
	content = regexp.MustCompile(`^['"\x60]|['"\x60]$`).ReplaceAllString(content, "")

	// Remove leading/trailing whitespace
	content = strings.TrimSpace(content)

	// Replace Windows line endings with Unix line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")

	return content
}
