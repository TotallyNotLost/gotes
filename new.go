package main

import "fmt"
import tea "github.com/charmbracelet/bubbletea"
import "github.com/charmbracelet/bubbles/textarea"
import "github.com/charmbracelet/bubbles/textinput"
import "github.com/charmbracelet/lipgloss"

const gap = "\n\n"

type state int

const (
	writingNote state = 0
	writingTags = 1
)

var (
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230"))

	endOfBufferStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("235"))

	focusedPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99"))
)

type newNoteModel struct {
	textarea textarea.Model
	tagsInput textinput.Model

	state state
}

func (m newNoteModel) Init() tea.Cmd {
	return nil
}

func (m newNoteModel) Update(msg tea.Msg) (newNoteModel, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		taCmd tea.Cmd
		vpCmd tea.Cmd
	)

	if m.state == writingNote {
		m.textarea, taCmd = m.textarea.Update(msg)
	}
	m.tagsInput, tiCmd = m.tagsInput.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.textarea.SetWidth(msg.Width)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyTab:
			m.state = writingTags
			m.tagsInput.Focus()
		}
	}

	return m, tea.Batch(tiCmd, taCmd, vpCmd)
}

func (m newNoteModel) View() string {
	return m.textarea.View() + gap + m.tagsInput.View()
}

func newNote() (*newNoteModel, error) {
	ta := textarea.New()
	ta.Placeholder = "Note..."
	ta.ShowLineNumbers = false
	ta.Cursor.Style = cursorStyle
	ta.FocusedStyle.Placeholder = focusedPlaceholderStyle
	ta.FocusedStyle.CursorLine = cursorLineStyle
	ta.FocusedStyle.EndOfBuffer = endOfBufferStyle
	ta.Focus()

	ti := textinput.New()
	ti.Placeholder = "Tags..."

	return &newNoteModel{
		textarea: ta,
		tagsInput: ti,
		state: writingNote,
	}, nil
}

