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
	mode    mode
	list    list.Model
	viewer  viewer.Model
	editor  editor.Model
	storage *storage.Storage
}

func (model model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd, vcmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		width := msg.Width - mainStyle.GetWidth()
		height := msg.Height - lipgloss.Height(mainStyle.Render(""))
		m.list.SetSize(width, height)
		m.viewer.SetHeight(height)
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
	entries, ok := m.storage.Get(entry.Id())

	if !ok {
		panic(fmt.Sprintf("Couldn't find revisions for %s", entry.Id()))
	}

	m.mode = viewing

	revisions := []storage.Entry{}
	for _, no := range slices.Backward(entries) {
		revisions = append(revisions, no)
	}
	m.viewer.SetRevisions(revisions)
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

func writeEntry(entry storage.Entry) {
	f, err := os.OpenFile(entry.File(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	f.WriteString("\n---\n")
	f.WriteString(entry.String())
}
func main() {
	store := storage.New(lo.Uniq(os.Args[1:]))
	verify(store)
	items := latestEntriesAsItems(store)
	items = lo.Filter(items, func(item *list.Item, index int) bool {
		return item.File() == os.Args[1]
	})

	m := &model{
		list:    list.New(),
		editor:  editor.New(),
		viewer:  viewer.New(store),
		storage: store,
	}

	if len(items) == 1 {
		m.viewEntry(items[0].Entry())
	}

	m.list.SetItems(items)

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
