package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mssantosdev/norn/internal/ui/styles"
)

// TreeNode represents a node in the interactive tree.
type TreeNode struct {
	ID       string
	Label    string
	Detail   string
	Status   string
	Expanded bool
	Children []*TreeNode
	Parent   *TreeNode
	Depth    int
}

// TreeModel is a Bubble Tea model for interactive tree navigation.
type TreeModel struct {
	root     *TreeNode
	cursor   int
	flat     []*TreeNode // flattened visible nodes
	width    int
	quitting bool
}

// NewTree creates a new interactive tree from a root node.
func NewTree(root *TreeNode) TreeModel {
	m := TreeModel{
		root:  root,
		width: styles.TerminalWidth() - 4,
	}
	m.rebuildFlat()
	return m
}

// rebuildFlat rebuilds the flattened visible node list.
func (m *TreeModel) rebuildFlat() {
	m.flat = make([]*TreeNode, 0)
	m.walk(m.root)
}

// walk recursively adds visible nodes to the flat list.
func (m *TreeModel) walk(node *TreeNode) {
	m.flat = append(m.flat, node)
	if node.Expanded {
		for _, child := range node.Children {
			m.walk(child)
		}
	}
}

// Init implements tea.Model.
func (m TreeModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m TreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap().Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, DefaultKeyMap().Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, DefaultKeyMap().Down):
			if m.cursor < len(m.flat)-1 {
				m.cursor++
			}
		case key.Matches(msg, DefaultKeyMap().Select):
			if m.cursor < len(m.flat) {
				node := m.flat[m.cursor]
				if len(node.Children) > 0 {
					node.Expanded = !node.Expanded
					m.rebuildFlat()
				}
			}
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m TreeModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	for i, node := range m.flat {
		line := m.renderNode(node, i == m.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

// renderNode renders a single tree node line.
func (m TreeModel) renderNode(node *TreeNode, selected bool) string {
	indent := strings.Repeat("  ", node.Depth)

	// Expand/collapse indicator
	indicator := "  "
	if len(node.Children) > 0 {
		if node.Expanded {
			indicator = styles.TreeBranch.Render("▼ ")
		} else {
			indicator = styles.TreeBranch.Render("▶ ")
		}
	} else {
		indicator = styles.TreeGuide.Render("  ")
	}

	// Tree guide for depth
	guide := ""
	if node.Depth > 0 {
		if node == node.Parent.Children[len(node.Parent.Children)-1] {
			guide = styles.TreeGuide.Render("└──")
		} else {
			guide = styles.TreeGuide.Render("├──")
		}
	}

	// Label
	label := styles.TreeLeaf.Render(node.Label)
	if selected {
		label = styles.SelectionBadge.Render(" "+node.Label+" ") + ""
	}

	// Detail
	detail := ""
	if node.Detail != "" {
		detail = " " + styles.Dimmed.Render(node.Detail)
	}

	// Status badge
	status := ""
	if node.Status != "" {
		status = " " + styles.StatusBadgeFor(node.Status)
	}

	return indent + guide + indicator + label + detail + status
}

// Selected returns the currently selected node.
func (m TreeModel) Selected() *TreeNode {
	if m.cursor < len(m.flat) {
		return m.flat[m.cursor]
	}
	return nil
}

// Run runs the interactive tree and returns the selected node.
func (m TreeModel) Run() (*TreeNode, error) {
	p := tea.NewProgram(m)
	model, err := p.Run()
	if err != nil {
		return nil, err
	}
	tm := model.(TreeModel)
	return tm.Selected(), nil
}

// TreeFromItems creates a simple tree from parent-child relationships.
func TreeFromItems(items []TreeItem) *TreeNode {
	root := &TreeNode{
		ID:    "root",
		Label: "root",
		Depth: -1,
	}
	nodeMap := make(map[string]*TreeNode)
	nodeMap["root"] = root

	for _, item := range items {
		node := &TreeNode{
			ID:     item.ID,
			Label:  item.Label,
			Detail: item.Detail,
			Status: item.Status,
			Depth:  item.Depth,
		}
		nodeMap[item.ID] = node
	}

	for _, item := range items {
		node := nodeMap[item.ID]
		parent := root
		if item.ParentID != "" {
			if p, ok := nodeMap[item.ParentID]; ok {
				parent = p
			}
		}
		node.Parent = parent
		parent.Children = append(parent.Children, node)
	}

	return root
}

// TreeItem represents a flat item to build a tree from.
type TreeItem struct {
	ID       string
	ParentID string
	Label    string
	Detail   string
	Status   string
	Depth    int
}

// StaticTree renders a tree as a static string (non-interactive).
func StaticTree(root *TreeNode) string {
	var b strings.Builder
	renderStaticNode(&b, root, "")
	return b.String()
}

func renderStaticNode(b *strings.Builder, node *TreeNode, prefix string) {
	if node.Depth >= 0 {
		line := prefix

		// Expand indicator
		if len(node.Children) > 0 {
			if node.Expanded {
				line += styles.TreeBranch.Render("▼ ")
			} else {
				line += styles.TreeBranch.Render("▶ ")
			}
		} else {
			line += styles.TreeGuide.Render("  ")
		}

		line += styles.TreeLeaf.Render(node.Label)

		if node.Detail != "" {
			line += " " + styles.Dimmed.Render(node.Detail)
		}
		if node.Status != "" {
			line += " " + styles.StatusBadgeFor(node.Status)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	if node.Expanded {
		childPrefix := prefix + styles.TreeGuide.Render("│   ")
		for i, child := range node.Children {
			if i == len(node.Children)-1 {
				childPrefix = prefix + styles.TreeGuide.Render("    ")
			}
			renderStaticNode(b, child, childPrefix)
		}
	}
}

// RunStaticTree runs a static tree view (always expanded, no interaction).
func RunStaticTree(root *TreeNode) string {
	root.Expanded = true
	for _, child := range root.Children {
		child.Expanded = true
	}
	return StaticTree(root)
}
