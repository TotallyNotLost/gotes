package markdown

type Entry struct {
	id                string
	text              string
	tags              []string
	relatedIdentifier string
}

func NewEntry(id string, text string, tags []string, relatedIdentifier string) Entry {
	return Entry{
		id:                id,
		text:              text,
		tags:              tags,
		relatedIdentifier: relatedIdentifier,
	}
}

func (n Entry) Id() string {
	return n.id
}

func (n Entry) Text() string {
	return n.text
}

func (n Entry) Tags() []string {
	return n.tags
}

func (n Entry) RelatedIdentifier() string {
	return n.relatedIdentifier
}
