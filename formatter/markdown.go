package formatter

import (
	"github.com/charmbracelet/glamour"
	"github.com/TotallyNotLost/gotes/markdown"
)

func NewMarkdownFormatter(file string) MarkdownFormatter {
	return MarkdownFormatter{
		parser: markdown.NewParser(file),
	}
}

type MarkdownFormatter struct {
	parser markdown.Parser
	width int
}

func (mf *MarkdownFormatter) SetWidth(width int) {
	mf.width = width
}

func (mf MarkdownFormatter) Format(s string) string {
	r, _ := glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
		// wrap output at specific width (default is 80)
		glamour.WithWordWrap(mf.width),
	)
	md, _ := r.Render(mf.parser.Expand(s))
	return md
}

