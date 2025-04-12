package tags

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	tagsStyle = lipgloss.NewStyle().MarginRight(1).MarginBottom(1).PaddingLeft(1).PaddingRight(1).Foreground(lipgloss.Color("232"))
	tagColors = []string{"#4d8dff", "#ffbf4d", "#bf4dff", "36", "204"}
)

func RenderTags(tags []string) string {
	tgs := []string{}

	for i, t := range tags {
		color := tagColors[i%len(tagColors)]
		tag := tagsStyle.Background(lipgloss.Color(color)).Render(t)
		tgs = append(tgs, tag)
	}

	return lipgloss.JoinHorizontal(lipgloss.Bottom, tgs...)
}
