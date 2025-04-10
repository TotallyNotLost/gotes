package cmd

import (
	"github.com/TotallyNotLost/gotes/storage"
	tea "github.com/charmbracelet/bubbletea"
)

func Back() tea.Msg {
	return BackMsg(1)
}

type BackMsg int

func NewEntry(file string, text string) tea.Cmd {
	return func() tea.Msg {
		return NewEntryMsg{
			entry: storage.NewEntry(file, text, 0, 0, 0),
		}
	}
}

type NewEntryMsg struct {
	entry storage.Entry
}

func (msg NewEntryMsg) GetEntry() storage.Entry {
	return msg.entry
}

func EditEntry(entry storage.Entry) tea.Cmd {
	return func() tea.Msg {
		return EditEntryMsg{
			entry: entry,
		}
	}
}

type EditEntryMsg struct {
	entry storage.Entry
}

func (msg EditEntryMsg) GetEntry() storage.Entry {
	return msg.entry
}

func ViewEntry(entry storage.Entry) tea.Cmd {
	return func() tea.Msg {
		return ViewEntryMsg{
			entry: entry,
		}
	}
}

type ViewEntryMsg struct {
	entry storage.Entry
}

func (msg ViewEntryMsg) GetEntry() storage.Entry {
	return msg.entry
}
