package viewer

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/TotallyNotLost/gotes/markdown"
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
	// The source file that the notes come from.
	// This is necessary in case the notes want to
	// reference/include other notes or lines in the file.
	file string
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
		body := m.expandIncludes(revision.Body())
		md, _ := glamour.Render(body, "dark")
		title := "Revision HEAD~" + strconv.Itoa(i)
		if i == 0 {
			title += " (latest)"
		}
		return tabs.NewTab(title, md)
	})
	m.tabs.SetTabs(tabs)
}

func (m *Model) SetFile(file string) {
	m.file = file
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

func (m Model) expandIncludes(md string) string {
	r, _ := regexp.Compile("\\[_metadata_:include\\]:# \"([^\"]*)\"")

	return r.ReplaceAllStringFunc(md, func(metadata string) string {
		incl := m.normalizeIncl(r.FindStringSubmatch(metadata)[1])

		parts := strings.SplitN(incl, ":", 2)
		file := parts[0]

		b, err := os.ReadFile(file)
		if err != nil {
			return fmt.Sprintf("{Error loading file \"%s\"}", incl)
		}

		selector := parts[1]

		if selector == "" {
			return strings.TrimSpace(string(b))
		}

		if strings.HasPrefix(selector, "#") {
			id := strings.TrimLeft(selector, "#")
			return markdown.GetEntry(string(b), id)
		}

		rng := strings.Split(selector, "-")
		start, _ := strconv.Atoi(rng[0])
		end, _ := strconv.Atoi(rng[1])

		lines := strings.Split(string(b), "\n")
		return strings.Join(lines[start:end], "\n")
	})
}

// Take the include and add optional parts.
// This will make it easier to parse/handle elsewhere.
//
// Normalized format:
// [file-path]:[selector]
func (m Model) normalizeIncl(incl string) string {
	var (file, selector string)

	selreg, _ := regexp.Compile("(^#.+$)|(^\\d+(-\\d+)?$)")

	if strings.Contains(incl, ":") {
		parts := strings.SplitN(incl, ":", 2)
		file = parts[0]
		selector = parts[1]
	} else if selreg.MatchString(incl) {
		file = m.file
		selector = incl
	} else {
		file = incl
		selector = ""
	}

	return fmt.Sprintf("%s:%s", m.normalizeInclFile(file), m.normalizeInclSelector(selector))
}

func (m Model) normalizeInclFile(file string) string {
	if file != "" {
		return file
	}

	return m.file
}

// Normalizes selector so that it fits one of these formats:
//
// 1. <empty> - Return the entire contents of the file.
// 2. #{id} -> The ID of a note within the file.
// 3. {start}-{end} -> Line numbers to include. Start is inclusive, end is exclusive.
func (m Model) normalizeInclSelector(selector string) string {
	if selector == "" {
		return selector
	}

	r, _ := regexp.Compile("^(#.+)|(\\d+-\\d+)$")

	if r.MatchString(selector) {
		return selector
	}

	return fmt.Sprintf("%s-%s", selector, selector)
}

func renderTitle(note note.Note) string {
	md := "# " + note.Title()
	out, _ := glamour.Render(md, "dark")

	return out
}

func helpView() string {
	return helpStyle("\n  ↑/↓: Navigate • q: Quit\n")
}
