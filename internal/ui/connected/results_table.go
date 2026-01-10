package connected

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/imran-vz/gosqlit/internal/db"
)

// ResultsTable displays query results
type ResultsTable struct {
	columns  []string
	rows     [][]any
	hasMore  bool
	page     int
	pageSize int
	cursor   int
	scroll   int
	width    int
	height   int
}

// NewResultsTable creates results table
func NewResultsTable() *ResultsTable {
	return &ResultsTable{
		pageSize: 50,
		page:     0,
	}
}

// Update handles messages
func (rt *ResultsTable) Update(msg tea.Msg) (*ResultsTable, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "up", "k":
			if rt.cursor > 0 {
				rt.cursor--
			}
		case "down", "j":
			if rt.cursor < len(rt.rows)-1 {
				rt.cursor++
			}
		case "pageup":
			rt.cursor -= 10
			if rt.cursor < 0 {
				rt.cursor = 0
			}
		case "pagedown":
			rt.cursor += 10
			maxCursor := max(len(rt.rows)-1, 0)
			if rt.cursor > maxCursor {
				rt.cursor = maxCursor
			}
		case "home":
			rt.cursor = 0
		case "end":
			rt.cursor = max(len(rt.rows)-1, 0)
		}

		// Adjust scroll
		if rt.cursor < rt.scroll {
			rt.scroll = rt.cursor
		}
		// title + header + separator + footer + padding
		visibleRows := max(rt.height-7, 1)
		if rt.cursor >= rt.scroll+visibleRows {
			rt.scroll = rt.cursor - visibleRows + 1
		}
	}

	return rt, nil
}

// View renders the table
func (rt *ResultsTable) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true).
		PaddingLeft(1)

	title := titleStyle.Render(fmt.Sprintf("Results (%d rows)", len(rt.rows)))

	if len(rt.columns) == 0 {
		noData := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			PaddingLeft(1).
			Render("No query executed yet")

		return title + "\n\n" + noData
	}

	// Available width for the table content (accounting for borders/padding)
	availableWidth := max(rt.width-2, 20)

	// Calculate column widths
	colWidths := rt.calculateColumnWidths(availableWidth)

	// Build the table
	var output strings.Builder

	// Render header row
	headerRow := rt.renderRow(rt.columns, colWidths, true, false)
	output.WriteString(headerRow)
	output.WriteString("\n")

	// Render separator
	separator := rt.renderSeparator(colWidths)
	output.WriteString(separator)
	output.WriteString("\n")

	// Calculate visible rows
	visibleRows := max(rt.height-7, 1)

	end := rt.scroll + visibleRows
	if end > len(rt.rows) {
		end = len(rt.rows)
	}

	// Render visible data rows
	for i := rt.scroll; i < end; i++ {
		row := rt.rows[i]
		rowStrings := make([]string, len(row))
		for j, cell := range row {
			rowStrings[j] = sanitizeCellContent(fmt.Sprintf("%v", cell))
		}
		isSelected := i == rt.cursor
		rowLine := rt.renderRow(rowStrings, colWidths, false, isSelected)
		output.WriteString(rowLine)
		output.WriteString("\n")
	}

	// Footer with info
	var footer string
	if len(rt.rows) == 0 {
		footer = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			PaddingLeft(1).
			Render("0 rows")
	} else if rt.hasMore {
		footer = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			PaddingLeft(1).
			Render(fmt.Sprintf("Showing %d-%d of %d+ rows (Ctrl+L to load more)", rt.scroll+1, end, len(rt.rows)))
	} else {
		footer = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			PaddingLeft(1).
			Render(fmt.Sprintf("Showing %d-%d of %d rows", rt.scroll+1, end, len(rt.rows)))
	}

	return title + "\n\n" + output.String() + footer
}

// calculateColumnWidths calculates optimal column widths based on content and available space
func (rt *ResultsTable) calculateColumnWidths(availableWidth int) []int {
	numCols := len(rt.columns)
	if numCols == 0 {
		return nil
	}

	// Minimum separator width between columns: " │ " = 3 chars
	separatorWidth := (numCols - 1) * 3
	contentWidth := max(availableWidth-separatorWidth, numCols)

	// Calculate natural widths (based on content)
	naturalWidths := make([]int, numCols)
	minWidths := make([]int, numCols) // minimum width per column

	// Start with header widths
	for i, col := range rt.columns {
		naturalWidths[i] = runeWidth(col)
		minWidths[i] = min(3, runeWidth(col)) // minimum 3 chars or header length
	}

	// Check row widths (sample first 100 rows for performance)
	sampleRows := rt.rows
	if len(sampleRows) > 100 {
		sampleRows = sampleRows[:100]
	}

	for _, row := range sampleRows {
		for i, cell := range row {
			if i >= numCols {
				break
			}
			cellStr := sanitizeCellContent(fmt.Sprintf("%v", cell))
			w := runeWidth(cellStr)
			if w > naturalWidths[i] {
				naturalWidths[i] = w
			}
		}
	}

	// Cap natural widths at a reasonable max
	maxColWidth := 40
	for i := range naturalWidths {
		if naturalWidths[i] > maxColWidth {
			naturalWidths[i] = maxColWidth
		}
	}

	// Calculate total natural width
	totalNatural := 0
	for _, w := range naturalWidths {
		totalNatural += w
	}

	// Distribute widths
	finalWidths := make([]int, numCols)

	if totalNatural <= contentWidth {
		// All columns fit naturally
		copy(finalWidths, naturalWidths)
	} else {
		// Need to shrink columns proportionally
		for i := range finalWidths {
			ratio := float64(naturalWidths[i]) / float64(totalNatural)
			finalWidths[i] = max(int(ratio*float64(contentWidth)), minWidths[i])
		}

		// Redistribute any remaining space
		totalAssigned := 0
		for _, w := range finalWidths {
			totalAssigned += w
		}
		if remaining := contentWidth - totalAssigned; remaining > 0 {
			for i := range finalWidths {
				finalWidths[i]++
				remaining--
				if remaining <= 0 {
					break
				}
			}
		}
	}

	return finalWidths
}

// renderRow renders a single row with proper column alignment
func (rt *ResultsTable) renderRow(cells []string, colWidths []int, isHeader, isSelected bool) string {
	parts := make([]string, len(colWidths))

	for i := range colWidths {
		var cellContent string
		if i < len(cells) {
			cellContent = cells[i]
		}

		// Truncate if necessary
		if runeWidth(cellContent) > colWidths[i] {
			cellContent = truncateString(cellContent, colWidths[i])
		}

		// Pad to width
		cellContent = padToWidth(cellContent, colWidths[i])
		parts[i] = cellContent
	}

	row := strings.Join(parts, " │ ")

	if isHeader {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("63")).
			Bold(true).
			PaddingLeft(1).
			Render(row)
	}

	if isSelected {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			PaddingLeft(1).
			Render(row)
	}

	return lipgloss.NewStyle().
		PaddingLeft(1).
		Render(row)
}

// renderSeparator renders the separator line between header and data
func (rt *ResultsTable) renderSeparator(colWidths []int) string {
	parts := make([]string, len(colWidths))

	for i, width := range colWidths {
		parts[i] = strings.Repeat("─", width)
	}

	separator := strings.Join(parts, "─┼─")

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		PaddingLeft(1).
		Render(separator)
}

// SetData sets table data
func (rt *ResultsTable) SetData(result db.QueryResult) {
	rt.columns = result.Columns
	rt.rows = result.Rows
	rt.hasMore = result.HasMore
	rt.cursor = 0
	rt.scroll = 0
}

// AppendData appends more rows (for load more)
func (rt *ResultsTable) AppendData(result db.QueryResult) {
	rt.rows = append(rt.rows, result.Rows...)
	rt.hasMore = result.HasMore
}

// SetDimensions sets width and height
func (rt *ResultsTable) SetDimensions(width, height int) {
	rt.width = width
	rt.height = height
}

// runeWidth returns the display width of a string (counting runes)
func runeWidth(s string) int {
	return len([]rune(s))
}

// truncateString truncates a string to fit within maxWidth, adding "…" if truncated
func truncateString(s string, maxWidth int) string {
	runes := []rune(s)
	if len(runes) <= maxWidth {
		return s
	}
	if maxWidth <= 1 {
		return "…"
	}
	return string(runes[:maxWidth-1]) + "…"
}

// padToWidth pads a string with spaces to reach the target width
func padToWidth(s string, width int) string {
	currentWidth := runeWidth(s)
	if currentWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-currentWidth)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// sanitizeCellContent replaces newlines and control characters with visible representations
func sanitizeCellContent(s string) string {
	var result strings.Builder
	for _, r := range s {
		switch r {
		case '\n':
			result.WriteString(`\n`)
		case '\r':
			result.WriteString(`\r`)
		case '\t':
			result.WriteString(`\t`)
		case '\\':
			result.WriteString(`\\`)
		default:
			if r < 32 {
				fmt.Fprintf(&result, "\\x%02x", r)
			} else {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}
