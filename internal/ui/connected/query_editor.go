package connected

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
		switch keyMsg.String() {
		case "up":
			if qe.cursorRow > 0 {
				qe.cursorRow--
				// Clamp cursor column
				if qe.cursorCol > len(qe.lines[qe.cursorRow]) {
					qe.cursorCol = len(qe.lines[qe.cursorRow])
				}
			}
		case "down":
			if qe.cursorRow < len(qe.lines)-1 {
				qe.cursorRow++
				// Clamp cursor column
				if qe.cursorCol > len(qe.lines[qe.cursorRow]) {
					qe.cursorCol = len(qe.lines[qe.cursorRow])
				}
			}
		case "left":
			if qe.cursorCol > 0 {
				qe.cursorCol--
			} else if qe.cursorRow > 0 {
				qe.cursorRow--
				qe.cursorCol = len(qe.lines[qe.cursorRow])
			}
		case "right":
			if qe.cursorCol < len(qe.lines[qe.cursorRow]) {
				qe.cursorCol++
			} else if qe.cursorRow < len(qe.lines)-1 {
				qe.cursorRow++
				qe.cursorCol = 0
			}
		case "home":
			qe.cursorCol = 0
		case "end":
			qe.cursorCol = len(qe.lines[qe.cursorRow])
		case "enter":
			// Split line at cursor
			currentLine := qe.lines[qe.cursorRow]
			before := currentLine[:qe.cursorCol]
			after := currentLine[qe.cursorCol:]

			qe.lines[qe.cursorRow] = before
			qe.lines = append(qe.lines[:qe.cursorRow+1], append([]string{after}, qe.lines[qe.cursorRow+1:]...)...)

			qe.cursorRow++
			qe.cursorCol = 0
		case "backspace":
			if qe.cursorCol > 0 {
				// Delete character before cursor
				line := qe.lines[qe.cursorRow]
				qe.lines[qe.cursorRow] = line[:qe.cursorCol-1] + line[qe.cursorCol:]
				qe.cursorCol--
			} else if qe.cursorRow > 0 {
				// Merge with previous line
				prevLine := qe.lines[qe.cursorRow-1]
				currentLine := qe.lines[qe.cursorRow]
				qe.lines[qe.cursorRow-1] = prevLine + currentLine
				qe.lines = append(qe.lines[:qe.cursorRow], qe.lines[qe.cursorRow+1:]...)
				qe.cursorRow--
				qe.cursorCol = len(prevLine)
			}
		default:
			// Insert character
			if len(keyMsg.String()) == 1 {
				line := qe.lines[qe.cursorRow]
				qe.lines[qe.cursorRow] = line[:qe.cursorCol] + keyMsg.String() + line[qe.cursorCol:]
				qe.cursorCol++
			}
		}

		// Adjust scroll
		if qe.cursorRow < qe.scroll {
			qe.scroll = qe.cursorRow
		}
		if qe.cursorRow >= qe.scroll+qe.height-4 {
			qe.scroll = qe.cursorRow - qe.height + 5
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

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 1)

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
		Render("\nCtrl+Enter: Execute  Tab: Switch pane")

	box := borderStyle.Width(qe.width - 2).Height(qe.height - 2).Render(title + "\n\n" + content + helpText)

	return box
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
