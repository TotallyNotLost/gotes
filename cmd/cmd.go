package cmd

import tea "github.com/charmbracelet/bubbletea"

func Back() tea.Msg {
	return BackMsg(1)
}

type BackMsg int
