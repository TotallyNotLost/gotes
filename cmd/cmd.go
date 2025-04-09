package cmd

import tea "github.com/charmbracelet/bubbletea"

func Back() tea.Msg {
	return BackMsg(1)
}

type BackMsg int

func NewEntry(id string, body string, tags []string) tea.Cmd {
	return func() tea.Msg {
		return NewEntryMsg{
			id:   id,
			body: body,
			tags: tags,
		}
	}
}

type NewEntryMsg struct {
	id   string
	body string
	tags []string
}

func (msg NewEntryMsg) GetId() string {
	return msg.id
}

func (msg NewEntryMsg) GetBody() string {
	return msg.body
}

func (msg NewEntryMsg) GetTags() []string {
	return msg.tags
}

func EditEntry(id string) tea.Cmd {
	return func() tea.Msg {
		return EditEntryMsg {
			id: id,
		}
	}
}

type EditEntryMsg struct {
	id string
}

func (msg EditEntryMsg) GetId() string {
	return msg.id
}

func ViewEntry(id string) tea.Cmd {
	return func() tea.Msg {
		return ViewEntryMsg {
			id: id,
		}
	}
}

type ViewEntryMsg struct {
	id string
}

func (msg ViewEntryMsg) GetId() string {
	return msg.id
}
