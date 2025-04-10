package storage

import (
	"github.com/samber/lo"
	"regexp"
)

type Entry struct {
	id             string
	file           string
	start          int
	end            int
	text           string
	tags           []string
	relatedTags    []string
	relatedRegexps []*regexp.Regexp
	// Index of the entry within the file
	// Used to order the entries roughly by age
	index int
}

func NewEntry(id string, file string, start int, end int, text string, tags []string, relatedTags []string, relatedRegexps []*regexp.Regexp, index int) Entry {
	return Entry{
		id:             id,
		file:           file,
		start:          start,
		end:            end,
		text:           text,
		tags:           tags,
		relatedTags:    relatedTags,
		relatedRegexps: relatedRegexps,
		index:          index,
	}
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

func (e Entry) Tags() []string {
	return e.tags
}

func (e Entry) IsRelated(e2 Entry) bool {
	return isRelated(e, e2) || isRelated(e2, e)
}

func isRelated(e1 Entry, e2 Entry) bool {
	some := lo.Some(e1.relatedTags, e2.Tags())

	if some {
		return true
	}

	matches := func(r *regexp.Regexp, index int) bool {
		return r.Match([]byte(e1.Text()))
	}
	return len(lo.Filter(e2.relatedRegexps, matches)) > 0
}
