package tabs

import (
	"github.com/TotallyNotLost/gotes/formatter"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
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
	docStyle          = lipgloss.NewStyle().Padding(0, 1, 0, 1)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "7"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

func New() Model {
	return Model{
		formatter: formatter.Default,
		keyMap:    defaultKeyMap(),
	}
}

type Model struct {
	tabs      []Tab
	activeTab int
	formatter formatter.Formatter
	height    int
	width     int
	keyMap    keyMap
}

func (m *Model) SetFormatter(formatter formatter.Formatter) {
	m.formatter = formatter
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
		switch {
		case key.Matches(msg, m.keyMap.Previous):
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		case key.Matches(msg, m.keyMap.Next):
			m.activeTab = min(m.activeTab+1, len(m.tabs)-1)
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
		style = style.Border(border).Width((m.width / len(m.tabs)) - 1)
		renderedTabs = append(renderedTabs, style.Render(tab.title))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(windowStyle.Width(m.width).Height(m.height).Render(m.body()))
	return docStyle.Render(doc.String())
}

func (m Model) body() string {
	if len(m.tabs) < m.activeTab {
		return ""
	}

	return m.formatter.Format(m.tabs[m.activeTab].body)
}

func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{
		m.keyMap.Previous,
		m.keyMap.Next,
	}
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.ShortHelp()}
}

func NewTab(title string, body string) Tab {
	return Tab{
		title: title,
		body:  body,
	}
}

type Tab struct {
	title string
	body  string
}

func (t Tab) GetBody() string {
	return t.body
}

type keyMap struct {
	Previous key.Binding
	Next     key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Previous: key.NewBinding(key.WithKeys("left", "h", "p", "shift+tab"), key.WithHelp("←/h", "previous")),
		Next:     key.NewBinding(key.WithKeys("right", "l", "n", "tab"), key.WithHelp("→/l", "next")),
	}
}

type defaultFormatter struct {
}

func (f defaultFormatter) Format(s string) string {
	return s
}

var DefaultFormatter = defaultFormatter{}
