package main

import "github.com/charmbracelet/bubbles/viewport"
import "github.com/charmbracelet/glamour"
import "github.com/charmbracelet/lipgloss"

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
func helpView() string {
	return helpStyle("  ↑/↓: Navigate • esc: Back • q: Quit\n")
}

func renderNote(i item, vp viewport.Model) (string, error) {
	const width = 78

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
		return "", err
	}

	return renderer.Render("# " + i.title + "\n" + i.content)
}

