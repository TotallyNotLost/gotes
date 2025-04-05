package markdown

type Entry struct {
	id    string
	title string
	body  string
	tags  []string
}

func NewEntry(id string, title string, body string, tags []string) Entry {
	return Entry{
		id:    id,
		title: title,
		body:  body,
		tags:  tags,
	}
}

func (n Entry) Id() string {
	return n.id
}

func (n Entry) Title() string {
	return n.title
}

func (n Entry) Body() string {
	return n.body
}

func (n Entry) Tags() []string {
	return n.tags
}
