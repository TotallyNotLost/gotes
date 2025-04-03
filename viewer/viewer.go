package viewer

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/TotallyNotLost/gotes/note"
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
	revisions []note.Note
}

func (m *Model) SetHeight(height int) {
	// Subtract 1 to account for helpView()
	m.viewport.Height = height - 1
}

func (m *Model) SetWidth(width int) {
	m.viewport.Width = width
}

func (m *Model) SetRevisions(revisions []note.Note) {
	m.revisions = revisions
}

func (m Model) View() string {
	latest := lo.LastOrEmpty(m.revisions)

	m.viewport.SetContent(renderNote(latest))

	return m.viewport.View() + helpView()
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
