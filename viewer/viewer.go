package viewer

import (
	"github.com/TotallyNotLost/gotes/cmd"
	"github.com/TotallyNotLost/gotes/formatter"
	"github.com/TotallyNotLost/gotes/markdown"
	"github.com/TotallyNotLost/gotes/tabs"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"strconv"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2).Render

func New(file string) Model {
	mdFormatter := formatter.NewMarkdown(file)
	tbs := tabs.New()
	tbs.SetFormatter(mdFormatter)
	return Model{
		tabs:              tbs,
		renderMarkdown:    true,
		markdownFormatter: mdFormatter,
		keyMap:            defaultKeyMap(),
		help:              help.New(),
	}
}

type Model struct {
	tabs              tabs.Model
	revisions         []markdown.Entry
	activeRevision    int
	renderMarkdown    bool
	markdownFormatter formatter.Formatter
	keyMap            keyMap
	help              help.Model
}

func (m *Model) SetHeight(height int) {
	// Subtract the height of helpView()
	// and an additional 2 to account for the title at the top
	m.tabs.SetHeight(height - lipgloss.Height(m.helpView()) - 2)
}

func (m *Model) SetWidth(width int) {
	m.tabs.SetWidth(width)
}

func (m *Model) SetRevisions(revisions []markdown.Entry) {
	m.revisions = revisions
	tabs := lo.Map(m.revisions, func(revision markdown.Entry, i int) tabs.Tab {
		title := "Revision HEAD~" + strconv.Itoa(i)
		if i == 0 {
			title += " (latest)"
		}
		return tabs.NewTab(title, revision.Body())
	})
	m.tabs.SetTabs(tabs)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	keyMap := defaultKeyMap()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, keyMap.Back) {
			return m, cmd.Back
		}
		if key.Matches(msg, keyMap.ToggleMarkdown) {
			m.renderMarkdown = !m.renderMarkdown
			if m.renderMarkdown {
				m.tabs.SetFormatter(m.markdownFormatter)
			} else {
				m.tabs.SetFormatter(tabs.DefaultFormatter)
			}

		}
	}

	var cmd tea.Cmd
	m.tabs, cmd = m.tabs.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	title := renderTitle(lo.LastOrEmpty(m.revisions))
	var body string

	// if len(m.revisions) == 1 {
		// body = m.tabs.GetTabs()[0].GetBody()
	// } else {
		body = m.tabs.View()
	// }

	return lipgloss.JoinVertical(lipgloss.Left, title, body, m.helpView())
}

func (m Model) ShortHelp() []key.Binding {
	bindings := []key.Binding{
		m.keyMap.Back,
		m.keyMap.ToggleMarkdown,
	}

	return append(bindings, m.tabs.ShortHelp()...)
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.ShortHelp()}
}

func renderTitle(note markdown.Entry) string {
	md := "# " + note.Title()
	out, _ := glamour.Render(md, "dark")

	return out
}

func (m Model) helpView() string {
	return helpStyle(m.help.View(m))
}

type keyMap struct {
	Back           key.Binding
	ToggleMarkdown key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Back:           key.NewBinding(key.WithKeys("backspace", "esc", "q"), key.WithHelp("esc/q", "back")),
		ToggleMarkdown: key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "toggle markdown")),
	}
}
