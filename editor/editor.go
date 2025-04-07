package editor

import (
	gotescmd "github.com/TotallyNotLost/gotes/cmd"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

const gap = "\n\n"

var (
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230"))

	placeholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))

	endOfBufferStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("235"))

	focusedPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99"))

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("238"))

	focusedInputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	blurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder())

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2).Render
)

func New() Model {
	ta := textarea.New()
	ta.Prompt = ""
	ta.Placeholder = "Note..."
	ta.ShowLineNumbers = true
	ta.Cursor.Style = cursorStyle
	ta.FocusedStyle.Placeholder = focusedPlaceholderStyle
	ta.BlurredStyle.Placeholder = placeholderStyle
	ta.FocusedStyle.CursorLine = cursorLineStyle
	ta.FocusedStyle.Base = focusedBorderStyle
	ta.BlurredStyle.Base = blurredBorderStyle
	ta.FocusedStyle.EndOfBuffer = endOfBufferStyle
	ta.BlurredStyle.EndOfBuffer = endOfBufferStyle
	ta.Focus()

	ti := textinput.New()
	ti.Placeholder = "Tags..."
	ti.PromptStyle.BorderStyle(lipgloss.RoundedBorder())

	return Model{
		textarea:  ta,
		tagsInput: ti,
		state:     writingNote,
		keyMap:    defaultKeyMap(),
		help:      help.New(),
	}
}

type Model struct {
	id        string
	textarea  textarea.Model
	tagsInput textinput.Model
	state     state
	keyMap    keyMap
	help      help.Model
}

func (m *Model) SetId(id string) {
	m.id = id
}

func (m *Model) SetText(text string) {
	m.textarea.SetValue(text)
}

func (m *Model) SetTags(tags string) {
	m.tagsInput.SetValue(tags)
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
		case key.Matches(msg, m.keyMap.Previous):
			m.state = writingNote
			m.tagsInput.Blur()
			m.textarea.Focus()
			m.tagsInput.PromptStyle = lipgloss.NewStyle()
			m.tagsInput.TextStyle = lipgloss.NewStyle()
		case key.Matches(msg, m.keyMap.Next):
			m.state = writingTags
			m.textarea.Blur()
			m.tagsInput.Focus()
			m.tagsInput.PromptStyle = focusedInputStyle
			m.tagsInput.TextStyle = focusedInputStyle
		case key.Matches(msg, m.keyMap.Submit) && m.state == writingTags:
			id := m.id
			text := m.textarea.Value()
			tags := m.tagsInput.Value()

			return m, gotescmd.NewEntry(id, text, strings.Split(tags, ","))
		}
	}

	var (
		tiCmd tea.Cmd
		taCmd tea.Cmd
	)

	if m.state == writingNote {
		m.textarea, taCmd = m.textarea.Update(msg)
	}
	m.tagsInput, tiCmd = m.tagsInput.Update(msg)

	return m, tea.Batch(tiCmd, taCmd)
}

func (m Model) View() string {
	return m.textarea.View() + gap + m.tagsInput.View() + gap + m.helpView()
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
	Back     key.Binding
	Previous key.Binding
	Next     key.Binding
	Submit   key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Back:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Previous: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shit+tab", "previous")),
		Next:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next")),
		Submit:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	}
}

type state int

const (
	writingNote state = 0
	writingTags       = 1
)
