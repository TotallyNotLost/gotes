package formatter

import (
	"github.com/TotallyNotLost/gotes/markdown"
	"github.com/TotallyNotLost/gotes/storage"
	"github.com/charmbracelet/glamour"
)

func NewMarkdownFormatter(storage *storage.Storage) *MarkdownFormatter {
	renderer, _ := glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
		// wrap output at specific width (default is 80)
		// glamour.WithWordWrap(mf.width),
	)
	return &MarkdownFormatter{
		renderer: renderer,
		parser: markdown.NewParser(storage),
		memo:   make(map[string]string),
	}
}

type MarkdownFormatter struct {
	renderer *glamour.TermRenderer
	parser *markdown.Parser
	width  int
	memo   map[string]string
}

func (mf *MarkdownFormatter) SetWidth(width int) {
	mf.width = width
}

func (mf *MarkdownFormatter) Format(s string) string {
	var (
		f  string
		ok bool
	)

	if f, ok = mf.memo[s]; !ok {
		f = mf.format(s)
		mf.memo[s] = f
	}

	return f
}

func (mf *MarkdownFormatter) format(s string) string {
	expanded, _ := mf.parser.Expand(s)
	md, _ := mf.renderer.Render(expanded)
	return md
}
