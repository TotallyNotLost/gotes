package list

import (
	gotescmd "github.com/TotallyNotLost/gotes/cmd"
	"github.com/TotallyNotLost/gotes/storage"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
	"os"
	"strings"
)

func EntryToItem(entry storage.Entry) Item {
	return Item{entry: entry}
}

type Item struct {
	entry storage.Entry
}

func (i Item) Id() string          { return i.entry.Id() }
func (i Item) File() string        { return i.entry.File() }
func (i Item) Title() string       { return lo.FirstOrEmpty(strings.Split(i.entry.Text(), "\n")) }
func (i Item) Description() string { return strings.Join(i.entry.Tags(), ",") }
func (i Item) FilterValue() string { return i.Title() + " " + i.Description() }

type Model struct {
	list list.Model
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd, vcmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "enter":
			i, ok := m.list.SelectedItem().(Item)
			if ok {
				return m, gotescmd.ViewEntry(i.entry.Id())
			}
		case "n":
			return m, gotescmd.EditEntry("")
		case "e":
			i, ok := m.list.SelectedItem().(Item)
			if ok {
				return m, gotescmd.EditEntry(i.entry.Id())
			}
		}
	}
	m.list, cmd = m.list.Update(msg)
	return m, tea.Batch(cmd, vcmd)
}

func (m Model) View() string {
	return m.list.View()
}

func (m *Model) SetSize(width int, height int) {
	m.list.SetSize(width, height)
}

func (m *Model) SetItems(items []Item) {
	m.list.SetItems(lo.Map(items, func(item Item, index int) list.Item {
		return item
	}))
}

func New() Model {
	return Model{
		list: newList(),
	}
}

func newList() list.Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = os.Args[1]

	return l
}
