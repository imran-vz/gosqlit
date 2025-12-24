package connected

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/imran-vz/gosqlit/internal/db"
	"github.com/imran-vz/gosqlit/pkg/treeview"
)

// SchemaBrowser displays database schema tree
type SchemaBrowser struct {
	tree    *treeview.Tree
	loading bool
	width   int
	height  int
}

// NewSchemaBrowser creates schema browser
func NewSchemaBrowser() *SchemaBrowser {
	root := &treeview.Node{
		ID:       "root",
		Label:    "Loading...",
		Children: []*treeview.Node{},
		Expanded: true,
	}

	return &SchemaBrowser{
		tree:    treeview.NewTree(root),
		loading: true,
	}
}

// Update handles messages
func (sb *SchemaBrowser) Update(msg tea.Msg) (*SchemaBrowser, tea.Cmd) {
	newTree, cmd := sb.tree.Update(msg)
	sb.tree = newTree
	return sb, cmd
}

// View renders the browser
func (sb *SchemaBrowser) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true).
		Padding(0, 1)

	title := titleStyle.Render("Schema")

	content := ""
	if sb.loading {
		content = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Loading schemas...")
	} else {
		content = sb.tree.View(sb.width-4, sb.height-6)
	}

	return title + "\n\n" + content
}

// SetSchemas populates tree with schemas
func (sb *SchemaBrowser) SetSchemas(schemas []db.Schema) {
	root := &treeview.Node{
		ID:       "root",
		Label:    "Schemas",
		Children: []*treeview.Node{},
		Expanded: true,
	}

	for _, schema := range schemas {
		schemaNode := &treeview.Node{
			ID:       "schema:" + schema.Name,
			Label:    "📂 " + schema.Name,
			Children: []*treeview.Node{},
			Expanded: false,
			Data:     schema,
		}

		for _, table := range schema.Tables {
			tableNode := &treeview.Node{
				ID:       "table:" + schema.Name + "." + table.Name,
				Label:    "📄 " + table.Name,
				Children: []*treeview.Node{},
				Data:     table,
			}
			schemaNode.Children = append(schemaNode.Children, tableNode)
		}

		root.Children = append(root.Children, schemaNode)
	}

	sb.tree.SetRoot(root)
	sb.loading = false
}

// GetSelectedTable returns selected table if any
func (sb *SchemaBrowser) GetSelectedTable() (schema, table string, ok bool) {
	selected := sb.tree.GetSelected()
	if selected == nil {
		return "", "", false
	}

	if tbl, ok := selected.Data.(db.Table); ok {
		return tbl.Schema, tbl.Name, true
	}

	return "", "", false
}

// SetDimensions sets width and height
func (sb *SchemaBrowser) SetDimensions(width, height int) {
	sb.width = width
	sb.height = height
}
