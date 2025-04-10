package editor

import (
	gotescmd "github.com/TotallyNotLost/gotes/cmd"
	"github.com/TotallyNotLost/gotes/storage"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n"

var (
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230"))

	endOfBufferStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("235"))

	focusedPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99"))

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("238"))

	focusedInputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2).Render
)

func New() Model {
	ta := textarea.New()
	ta.Prompt = ""
	ta.Placeholder = "Note..."
	ta.ShowLineNumbers = true
	ta.Cursor.Style = cursorStyle
	ta.FocusedStyle.Placeholder = focusedPlaceholderStyle
	ta.FocusedStyle.CursorLine = cursorLineStyle
	ta.FocusedStyle.Base = focusedBorderStyle
	ta.FocusedStyle.EndOfBuffer = endOfBufferStyle
	ta.Focus()

	return Model{
		textarea: ta,
		keyMap:   defaultKeyMap(),
		help:     help.New(),
	}
}

type Model struct {
	entry    storage.Entry
	textarea textarea.Model
	keyMap   keyMap
	help     help.Model
}

func (m *Model) SetEntry(entry storage.Entry) {
	m.entry = entry
	m.textarea.SetValue(entry.String())
}

func (m *Model) SetHeight(height int) {
	m.textarea.SetHeight(height - 5)
}

func (m *Model) SetWidth(width int) {
	m.textarea.SetWidth(width)
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Back):
			return m, gotescmd.Back
		case key.Matches(msg, m.keyMap.Submit):
			text := m.textarea.Value()

			return m, gotescmd.NewEntry(m.entry.File(), text)
		}
	}

	var cmd tea.Cmd

	m.textarea, cmd = m.textarea.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	return m.textarea.View() + gap + m.helpView()
}

func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{
		m.keyMap.Back,
		m.keyMap.Submit,
	}
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.ShortHelp()}
}

func (m Model) helpView() string {
	return helpStyle(m.help.View(m))
}

type keyMap struct {
	Back   key.Binding
	Submit key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Back:   key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "back")),
		Submit: key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "submit")),
	}
}
