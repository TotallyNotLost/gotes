package main

import "os"
import tea "github.com/charmbracelet/bubbletea"
import "github.com/charmbracelet/bubbles/list"
import "github.com/charmbracelet/bubbles/viewport"
import "github.com/charmbracelet/lipgloss"
import "github.com/google/uuid"
import "github.com/samber/lo"
import gotescmd "github.com/TotallyNotLost/gotes/cmd"
import "github.com/TotallyNotLost/gotes/editor"
import "github.com/TotallyNotLost/gotes/markdown"
import "github.com/TotallyNotLost/gotes/viewer"
import "log"
import "regexp"
import "slices"
import "strings"

type item struct {
	id, title, content string
	tags               []string
}

type mode int

const (
	browsing mode = 0
	viewing       = 1
	editing       = 2
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
	viewport  viewport.Model
	list      list.Model
	viewer    viewer.Model
	editor    editor.Model
	mode      mode
	notes     []markdown.Entry
	noteInfos map[string][]noteInfo
}

func (model viewportModel) Init() tea.Cmd {
	return nil
}

func (m viewportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd, vcmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Height = msg.Height - 2
		m.viewport.Width = msg.Width - 10
		m.list.SetSize(m.viewport.Width, m.viewport.Height)
		m.viewer.SetHeight(m.viewport.Height)
		m.viewer.SetWidth(m.viewport.Width)
		m.editor.SetHeight(m.viewport.Height)
		m.editor.SetWidth(m.viewport.Width)
	case gotescmd.BackMsg:
		m.mode = browsing
		return m, nil
	case gotescmd.NewEntryMsg:
		writeEntry(msg.GetId(), msg.GetBody(), msg.GetTags())
		m.notes = loadEntries()
		m.list.SetItems(loadItems(m.notes))
		m.noteInfos = makeNoteInfos(m.notes)
		m.mode = browsing
		return m, nil
	}

	m.viewer, vcmd = m.viewer.Update(msg)
	if m.mode == viewing {
		return m, vcmd
	}

	m.editor, cmd = m.editor.Update(msg)
	if m.mode == editing {
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
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				noteInfos := m.noteInfos[i.id]
				notes := []markdown.Entry{}
				for _, noteInfo := range noteInfos {
					notes = append(notes, m.notes[noteInfo.index])
				}

				m.mode = viewing

				revisions := []markdown.Entry{}
				for _, no := range slices.Backward(notes) {
					revisions = append(revisions, no)
				}
				m.viewer.SetRevisions(revisions)
				return m, nil
			}
		case "n":
			m.editor.SetId(uuid.New().String())
			m.editor.SetBody("")
			m.editor.SetTags("")
			m.mode = editing
		case "e":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				body := i.title + "\n\n" + strings.TrimSpace(removeMetadata(removeMetadata(i.content, "id"), "tags"))
				metadata := markdown.GetMetadata(i.content)
				tags, _ := metadata["tags"]

				m.editor.SetId(i.id)
				m.editor.SetBody(body)
				m.editor.SetTags(tags)
				m.mode = editing
			}
		}
	}
	m.list, cmd = m.list.Update(msg)
	return m, tea.Batch(cmd, vcmd)
}

func (m viewportModel) View() string {
	var view string

	switch m.mode {
	case browsing:
		view = m.list.View()
	case viewing:
		view = m.viewer.View()
	case editing:
		view = m.editor.View()
	}

	m.viewport.SetContent(pageStyle.Render(view))

	return mainStyle.Render(m.viewport.View())
}

func newList(notes []markdown.Entry) list.Model {
	items := loadItems(notes)

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = os.Args[1]

	return l
}

func loadItems(notes []markdown.Entry) []list.Item {
	items := []list.Item{}

	noteInfos := makeNoteInfos(notes)

	for i := range slices.Backward(notes) {
		note := notes[i]

		isLatestRevision := i == noteInfos[note.Id()][len(noteInfos[note.Id()])-1].index
		if isLatestRevision && !lo.Contains(note.Tags(), "Done") {
			itm := item{id: note.Id(), title: note.Title(), content: note.Body(), tags: note.Tags()}
			items = append(items, itm)
		}
	}

	return items
}

func loadEntries() []markdown.Entry {
	items := []markdown.Entry{}

	notes := strings.SplitSeq(ReadFile(os.Args[1]), "\n---\n")

	for n := range notes {
		parts := strings.SplitN(n, "\n", 2)
		title := parts[0]
		body := parts[1]

		var id string
		var tags []string

		metadata := markdown.GetMetadata(n)
		if i, ok := metadata["id"]; ok {
			id = i
		}
		if t, ok := metadata["tags"]; ok {
			tags = strings.Split(t, ",")
		}

		itm := markdown.NewEntry(id, title, body, tags)

		items = append(items, itm)
	}

	return items
}

func writeEntry(id string, body string, tags []string) {
	f, err := os.OpenFile(os.Args[1], os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	f.WriteString("\n---\n")
	f.WriteString(body)

	tgs := strings.Join(tags, ",")

	if (id + tgs) != "" {
		f.WriteString("\n")
	}
	if id != "" {
		f.WriteString("\n[_metadata_:id]:# \"" + id + "\"")
	}
	if tgs != "" {
		f.WriteString("\n[_metadata_:tags]:# \"" + tgs + "\"")
	}
}

type noteInfo struct {
	index int
}

func makeNoteInfos(notes []markdown.Entry) map[string][]noteInfo {
	m := make(map[string][]noteInfo)

	for index, n := range notes {
		if _, ok := m[n.Id()]; !ok {
			m[n.Id()] = []noteInfo{}
		}

		m[n.Id()] = append(m[n.Id()], noteInfo{index: index})
	}

	return m
}

func main() {
	vp := viewport.New(0, 0)

	notes := loadEntries()
	l := newList(notes)
	noteInfos := makeNoteInfos(notes)
	viewer := viewer.New(os.Args[1])

	p := tea.NewProgram(&viewportModel{
		viewport:  vp,
		list:      l,
		viewer:    viewer,
		editor: editor.New(),
		notes:     notes,
		noteInfos: noteInfos,
	}, tea.WithAltScreen())

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

func removeMetadata(md string, key string) string {
	r, _ := regexp.Compile("\\[_metadata_:" + key + "\\]:# \"[^\"]*\"")

	return r.ReplaceAllString(md, "")
}
