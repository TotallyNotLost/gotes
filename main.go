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
	"github.com/charmbracelet/log"
	"github.com/samber/lo"
	"os"
	"slices"
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
	mode         mode
	list         list.Model
	viewer       viewer.Model
	editor       editor.Model
	storage      *storage.Storage
	selectedFile string
	width        int
}

func (model model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd, vcmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		width := msg.Width - mainStyle.GetWidth()
		height := msg.Height
		m.list.SetSize(width, height)
		if m.width >= 100 {
			m.list.SetSize(min(40, int(0.3*float64(m.width))), height)
			width -= lipgloss.Width(m.list.View())
		}
		m.viewer.SetHeight(height - 2)
		m.viewer.SetWidth(width)
		m.editor.SetHeight(height)
		m.editor.SetWidth(width)
	case gotescmd.BackMsg:
		m.mode = browsing
		return m, nil
	case gotescmd.NewEntryMsg:
		m.newEntry(msg.GetEntry())
		return m, gotescmd.ViewEntry(msg.GetEntry())
	case gotescmd.EditEntryMsg:
		m.editor.SetEntry(msg.GetEntry())
		m.mode = editing
		return m, nil
	case gotescmd.ViewEntryMsg:
		m.viewEntry(msg.GetEntry())
		m.mode = viewing
		return m, nil
	}

	if m.mode == viewing {
		m.viewer, vcmd = m.viewer.Update(msg)
		return m, vcmd
	}

	if m.mode == editing {
		m.editor, cmd = m.editor.Update(msg)
		return m, cmd
	}

	l, cmd := m.list.Update(msg)
	m.list = l.(list.Model)
	m.viewEntry(m.list.SelectedItem().Entry())
	return m, tea.Batch(cmd, vcmd)
}

func (m model) newEntry(entry storage.Entry) {
	writeEntry(entry)
	m.storage.AddEntry(entry)
	items := latestEntriesAsItems(m.storage)
	m.list.SetItems(lo.Filter(items, func(item *list.Item, index int) bool {
		return item.File() == os.Args[1]
	}))
}

func (m *model) viewEntry(entry storage.Entry) {
	m.selectedFile = entry.File()
	m.SetItems()
	entries, ok := m.storage.Get(entry.Id())

	if !ok {
		panic(fmt.Sprintf("Couldn't find revisions for %s", entry.Id()))
	}

	revisions := []storage.Entry{}
	for _, no := range slices.Backward(entries) {
		revisions = append(revisions, no)
	}
	m.viewer.SetRevisions(revisions)
}

func (m model) View() string {
	m.viewer.SetFocused(false)
	m.list.SetFocused(false)

	var view string

	switch m.mode {
	case editing:
		view = m.editor.View()
	case browsing:
		m.list.SetFocused(true)
		view = m.listView()
	case viewing:
		m.viewer.SetFocused(true)
		view = m.viewerView()
	}

	return view
}

func (m model) listView() string {
	if m.width < 100 {
		return m.list.View()
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, m.list.View(), m.viewer.View())
}

func (m model) viewerView() string {
	if m.width < 100 {
		return m.viewer.View()
	}

	return m.listView()
}

func writeEntry(entry storage.Entry) {
	f, err := os.OpenFile(entry.File(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	f.WriteString("\n---\n")
	f.WriteString(entry.String())
}

func (m *model) SetItems() {
	items := latestEntriesAsItems(m.storage)
	items = lo.Filter(items, func(item *list.Item, index int) bool {
		return item.File() == m.selectedFile
	})

	if len(items) == 1 {
		m.viewEntry(items[0].Entry())
	}

	m.list.SetItems(items)
	m.list.SetTitle(m.selectedFile)
}

func main() {
	store := storage.New(lo.Uniq(os.Args[1:]))
	verify(store)

	m := &model{
		list:    list.New(),
		editor:  editor.New(),
		viewer:  viewer.New(store),
		storage: store,
	}

	m.selectedFile = os.Args[1]
	m.SetItems()

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

func verify(s *storage.Storage) {
	var unresolved []string
	for _, entry := range s.GetLatestEntries() {
		_, u := markdown.NewParser(s).Expand(entry.Text())
		unresolved = append(unresolved, u...)

		for _, id := range entry.RelatedIds() {
			_, ok := s.GetLatest(id)
			if !ok {
				unresolved = append(unresolved, id)
			}
		}
	}

	for _, identifier := range unresolved {
		log.Errorf("Couldn't resolve identifier %s", identifier)
	}
	if len(unresolved) != 0 {
		os.Exit(1)
	}
}

func latestEntriesAsItems(s *storage.Storage) []*list.Item {
	return lo.Map(s.GetLatestEntries(), func(entry storage.Entry, index int) *list.Item {
		return list.EntryToItem(s, entry)
	})
}
