package main

import "os"
import tea "github.com/charmbracelet/bubbletea"
import "github.com/charmbracelet/bubbles/list"
import "github.com/charmbracelet/bubbles/viewport"
import "github.com/charmbracelet/lipgloss"
import "github.com/google/uuid"
import "github.com/samber/lo"
import "log"
import "regexp"
import "slices"
import "strings"

type item struct {
	title, content, desc string
}

type mode int

const (
	browsing mode = 0
	viewing = 1
	creating = 2
)

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
	newNote *newNoteModel
	mode mode
	helpViewport viewport.Model
}

func (model viewportModel) Init() tea.Cmd {
	return nil
}

func (m viewportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.mode == creating {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				m.mode = browsing
				return m, nil
			}

			if msg.String() == "enter" && m.newNote.state == writingTags {
				note := "\n---\n" + m.newNote.textarea.Value() + "\n\n[_metadata_:id]:# \"" + uuid.New().String() + "\"\n[_metadata_:tags]:# \"" + m.newNote.tagsInput.Value() + "\""

				f, err := os.OpenFile(os.Args[1], os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				if err != nil {
					panic(err)
				}

				defer f.Close()

				if _, err = f.WriteString(note); err != nil {
					panic(err)
				}

				m.list.SetItems(loadNotes())
				m.mode = browsing
				return m, nil
			}
		}

		n, cmd := m.newNote.Update(msg)
		m.newNote = &n
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.viewport.Height += m.helpViewport.Height
			m.mode = browsing
			return m, nil
		case "enter":
			if m.mode == browsing {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					n, err := renderNote(i, m.viewport)
					if err != nil {
						panic(err)
					}
					m.viewport.SetContent(n)
					m.helpViewport.SetContent(helpView())
					m.viewport.Height -= m.helpViewport.Height
					m.mode = viewing
					return m, nil
				}
				return m, tea.Quit
			}
		case "n":
			m.newNote, _ = newNote()
			m.mode = creating
		case "e":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.newNote, _ = newNote()
				m.newNote.textarea.SetValue(i.content)
				metadata := getMetadata(i.content)
				if tags, ok := metadata["tags"]; ok {
					m.newNote.tagsInput.SetValue(tags)
				}
				m.mode = creating
			}
		default:
			if m.mode == viewing {
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.viewport.Height = msg.Height - 10
		if m.mode == browsing {
			m.list.SetSize(m.viewport.Width, m.viewport.Height)
		}
	}
	return m, nil
}

func (m viewportModel) View() string {
	if m.mode == browsing {
		m.viewport.SetContent(pageStyle.Render(m.list.View()))
	} else if m.mode == creating {
		m.viewport.SetContent(pageStyle.Render(m.newNote.View()))
	}

	if m.mode == viewing {

		m.helpViewport.SetContent(helpView())
		return mainStyle.Render(m.viewport.View() + "\n" + m.helpViewport.View())
	}

	return mainStyle.Render(m.viewport.View())
}

func newList() list.Model {
	items := loadNotes()

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = os.Args[1]

	return l
}

func loadNotes() []list.Item {
	items := []list.Item{}

	notes := strings.Split(ReadFile(os.Args[1]), "\n---\n")

	for i := range slices.Backward(notes) {
		note := notes[i]
		title := strings.Split(normalize(note), "\n")[0]
		metadata := getMetadata(note)
		var itm item
		itm = item{title: title, content: note}
		if tags, ok := metadata["tags"]; ok {
			itm.desc = tags
		}
		items = append(items, itm)
	}

	return items
}

func main() {
	const width = 78

	vp := viewport.New(width, 40)

	l := newList()

	p := tea.NewProgram(&viewportModel{ viewport: vp, list: l, helpViewport: viewport.New(width, 1) }, tea.WithAltScreen())

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
