package main

import tea "github.com/charmbracelet/bubbletea"
import "github.com/charmbracelet/bubbles/viewport"
import "github.com/charmbracelet/glamour"
import "github.com/charmbracelet/lipgloss"

type example struct {
	viewport viewport.Model
}

func (e example) Init() tea.Cmd {
	return nil
}

func (e example) View() string {
	return e.viewport.View() + e.helpView()
}

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
func (e example) helpView() string {
	return helpStyle("\n  ↑/↓: Navigate • esc: Back • q: Quit\n")
}

func newExample(i item) (*example, error) {
	const width = 78

	vp := viewport.New(width, 20)

	// We need to adjust the width of the glamour render from our main width
	// to account for a few things:
	//
	//  * The viewport border width
	//  * The viewport padding
	//  * The viewport margins
	//  * The gutter glamour applies to the left side of the content
	//
	const glamourGutter = 2
	glamourRenderWidth := width - vp.Style.GetHorizontalFrameSize() - glamourGutter

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(glamourRenderWidth),
	)
	if err != nil {
		return nil, err
	}

	str, err := renderer.Render(i.content)
	if err != nil {
		return nil, err
	}

	vp.SetContent(str)

	return &example{
		viewport: vp,
	}, nil
}

