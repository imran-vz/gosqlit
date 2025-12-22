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
	columns     []string
	rows        [][]interface{}
	fetchOffset int
	hasMore     bool
	page        int
	pageSize    int
	cursor      int
	scroll      int
	width       int
	height      int
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
			if rt.cursor >= len(rt.rows) {
				rt.cursor = len(rt.rows) - 1
			}
		}

		// Adjust scroll
		if rt.cursor < rt.scroll {
			rt.scroll = rt.cursor
		}
		if rt.cursor >= rt.scroll+rt.height-6 {
			rt.scroll = rt.cursor - rt.height + 7
		}
	}

	return rt, nil
}

// View renders the table
func (rt *ResultsTable) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true).
		Padding(0, 1)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 1)

	title := titleStyle.Render(fmt.Sprintf("Results (%d rows)", len(rt.rows)))

	if len(rt.columns) == 0 {
		noData := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No query executed yet")

		return borderStyle.Width(rt.width - 2).Height(rt.height - 2).Render(title + "\n\n" + noData)
	}

	// Calculate column widths
	colWidths := make([]int, len(rt.columns))
	for i, col := range rt.columns {
		colWidths[i] = len(col)
	}

	// Check row widths
	for _, row := range rt.rows {
		for i, cell := range row {
			cellStr := fmt.Sprintf("%v", cell)
			if len(cellStr) > colWidths[i] {
				colWidths[i] = len(cellStr)
			}
			// Max width 30
			if colWidths[i] > 30 {
				colWidths[i] = 30
			}
		}
	}

	// Render header
	headerParts := []string{}
	for i, col := range rt.columns {
		headerParts = append(headerParts, padString(col, colWidths[i]))
	}
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true).
		Render(strings.Join(headerParts, " │ "))

	separator := strings.Repeat("─", rt.width-6)

	// Render visible rows
	var rowStrs []string
	end := rt.scroll + rt.height - 8
	if end > len(rt.rows) {
		end = len(rt.rows)
	}

	for i := rt.scroll; i < end; i++ {
		row := rt.rows[i]
		rowParts := []string{}

		for j, cell := range row {
			cellStr := fmt.Sprintf("%v", cell)
			if len(cellStr) > 30 {
				cellStr = cellStr[:27] + "..."
			}
			rowParts = append(rowParts, padString(cellStr, colWidths[j]))
		}

		rowStr := strings.Join(rowParts, " │ ")

		if i == rt.cursor {
			rowStr = lipgloss.NewStyle().
				Background(lipgloss.Color("237")).
				Render(rowStr)
		}

		rowStrs = append(rowStrs, rowStr)
	}

	content := header + "\n" + separator + "\n" + strings.Join(rowStrs, "\n")

	// Footer with info
	footer := ""
	if rt.hasMore {
		footer = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(fmt.Sprintf("\nShowing %d-%d of %d+ rows (Ctrl+L to load more)", rt.scroll+1, end, len(rt.rows)))
	} else {
		footer = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(fmt.Sprintf("\nShowing %d-%d of %d rows", rt.scroll+1, end, len(rt.rows)))
	}

	box := borderStyle.Width(rt.width - 2).Height(rt.height - 2).Render(title + "\n\n" + content + footer)

	return box
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

// padString pads string to width
func padString(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
