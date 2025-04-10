package tabs

import (
	"fmt"
	"github.com/TotallyNotLost/gotes/formatter"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var (
	docStyle      = lipgloss.NewStyle()
	viewportStyle = docStyle.
			BorderForeground(highlightColor).
			Border(lipgloss.NormalBorder())
	tabsStyle        = lipgloss.NewStyle()
	highlightColor   = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "7"}
	inactiveTabStyle = lipgloss.NewStyle().
				BorderForeground(lipgloss.Color("241")).
				Foreground(lipgloss.Color("241")).
				Border(lipgloss.NormalBorder())
	activeTabStyle = inactiveTabStyle.
			BorderForeground(lipgloss.Color("7")).
			Foreground(lipgloss.Color("7"))
	statusBarStyle = docStyle.
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})
	statusStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#6124DF")).
			Padding(0, 1).
			MarginRight(1)
	statusTextStyle = lipgloss.NewStyle().Inherit(statusBarStyle)
	scrollStyle     = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1).
			Background(lipgloss.Color("#A550DF")).
			Align(lipgloss.Right)
)

func New() Model {
	vp := viewport.New(0, 0)
	vp.Style = viewportStyle
	return Model{
		viewport:  vp,
		formatter: formatter.Default,
		keyMap:    defaultKeyMap(),
	}
}

type Model struct {
	viewport  viewport.Model
	entryId   string
	tabs      []Tab
	ActiveTab int
	formatter formatter.Formatter
	height    int
	width     int
	keyMap    keyMap
}

func (m *Model) SetEntryId(entryId string) {
	m.entryId = entryId
}

func (m *Model) SetFormatter(formatter formatter.Formatter) {
	m.formatter = formatter
}

func (m *Model) SetHeight(height int) {
	m.height = height
	m.AdjustHeight()
}

func (m *Model) AdjustHeight() {
	verticalHeight := lipgloss.Height(m.footerView())
	if len(m.tabs) != 1 {
		verticalHeight += lipgloss.Height(m.tabsView())
	}
	m.viewport.Height = m.height - verticalHeight
}

func (m *Model) SetWidth(width int) {
	m.width = width
	m.viewport.Width = width
}

func (m Model) GetTabs() []Tab {
	return m.tabs
}

func (m *Model) SetTabs(tabs []Tab) {
	m.ActiveTab = 0
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
			m.ActiveTab = max(m.ActiveTab-1, 0)
			return m, nil
		case key.Matches(msg, m.keyMap.Next):
			m.ActiveTab = min(m.ActiveTab+1, len(m.tabs)-1)
			return m, nil
		}
	}

	var cmd tea.Cmd

	m.viewport.SetContent(m.content())
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	m.viewport.SetContent(m.content())
	if len(m.tabs) == 1 {
		return lipgloss.JoinVertical(lipgloss.Left, m.viewport.View(), m.footerView())
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.tabsView(), m.viewport.View(), m.footerView())
}

func (m Model) tabsView() string {
	doc := strings.Builder{}

	var renderedTabs []string
	var totalWidth int

	for i, tab := range m.tabs {
		var style lipgloss.Style
		isActive := i == m.ActiveTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		tb := style.Render(tab.title)
		totalWidth += lipgloss.Width(tb)
		if totalWidth >= m.width {
			row := lipgloss.JoinHorizontal(lipgloss.Bottom, renderedTabs...)
			doc.WriteString(row)
			doc.WriteString("\n")
			renderedTabs = []string{}
			totalWidth = lipgloss.Width(tb)
		}
		renderedTabs = append(renderedTabs, tb)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Bottom, renderedTabs...)
	doc.WriteString(row)
	return tabsStyle.Render(doc.String())
}

func (m Model) content() string {
	return docStyle.Render(m.body())
}

func (m Model) body() string {
	if len(m.tabs) <= m.ActiveTab {
		return ""
	}

	return m.formatter.Format(m.tabs[m.ActiveTab].body)
}

func (m Model) footerView() string {
	scrollPercent := fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)

	w := lipgloss.Width

	format := "Markdown"
	if m.formatter == formatter.Default {
		format = "Raw"
		statusStyle = statusStyle.
			Background(lipgloss.Color("#FF5F87"))
	}
	statusKey := statusStyle.Render(format)
	scroll := scrollStyle.Render(scrollPercent)
	statusVal := statusTextStyle.
		Width(m.width - w(statusKey) - w(scroll)).
		Render(m.entryId)

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		statusKey,
		statusVal,
		scroll,
	)

	return statusBarStyle.Width(m.width).Render(bar)
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
