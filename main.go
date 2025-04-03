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
	id, title, content string; tags []string
}

type mode int

const (
	browsing mode = 0
	viewing = 1
	creating = 2
)

func (i item) Title() string       { return i.title }
func (i item) Description() string { return strings.Join(i.tags, ",") }
func (i item) FilterValue() string { return i.title + " " + i.Description() }

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
	notes []note
	noteInfos map[string][]noteInfo
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
				f, err := os.OpenFile(os.Args[1], os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				if err != nil {
					panic(err)
				}

				defer f.Close()

				f.WriteString("\n---\n")
				f.WriteString(m.newNote.textarea.Value())

				id := m.newNote.id
				tags := m.newNote.tagsInput.Value()

				if (id + tags) != "" {
					f.WriteString("\n")
				}
				if id != "" {
					f.WriteString("\n[_metadata_:id]:# \"" + id + "\"")
				}
				if tags != "" {
					f.WriteString("\n[_metadata_:tags]:# \"" + tags + "\"")
				}

				m.list.SetItems(loadItems())
				m.notes = loadNotes()
				m.noteInfos = makeNoteInfos(m.notes)
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
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
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
					noteInfos := m.noteInfos[i.id]
					notes := []note{}
					for _, noteInfo := range noteInfos {
						notes = append(notes, m.notes[noteInfo.index])
					}
					n, err := renderNote(notes, m.viewport)
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
			m.newNote, _ = newNote(&m.viewport, uuid.New().String())
			m.mode = creating
		case "e":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.newNote, _ = newNote(&m.viewport, i.id)
				m.newNote.textarea.SetValue(i.title + "\n\n" + strings.TrimSpace(removeMetadata(i.content)))
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
		m.viewport.Height = msg.Height - 2
		m.viewport.Width = msg.Width - 10
		if m.mode == browsing {
			m.list.SetSize(m.viewport.Width, m.viewport.Height)
		}
	}
	m.list, cmd = m.list.Update(msg)
	return m, cmd
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
	items := loadItems()

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = os.Args[1]

	return l
}

func loadItems() []list.Item {
	items := []list.Item{}

	notes := loadNotes()
	noteInfos := makeNoteInfos(notes)

	for i := range slices.Backward(notes) {
		note := notes[i]

		if i == noteInfos[note.id][len(noteInfos[note.id]) - 1].index {
			itm := item{id: note.id, title: note.title, content: note.body, tags: note.tags}
			items = append(items, itm)
		}
	}

	return items
}

type note struct {
	id string
	title string
	body string
	tags []string
}
func loadNotes() []note {
	items := []note{}

	notes := strings.SplitSeq(ReadFile(os.Args[1]), "\n---\n")

	for n := range notes {
		parts := strings.SplitN(n, "\n", 2)
		title := parts[0]
		body := parts[1]
		itm := note{title:title, body:body}

		metadata := getMetadata(n)
		if id, ok := metadata["id"]; ok {
			itm.id = id
		}
		if tags, ok := metadata["tags"]; ok {
			itm.tags = strings.Split(tags, ",")
		}
		items = append(items, itm)
	}

	return items
}

type noteInfo struct {
	index int
}
func makeNoteInfos(notes []note) map[string][]noteInfo{
	m := make(map[string][]noteInfo)

	for index, n := range notes {
		if _, ok := m[n.id]; !ok {
			m[n.id] = []noteInfo{}
		}

		m[n.id] = append(m[n.id], noteInfo{ index: index })
	}

	return m
}

func main() {
	vp := viewport.New(0, 0)

	l := newList()
	notes := loadNotes()
	noteInfos := makeNoteInfos(notes)

	p := tea.NewProgram(&viewportModel{ viewport: vp, list: l, helpViewport: viewport.New(0, 1), notes: notes, noteInfos: noteInfos }, tea.WithAltScreen())

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
