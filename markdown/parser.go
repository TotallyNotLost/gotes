package markdown

import (
	"fmt"
	"github.com/TotallyNotLost/gotes/storage"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"regexp"
	"strings"
)

func NewParser(storage *storage.Storage) *Parser {
	return &Parser{storage: storage}
}

type Parser struct {
	storage *storage.Storage
}

// Returns the expanded string and a list of all the IDs that couldn't be resolved.
func (p *Parser) Expand(md string) (string, []string) {
	link, u1 := p.expandLink(p.expandLinkShortSyntax(md))
	expanded, u2 := p.expandIncludes(link)
	return expanded, append(u1, u2...)
}

func (p *Parser) expandLinkShortSyntax(md string) string {
	r, _ := regexp.Compile("\\{\\$([-0-9a-zA-Z]+)\\}")

	return r.ReplaceAllStringFunc(md, func(metadata string) string {
		id := r.FindStringSubmatch(metadata)[1]

		return fmt.Sprintf("[_metadata_:link]:# \"$%s\"", id)
	})
}

func (p *Parser) expandLink(md string) (string, []string) {
	r, _ := regexp.Compile("\\[_metadata_:link\\]:# \"([^\"]*)\"")
	var unresolved []string

	expanded := r.ReplaceAllStringFunc(md, func(metadata string) string {
		identifier := p.normalizeIdentifier(r.FindStringSubmatch(metadata)[1])
		if strings.HasPrefix(identifier, "$") {
			id := strings.TrimLeft(identifier, "$")
			entry, ok := p.storage.GetLatest(id)
			if !ok {
				unresolved = append(unresolved, identifier)
			}
			text := entry.Text()
			title := lo.FirstOrEmpty(strings.Split(text, "\n"))

			return fmt.Sprintf("[%s](%s)", title, identifier)
		}
		return ""
	})

	return expanded, unresolved
}

func (p *Parser) expandIncludes(md string) (string, []string) {
	r, _ := regexp.Compile("\\[_metadata_:include\\]:# \"([^\"]*)\"")
	var unresolved []string

	expanded := r.ReplaceAllStringFunc(md, func(metadata string) string {
		identifier := p.normalizeIdentifier(r.FindStringSubmatch(metadata)[1])
		text, ok := p.getTextForIdentifier(identifier)
		if !ok {
			unresolved = append(unresolved, identifier)
		}
		sanitized := strings.TrimSpace(RemoveMetadata(RemoveMetadata(text, "id"), "tags"))

		return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(sanitized)
	})

	return expanded, unresolved
}

// Take the include and add optional parts.
// This will make it easier to parse/handle elsewhere.
//
// Normalized format:
// [selector]
func (p *Parser) normalizeIdentifier(incl string) string {
	return p.normalizeInclSelector(incl)
}

// Normalizes selector so that it fits one of these formats:
//
// 1. <epty> - Return the entire contents of the file.
// 2. #{id} -> The ID of a note within the file.
// 3. {start}-{end} -> Line numbers to include. Start is inclusive, end is exclusive.
func (p *Parser) normalizeInclSelector(selector string) string {
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

func (p *Parser) getTextForIdentifier(identifier string) (string, bool) {
	if strings.HasPrefix(identifier, "$") {
		id := strings.TrimLeft(identifier, "$")
		entry, ok := p.storage.GetLatest(id)
		if !ok {
			return "", false
		}
		return entry.Text(), true
	}

	return identifier, true
}
