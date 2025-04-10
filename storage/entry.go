package storage

type Entry struct {
	id          string
	file        string
	start       int
	end         int
	text        string
	tags        []string
	relatedTags []string
}

func NewEntry(id string, file string, start int, end int, text string, tags []string, relatedTags []string) Entry {
	return Entry{
		id:          id,
		file:        file,
		start:       start,
		end:         end,
		text:        text,
		tags:        tags,
		relatedTags: relatedTags,
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

func (e Entry) RelatedTags() []string {
	return e.relatedTags
}
