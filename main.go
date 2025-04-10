package main

import (
	"fmt"
	gotescmd "github.com/TotallyNotLost/gotes/cmd"
	"github.com/TotallyNotLost/gotes/editor"
	"github.com/TotallyNotLost/gotes/list"
	"github.com/TotallyNotLost/gotes/markdown"
	"github.com/TotallyNotLost/gotes/storage"
	"github.com/TotallyNotLost/gotes/viewer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"os"
	"slices"
	"strings"
)

type mode int

const (
	browsing mode = 0
	viewing       = 1
	editing       = 2
)

var mainStyle = lipgloss.NewStyle().
	Margin(0, 2).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62"))

type model struct {
	mode      mode
	list      list.Model
	viewer    viewer.Model
	editor    editor.Model
	entries   []storage.Entry
	noteInfos map[string][]storage.Entry
}

func (model model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd, vcmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		width := msg.Width - mainStyle.GetWidth()
		height := msg.Height - 2
		m.list.SetSize(width, height)
		m.viewer.SetHeight(height)
		m.viewer.SetWidth(width)
		m.editor.SetHeight(height)
		m.editor.SetWidth(width)
	case gotescmd.BackMsg:
		m.mode = browsing
		return m, nil
	case gotescmd.NewEntryMsg:
		writeEntry(msg.GetId(), msg.GetBody(), msg.GetTags())
		m.entries = markdown.LoadEntries(os.Args[1], markdown.AllEntriesFilter)
		m.noteInfos = makeNoteInfos(m.entries)
		m.list.SetItems(loadItems(m.noteInfos))
		m.mode = browsing
		return m, nil
	case gotescmd.EditEntryMsg:
		var entry storage.Entry
		id := msg.GetId()

		if id == "" {
			id = uuid.New().String()
		} else {
			var ok bool
			entry, _, ok = lo.FindLastIndexOf(m.entries, func(entry storage.Entry) bool {
				return entry.Id() == id
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
		entries := m.noteInfos[msg.GetId()]

		m.mode = viewing

		revisions := []storage.Entry{}
		for _, no := range slices.Backward(entries) {
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

	l, cmd := m.list.Update(msg)
	m.list = l.(list.Model)
	return m, tea.Batch(cmd, vcmd)
}

func (m model) View() string {
	var view string

	switch m.mode {
	case browsing:
		view = m.list.View()
	case viewing:
		view = m.viewer.View()
	case editing:
		view = m.editor.View()
	}

	return view
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
func main() {
	entries := markdown.LoadEntries(os.Args[1], markdown.AllEntriesFilter)
	noteInfos := makeNoteInfos(entries)
	items := loadItems(noteInfos)

	m := &model{
		list:      list.New(),
		editor:    editor.New(),
		viewer:    viewer.New(os.Args[1]),
		entries:   entries,
		noteInfos: noteInfos,
	}
	m.list.SetItems(items)

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

func loadItems(noteInfos map[string][]storage.Entry) []list.Item {
	return lo.Map(lo.Values(noteInfos), func(entries []storage.Entry, index int) list.Item {
		entry := lo.LastOrEmpty(entries)
		return list.EntryToItem(entry)
	})
}

func makeNoteInfos(entries []storage.Entry) map[string][]storage.Entry {
	m := make(map[string][]storage.Entry)

	for _, n := range entries {
		if _, ok := m[n.Id()]; !ok {
			m[n.Id()] = []storage.Entry{}
		}
		m[n.Id()] = append(m[n.Id()], n)
	}

	return m
}
