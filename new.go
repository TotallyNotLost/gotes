package main

import "fmt"
import tea "github.com/charmbracelet/bubbletea"
import "github.com/charmbracelet/bubbles/textarea"
import "github.com/charmbracelet/bubbles/textinput"
import "github.com/charmbracelet/bubbles/viewport"
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
)

type newNoteModel struct {
	textarea textarea.Model
	tagsInput textinput.Model
	id string
	state state
}

func (m newNoteModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m newNoteModel) Update(msg tea.Msg) (newNoteModel, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		taCmd tea.Cmd
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
		case tea.KeyShiftTab:
			m.state = writingNote
			m.tagsInput.Blur()
			m.textarea.Focus()
			m.tagsInput.PromptStyle = lipgloss.NewStyle()
			m.tagsInput.TextStyle = lipgloss.NewStyle()
		case tea.KeyTab:
			m.state = writingTags
			m.textarea.Blur()
			m.tagsInput.Focus()
			m.tagsInput.PromptStyle = focusedInputStyle
			m.tagsInput.TextStyle = focusedInputStyle
		}
	}

	return m, tea.Batch(tiCmd, taCmd)
}

func (m newNoteModel) View() string {
	return m.textarea.View() + gap + m.tagsInput.View()
}

func newNote(vp *viewport.Model, id string) (*newNoteModel, error) {
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

	ta.SetHeight(vp.Height - 5)
	ta.SetWidth(vp.Width)

	return &newNoteModel{
		textarea: ta,
		tagsInput: ti,
		id: id,
		state: writingNote,
	}, nil
}

