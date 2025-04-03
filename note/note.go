package note

type Note struct {
	id string
	title string
	body string
	tags []string
}

func New(id string, title string, body string, tags []string) Note {
	return Note{
		id: id,
		title: title,
		body: body,
		tags: tags,
	}
}

func (n Note) Id() string {
	return n.id
}

func (n Note) Title() string {
	return n.title
}

func (n Note) Body() string {
	return n.body
}

func (n Note) Tags() []string {
	return n.tags
}
