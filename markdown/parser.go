package markdown

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func NewParser(file string) Parser {
	return Parser{file: file}
}

type Parser struct {
	// The source file that the markdown comes from.
	// This is necessary in case the markdown has metadata that
	// references/includes other notes or lines in the file.
	file string
}

func (p Parser) Expand(md string) string {
	return p.expandIncludes(p.expandLink(p.expandLinkShortSyntax(md)))
}

func (p Parser) expandLinkShortSyntax(md string) string {
	r, _ := regexp.Compile("\\{([-0-9a-zA-Z]+)\\}")

	return r.ReplaceAllStringFunc(md, func(metadata string) string {
		id := r.FindStringSubmatch(metadata)[1]

		return fmt.Sprintf("[_metadata_:link]:# \"$%s\"", id)
	})
}

func (p Parser) expandLink(md string) string {
	r, _ := regexp.Compile("\\[_metadata_:link\\]:# \"([^\"]*)\"")
	return r.ReplaceAllStringFunc(md, func(metadata string) string {
		identifier := p.normalizeIdentifier(r.FindStringSubmatch(metadata)[1])
		text := p.getTextForIdentifier(identifier)
		title := lo.FirstOrEmpty(strings.Split(text, "\n"))

		return fmt.Sprintf("[%s](%s)", title, identifier)
	})
}

func (p Parser) expandIncludes(md string) string {
	r, _ := regexp.Compile("\\[_metadata_:include\\]:# \"([^\"]*)\"")

	return r.ReplaceAllStringFunc(md, func(metadata string) string {
		identifier := p.normalizeIdentifier(r.FindStringSubmatch(metadata)[1])
		text := p.getTextForIdentifier(identifier)
		sanitized := strings.TrimSpace(RemoveMetadata(RemoveMetadata(text, "id"), "tags"))

		return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(sanitized)
	})
}

// Take the include and add optional parts.
// This will make it easier to parse/handle elsewhere.
//
// Normalized format:
// [file-path]:[selector]
func (p Parser) normalizeIdentifier(incl string) string {
	var (
		file, selector string
	)

	selreg, _ := regexp.Compile("(^\\$.+$)|(^#.+$)|(^\\d+(-\\d+)?$)")

	if strings.Contains(incl, ":") {
		parts := strings.SplitN(incl, ":", 2)
		file = parts[0]
		selector = parts[1]
	} else if selreg.MatchString(incl) {
		file = p.file
		selector = incl
	} else {
		file = incl
		selector = ""
	}

	return fmt.Sprintf("%s:%s", p.normalizeInclFile(file), p.normalizeInclSelector(selector))
}

func (p Parser) normalizeInclFile(file string) string {
	if file != "" {
		return file
	}

	return p.file
}

// Normalizes selector so that it fits one of these formats:
//
// 1. <epty> - Return the entire contents of the file.
// 2. #{id} -> The ID of a note within the file.
// 3. {start}-{end} -> Line numbers to include. Start is inclusive, end is exclusive.
func (p Parser) normalizeInclSelector(selector string) string {
	if selector == "" {
		return selector
	}

	valid, _ := regexp.Compile("^(\\$.+)|(\\d+-\\d+)$")

	if valid.MatchString(selector) {
		return selector
	}

	// This appears to be a line number selector.
	// Expand this to be a line range.
	// TODO: This should go to selector+1 since end is exclusive.
	return fmt.Sprintf("%s-%s", selector, selector)
}

func (p Parser) getTextForIdentifier(identifier string) string {
	parts := strings.SplitN(identifier, ":", 2)
	file := parts[0]

	b, err := os.ReadFile(file)
	if err != nil {
		return fmt.Sprintf("{Error loading file \"%s\"}", identifier)
	}

	selector := parts[1]

	if selector == "" {
		return strings.TrimSpace(string(b))
	}

	if strings.HasPrefix(selector, "$") {
		id := strings.TrimLeft(selector, "$")
		return GetEntry(string(b), id)
	}

	rng := strings.Split(selector, "-")
	start, _ := strconv.Atoi(rng[0])
	end, _ := strconv.Atoi(rng[1])

	lines := strings.Split(string(b), "\n")
	return strings.Join(lines[start:end], "\n")
}
