package main

import "os"
import tea "github.com/charmbracelet/bubbletea"
import "github.com/charmbracelet/bubbles/list"
import "github.com/charmbracelet/bubbles/viewport"
import "github.com/charmbracelet/lipgloss"
import "github.com/samber/lo"
import "log"
import "regexp"
import "strings"

type item struct {
	title, content, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

var mainStyle = lipgloss.NewStyle().
		MarginLeft(2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)
var pageStyle = lipgloss.NewStyle()

type viewportModel struct {
	viewport viewport.Model
	list list.Model
	Chosen bool
}

func (model viewportModel) Init() tea.Cmd {
	return nil
}

func (m viewportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.Chosen = false
			return m, nil
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				e, _ := newExample(i)
				m.viewport.SetContent(pageStyle.Render(e.View()))
				m.Chosen = true
				return m, nil
			}
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			if m.Chosen {
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		if !m.Chosen {
			h, v := mainStyle.GetFrameSize()
			m.list.SetSize(msg.Width-h, msg.Height-v)
			m.list.SetSize(78, 40)
		}
	}
	return m, nil
}

func (m viewportModel) View() string {
	if !m.Chosen {
		m.viewport.SetContent(pageStyle.Render(m.list.View()))
	} else {
	}

	return mainStyle.Render(m.viewport.View())
}

func newList() list.Model {
	items := []list.Item{}

	notes := strings.SplitSeq(ReadFile(os.Args[1]), "\n---\n")

	for note := range notes {
		title := strings.Split(normalize(note), "\n")[0]
		metadata := getMetadata(note)
		var itm item
		itm = item{title: title, content: note}
		if tags, ok := metadata["tags"]; ok {
			itm.desc = tags
		}
		items = append(items, itm)
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = os.Args[1]

	return l
}

func main() {
	items := []list.Item{}

	notes := strings.SplitSeq(ReadFile(os.Args[1]), "\n---\n")

	for note := range notes {
		title := strings.Split(normalize(note), "\n")[0]
		metadata := getMetadata(note)
		var itm item
		itm = item{title: title, content: note}
		if tags, ok := metadata["tags"]; ok {
			itm.desc = tags
		}
		items = append(items, itm)
	}

	const width = 78

	vp := viewport.New(width, 40)

	l := newList()

	p := tea.NewProgram(&viewportModel{ viewport: vp, list: l }, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

func ReadFile(file string) string {
	b, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	return string(b)
}

func SplitNotes(text string) []string {
	notes := strings.Split(text, "---")

	for index, note := range notes {
		notes[index] = strings.TrimSpace(note)
	}

	return notes
}

func normalize(text string) string {
	return strings.TrimSpace(removeMetadata(text))
}

func getMetadata(text string) map[string]string {
	lines := strings.Split(text, "\n")

	metaLines := lo.Filter(lines, func(line string, index int) bool {
		return isMetadata(line)
	})

	o := make(map[string]string)

	for _, ml := range metaLines {
		r, _ := regexp.Compile("^\\[_metadata_:*(\\w+)\\]:# \"(.*)\"$")
		key := r.FindStringSubmatch(ml)[1]
		value := r.FindStringSubmatch(ml)[2]
		o[key] = value
	}

	return o
}

func removeMetadata(text string) string {
	lines := strings.Split(text, "\n")

	filtered := lo.Filter(lines, func(line string, index int) bool {
		return !isMetadata(line)
	})

	return strings.Join(filtered, "\n")
}

func isMetadata(text string) bool {
	r, _ := regexp.Compile("^\\[_metadata_:\\w+\\]:# \".*\"$")

	return r.MatchString(text)
}
