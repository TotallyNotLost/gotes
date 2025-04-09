package main

import (
	"github.com/samber/lo"
	"slices"
	"fmt"
	"github.com/google/uuid"
	"github.com/TotallyNotLost/gotes/editor"
	gotescmd "github.com/TotallyNotLost/gotes/cmd"
	"github.com/TotallyNotLost/gotes/list"
	"github.com/TotallyNotLost/gotes/viewer"
	"strings"
	"github.com/TotallyNotLost/gotes/markdown"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	blist "github.com/charmbracelet/bubbles/list"
	"os"
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
	mode   mode
	list   list.Model
	viewer viewer.Model
	editor editor.Model
	notes     []markdown.Entry
	noteInfos map[string][]noteInfo
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
			entry, _, ok = lo.FindLastIndexOf(m.notes, func(entry markdown.Entry) bool {
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
	items := loadItems(entries)
	p := tea.NewProgram(&model{
		list: list.New(lo.Map(items, func(item list.Item, index int) blist.Item {
			return item
		})),
		editor: editor.New(),
		viewer: viewer.New(os.Args[1]),
		notes: entries,
		noteInfos: noteInfos,
	}, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

func loadItems(entries []markdown.Entry) []list.Item {
	items := []list.Item{}

	noteInfos := makeNoteInfos(entries)

	for i := range slices.Backward(entries) {
		entry := entries[i]

		isLatestRevision := i == noteInfos[entry.Id()][len(noteInfos[entry.Id()])-1].index
		if isLatestRevision && !lo.Contains(entry.Tags(), "Done") {
			itm := list.EntryToItem(entry)
			items = append(items, itm)
		}
	}

	return items
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
