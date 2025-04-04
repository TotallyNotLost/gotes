package tabs

import (
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "7"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

func New() Model {
	return Model{}
}

type Model struct {
	tabs []Tab
	activeTab int
	height int
	width int
}

func (m *Model) SetHeight(height int) {
	// Subtract to account for the tab view top
	m.height = height - 6
}

func (m *Model) SetWidth(width int) {
	// Subtract to account for the tab view sides
	m.width = width - 4
}

func (m Model) GetTabs() []Tab {
	return m.tabs
}

func (m *Model) SetTabs(tabs []Tab) {
	m.tabs = tabs
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "right", "l", "n", "tab":
			m.activeTab = min(m.activeTab+1, len(m.tabs)-1)
			return m, nil
		case "left", "h", "p", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, tab := range m.tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.tabs)-1, i == m.activeTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border).Width((m.width / len(m.tabs))-1)
		renderedTabs = append(renderedTabs, style.Render(tab.title))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(windowStyle.Width(m.width).Height(m.height).Render(m.tabs[m.activeTab].body))
	return docStyle.Render(doc.String())
}

func NewTab(title string, body string) Tab {
	return Tab{
		title: title,
		body: body,
	}
}

type Tab struct {
	title string
	body string
}

func (t Tab) GetBody() string {
	return t.body
}
