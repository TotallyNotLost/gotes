package list

import (
	"fmt"
	gotescmd "github.com/TotallyNotLost/gotes/cmd"
	"github.com/TotallyNotLost/gotes/storage"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"os"
	"slices"
	"strings"
)

func EntryToItem(s *storage.Storage, entry storage.Entry) *Item {
	var getLatestEntry = func(id string) (storage.Entry, bool) {
		return s.GetLatest(id)
	}
	return &Item{entry: entry, getLatestEntry: getLatestEntry}
}

type Item struct {
	storage        *storage.Storage
	entry          storage.Entry
	getLatestEntry func(id string) (storage.Entry, bool)
}

func title(entry storage.Entry) string {
	return lo.FirstOrEmpty(strings.Split(entry.Text(), "\n"))
}

func (i *Item) Entry() storage.Entry { return i.entry }
func (i *Item) File() string         { return i.entry.File() }
func (i *Item) Title() string        { return title(i.entry) }
func (i *Item) Description() string {
	tags := lo.Map(i.entry.RelatedIds(), func(id string, index int) string {
		entry, _ := i.getLatestEntry(id)
		return title(entry)
	})
	return strings.Join(tags, ",")
}
func (i *Item) FilterValue() string { return i.Title() + " " + i.Description() }

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
			i := m.SelectedItem()
			return m, gotescmd.ViewEntry(i.entry)
		case "n":
			text := fmt.Sprintf("[_metadata_:id]:# \"%s\"\n[_metadata_:related]:# \"\"", uuid.New().String())
			entry := storage.NewEntry(os.Args[1], text, 0, 0, 0)
			return m, gotescmd.EditEntry(entry)
		case "e":
			i, ok := m.list.SelectedItem().(*Item)
			if ok {
				return m, gotescmd.EditEntry(i.entry)
			}
		}
	}
	m.list, cmd = m.list.Update(msg)
	return m, tea.Batch(cmd, vcmd)
}

func (m Model) SelectedItem() *Item {
	return m.list.SelectedItem().(*Item)
}

func (m Model) View() string {
	return m.list.View()
}

func (m *Model) SetSize(width int, height int) {
	m.list.SetSize(width, height)
}

func (m *Model) SetItems(items []*Item) {
	notDoneItems := lo.Filter(items, func(item *Item, index int) bool {
		return !slices.Contains(item.entry.RelatedIds(), "#Done")
	})
	m.list.SetItems(lo.Map(notDoneItems, func(item *Item, index int) list.Item {
		return item
	}))
}

func (m *Model) SetFocused(focused bool) {
	var d list.DefaultDelegate
	d = list.NewDefaultDelegate()

	if !focused {
		// Change colors
		d.Styles.SelectedTitle = d.Styles.NormalTitle
		d.Styles.SelectedDesc = d.Styles.NormalDesc
	}

	m.list.SetDelegate(d)
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
