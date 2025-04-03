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

	endOfBufferStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("235"))

	focusedPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99"))
)

type newNoteModel struct {
	textarea textarea.Model
	tagsInput textinput.Model
	viewport viewport.Model

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
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEsc:
			println("Gobak")
		case tea.KeyTab:
			if m.state == writingNote {
				m.viewport.Height = m.textarea.Height()
				m.viewport.SetContent(m.textarea.Value())
			}
			m.state = writingTags
			m.tagsInput.Focus()
		}
	}

	return m, tea.Batch(tiCmd, taCmd, vpCmd)
}

func (m newNoteModel) View() string {
	if m.state == writingTags {
		return fmt.Sprintf(
			"%s%s%s",
			lipgloss.NewStyle().Render(m.viewport.View()),
			gap,
			m.tagsInput.View(),
		)
	}
	return m.textarea.View() + gap + m.tagsInput.View()
}

func newNote() (*newNoteModel, error) {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().PaddingLeft(2)

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
		viewport: vp,
		textarea: ta,
		tagsInput: ti,
		state: writingNote,
	}, nil
}

