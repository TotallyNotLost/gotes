package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/samber/lo"
	"regexp"
	"strings"
)

type Entry struct {
	id             string
	file           string
	start          int
	end            int
	text           string
	relatedIds     []string
	relatedRegexps []*regexp.Regexp
	// Index of the entry within the file
	// Used to order the entries roughly by age
	index int
}

// start(inclusive) and end(exclusive) are the
// offsets of text within file.
// index is the index of the text within the file.
// E.g. The first entry in the file has index 0, second index 2, etc.
func NewEntry(file string, text string, start int, end int, index int) Entry {
	metadata := GetMetadata(text)

	var id string
	ids, ok := metadata["id"]
	if !ok || len(ids) == 0 {
		h := sha1.New()
		h.Write([]byte(text))
		id = hex.EncodeToString(h.Sum(nil))
		text += fmt.Sprintf("\n[_metadata_:id]:# \"%s\"", id)
	} else {
		id = lo.LastOrEmpty(ids)
	}

	isNotEmpty := func(s string, index int) bool {
		return s != ""
	}
	relatedIdentifiers := lo.Filter(metadata["related"], isNotEmpty)
	hasPrefix := func(prefix string) func(string, int) bool {
		return func(identifier string, index int) bool {
			return strings.HasPrefix(identifier, prefix)
		}
	}
	isId := hasPrefix("id=")
	removePrefix := func(prefix string) func(string, int) string {
		return func(identifier string, index int) string {
			return strings.TrimLeft(identifier, prefix)
		}
	}
	relatedIds := lo.Map(lo.Filter(relatedIdentifiers, isId), removePrefix("id="))

	createRegexp := func(identifier string, index int) *regexp.Regexp {
		r, _ := regexp.Compile(identifier)
		return r
	}
	// Auto-match when an entry has this entry's id in its body.
	relatedRegexps := []*regexp.Regexp{createRegexp(fmt.Sprintf("\\$%s", id), 0)}

	isRegexp := hasPrefix("regexp=")
	relatedRegexps = append(relatedRegexps, lo.Map(lo.Filter(relatedIdentifiers, isRegexp), createRegexp)...)

	return Entry{
		id:             id,
		file:           file,
		start:          start,
		end:            end,
		text:           text,
		relatedIds:     relatedIds,
		relatedRegexps: relatedRegexps,
		index:          index,
	}
}

func (e Entry) String() string {
	return e.text
}

func (e Entry) Id() string {
	return e.id
}

func (e Entry) File() string {
	return e.file
}

func (e Entry) Start() int {
	return e.start
}

func (e Entry) End() int {
	return e.end
}

func (e Entry) Text() string {
	return e.text
}

func (e Entry) RelatedIds() []string {
	return e.relatedIds
}

func (e Entry) IsRelated(e2 Entry) bool {
	return isRelated(e, e2) || isRelated(e2, e)
}

func isRelated(e1 Entry, e2 Entry) bool {
	check := lo.Contains(e2.RelatedIds(), e1.Id())
	if check {
		return true
	}

	matches := func(r *regexp.Regexp, index int) bool {
		return r.Match([]byte(e1.Text()))
	}
	return len(lo.Filter(e2.relatedRegexps, matches)) > 0
}
