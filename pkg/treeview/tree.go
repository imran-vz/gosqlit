package treeview

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Node represents a tree node
type Node struct {
	ID       string
	Label    string
	Children []*Node
	Expanded bool
	Data     interface{} // Custom data
}

// Tree is a reusable tree component
type Tree struct {
	Root     *Node
	cursor   int     // Current cursor position in flat list
	flatList []*Node // Flattened view of visible nodes
	selected *Node   // Currently selected node
	scroll   int     // Scroll offset
}

// NewTree creates a new tree
func NewTree(root *Node) *Tree {
	t := &Tree{
		Root: root,
	}
	t.rebuild()
	return t
}

// rebuild flattens the tree into visible nodes
func (t *Tree) rebuild() {
	t.flatList = []*Node{}
	t.flattenNode(t.Root, 0)

	// Ensure cursor is valid
	if t.cursor >= len(t.flatList) {
		t.cursor = len(t.flatList) - 1
	}
	if t.cursor < 0 {
		t.cursor = 0
	}

	// Update selected node
	if t.cursor >= 0 && t.cursor < len(t.flatList) {
		t.selected = t.flatList[t.cursor]
	}
}

// flattenNode recursively flattens tree
func (t *Tree) flattenNode(node *Node, depth int) {
	if node == nil {
		return
	}

	t.flatList = append(t.flatList, node)

	if node.Expanded {
		for _, child := range node.Children {
			t.flattenNode(child, depth+1)
		}
	}
}

// Update handles messages
func (t *Tree) Update(msg tea.Msg) (*Tree, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "up", "k":
			if t.cursor > 0 {
				t.cursor--
				t.selected = t.flatList[t.cursor]
			}
		case "down", "j":
			if t.cursor < len(t.flatList)-1 {
				t.cursor++
				t.selected = t.flatList[t.cursor]
			}
		case "right", "l":
			// Expand node
			if t.selected != nil && len(t.selected.Children) > 0 {
				t.selected.Expanded = true
				t.rebuild()
			}
		case "left", "h":
			// Collapse node
			if t.selected != nil {
				t.selected.Expanded = false
				t.rebuild()
			}
		case "enter", " ":
			// Toggle expand
			if t.selected != nil && len(t.selected.Children) > 0 {
				t.selected.Expanded = !t.selected.Expanded
				t.rebuild()
			}
		}
	}

	return t, nil
}

// View renders the tree
func (t *Tree) View(width, height int) string {
	if len(t.flatList) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No items")
	}

	var lines []string

	// Adjust scroll
	if t.cursor < t.scroll {
		t.scroll = t.cursor
	}
	if t.cursor >= t.scroll+height {
		t.scroll = t.cursor - height + 1
	}

	// Render visible nodes
	end := t.scroll + height
	if end > len(t.flatList) {
		end = len(t.flatList)
	}

	for i := t.scroll; i < end; i++ {
		node := t.flatList[i]
		depth := t.getDepth(node)

		// Indent
		indent := strings.Repeat("  ", depth)

		// Expand indicator
		indicator := "  "
		if len(node.Children) > 0 {
			if node.Expanded {
				indicator = "▼ "
			} else {
				indicator = "▶ "
			}
		}

		line := indent + indicator + node.Label

		// Highlight selected
		if i == t.cursor {
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("63")).
				Background(lipgloss.Color("237")).
				Bold(true).
				Render(line)
		} else {
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Render(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// getDepth calculates node depth
func (t *Tree) getDepth(target *Node) int {
	return t.getDepthRecursive(t.Root, target, 0)
}

// getDepthRecursive helper
func (t *Tree) getDepthRecursive(node *Node, target *Node, depth int) int {
	if node == target {
		return depth
	}

	for _, child := range node.Children {
		if d := t.getDepthRecursive(child, target, depth+1); d >= 0 {
			return d
		}
	}

	return -1
}

// GetSelected returns selected node
func (t *Tree) GetSelected() *Node {
	return t.selected
}

// SetRoot sets new root and rebuilds
func (t *Tree) SetRoot(root *Node) {
	t.Root = root
	t.cursor = 0
	t.scroll = 0
	t.rebuild()
}
