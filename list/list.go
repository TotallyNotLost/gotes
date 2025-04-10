package list

import (
	gotescmd "github.com/TotallyNotLost/gotes/cmd"
	"github.com/TotallyNotLost/gotes/markdown"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
	"os"
	"strings"
)

func EntryToItem(entry markdown.Entry) Item {
	return Item{id: entry.Id(), text: entry.Text(), tags: entry.Tags()}
}

type Item struct {
	id, text string
	tags     []string
}

func (i Item) Title() string       { return lo.FirstOrEmpty(strings.Split(i.text, "\n")) }
func (i Item) Description() string { return strings.Join(i.tags, ",") }
func (i Item) FilterValue() string { return i.text + " " + i.Description() }

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
				return m, gotescmd.ViewEntry(i.id)
			}
		case "n":
			return m, gotescmd.EditEntry("")
		case "e":
			i, ok := m.list.SelectedItem().(Item)
			if ok {
				return m, gotescmd.EditEntry(i.id)
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
