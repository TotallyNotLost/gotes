package main

import "fmt"
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
import "slices"
import "strings"

type item struct {
	id, text string
	tags     []string
}

type mode int

const (
	browsing mode = 0
	viewing       = 1
	editing       = 2
)

func (i item) Title() string       { return lo.FirstOrEmpty(strings.Split(i.text, "\n")) }
func (i item) Description() string { return strings.Join(i.tags, ",") }
func (i item) FilterValue() string { return i.text + " " + i.Description() }

var mainStyle = lipgloss.NewStyle().
	Margin(0, 2).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62"))
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
		m.viewport.Width = msg.Width - lipgloss.Width(mainStyle.Render(""))
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
		m.notes = markdown.LoadEntries(os.Args[1], markdown.AllEntriesFilter)
		m.list.SetItems(loadItems(m.notes))
		m.noteInfos = makeNoteInfos(m.notes)
		m.mode = browsing
		return m, nil
	case gotescmd.EditEntryMsg:
		var entry markdown.Entry
		id := msg.GetId()

		if id == "" {
			id = uuid.New().String()
		} else {
			var ok bool
			entry, _, ok = lo.FindLastIndexOf(m.notes, func(item markdown.Entry) bool {
				return item.Id() == id
			})
			if !ok {
				panic(fmt.Sprintf("Can't find entry %s", id))
			}
		}

		text := strings.TrimSpace(markdown.RemoveMetadata(markdown.RemoveMetadata(entry.Text(), "id"), "tags"))
		metadata := markdown.GetMetadata(entry.Text())
		tags, _ := metadata["tags"]

		m.editor.SetId(id)
		m.editor.SetText(strings.TrimSpace(text))
		m.editor.SetTags(tags)
		m.mode = editing
		return m, nil
	case gotescmd.ViewEntryMsg:
		noteInfos := m.noteInfos[msg.GetId()]
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
				return m, gotescmd.ViewEntry(i.id)
			}
		case "n":
			return m, gotescmd.EditEntry("")
		case "e":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				return m, gotescmd.EditEntry(i.id)
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
			itm := item{id: note.Id(), text: note.Text(), tags: note.Tags()}
			items = append(items, itm)
		}
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

	notes := markdown.LoadEntries(os.Args[1], markdown.AllEntriesFilter)
	l := newList(notes)
	noteInfos := makeNoteInfos(notes)
	viewer := viewer.New(os.Args[1])

	p := tea.NewProgram(&viewportModel{
		viewport:  vp,
		list:      l,
		viewer:    viewer,
		editor:    editor.New(),
		notes:     notes,
		noteInfos: noteInfos,
	}, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
