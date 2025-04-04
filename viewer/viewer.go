package viewer

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/TotallyNotLost/gotes/note"
	"github.com/TotallyNotLost/gotes/tabs"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render

func New() Model {
	vp := viewport.New(0, 0)

	return Model{
		viewport: vp,
	}
}

type Model struct {
	viewport viewport.Model
	tabs tabs.Model
	revisions []note.Note
	activeRevision int
}

func (m *Model) SetHeight(height int) {
	// Subtract 1 to account for helpView()
	m.viewport.Height = height - 1
	m.tabs.SetHeight(height - 1)
}

func (m *Model) SetWidth(width int) {
	m.viewport.Width = width
	m.tabs.SetWidth(width)
}

func (m *Model) SetRevisions(revisions []note.Note) {
	m.revisions = revisions
	tabs := lo.Map(m.revisions, func(revision note.Note, i int) tabs.Tab {
		md, _ := glamour.Render(revision.Body(), "dark")
		return tabs.NewTab(revision.Title(), md)
	})
	m.tabs.SetTabs(tabs)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.tabs, cmd = m.tabs.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	// m.viewport.SetContent(renderNote(latest))

	// return m.viewport.View() + helpView()
	return m.tabs.View() + helpView()
}

func renderNote(note note.Note) string {
	heading := "# " + note.Title()
	markdown := heading + "\n" + note.Body()
	out, _ := glamour.Render(markdown, "dark")

	return out
}

func helpView() string {
	return helpStyle("\n  ↑/↓: Navigate • q: Quit\n")
}
