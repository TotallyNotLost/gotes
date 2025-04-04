package viewer

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/TotallyNotLost/gotes/note"
	"github.com/TotallyNotLost/gotes/tabs"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render

func New() Model {
	return Model{}
}

type Model struct {
	tabs tabs.Model
	revisions []note.Note
	activeRevision int
}

func (m *Model) SetHeight(height int) {
	// Subtract 1 to account for helpView()
	// and an additional 3 to account for the title at the top
	m.tabs.SetHeight(height - 4)
}

func (m *Model) SetWidth(width int) {
	m.tabs.SetWidth(width)
}

func (m *Model) SetRevisions(revisions []note.Note) {
	m.revisions = revisions
	tabs := lo.Map(m.revisions, func(revision note.Note, i int) tabs.Tab {
		body := expandIncludes(revision.Body())
		md, _ := glamour.Render(body, "dark")
		title := "Revision HEAD~" + strconv.Itoa(i)
		if i == 0 {
			title += " (latest)"
		}
		return tabs.NewTab(title, md)
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
	title := renderTitle(lo.LastOrEmpty(m.revisions))

	if len(m.revisions) == 1 {
		return title + m.tabs.GetTabs()[0].GetBody() + helpView()
	}

	return title + m.tabs.View() + helpView()
}

func expandIncludes(md string) string {
	r, _ := regexp.Compile("\\[_metadata_:include\\]:# \"([^\"]*)\"")

	return r.ReplaceAllStringFunc(md, func(incl string) string {
		file := r.FindStringSubmatch(incl)[1]
		b, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		return strings.TrimSpace(string(b))
	})
}

func renderTitle(note note.Note) string {
	md := "# " + note.Title()
	out, _ := glamour.Render(md, "dark")

	return out
}

func helpView() string {
	return helpStyle("\n  ↑/↓: Navigate • q: Quit\n")
}
