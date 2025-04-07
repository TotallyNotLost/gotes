package markdown

type Entry struct {
	id   string
	text string
	tags []string
}

func NewEntry(id string, text string, tags []string) Entry {
	return Entry{
		id:   id,
		text: text,
		tags: tags,
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
