package main

import "os"
import tea "github.com/charmbracelet/bubbletea"
import "github.com/charmbracelet/bubbles/list"
import "github.com/samber/lo"
import "log"
import "regexp"
import "strings"

func main() {
	items := []list.Item{}

	notes := strings.SplitSeq(ReadFile(os.Args[1]), "\n---\n")

	for note := range notes {
		title := strings.Split(normalize(note), "\n")[0]
		metadata := getMetadata(note)
		var itm item
		itm = item{title: title, content: note}
		if tags, ok := metadata["tags"]; ok {
			itm.desc = tags
		}
		items = append(items, itm)
	}

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = os.Args[1]

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

func ReadFile(file string) string {
	b, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	return string(b)
}

func SplitNotes(text string) []string {
	notes := strings.Split(text, "---")

	for index, note := range notes {
		notes[index] = strings.TrimSpace(note)
	}

	return notes
}

func normalize(text string) string {
	return strings.TrimSpace(removeMetadata(text))
}

func getMetadata(text string) map[string]string {
	lines := strings.Split(text, "\n")

	metaLines := lo.Filter(lines, func(line string, index int) bool {
		return isMetadata(line)
	})

	o := make(map[string]string)

	for _, ml := range metaLines {
		r, _ := regexp.Compile("^\\[_metadata_:*(\\w+)\\]:# \"(.*)\"$")
		key := r.FindStringSubmatch(ml)[1]
		value := r.FindStringSubmatch(ml)[2]
		o[key] = value
	}

	return o
}

func removeMetadata(text string) string {
	lines := strings.Split(text, "\n")

	filtered := lo.Filter(lines, func(line string, index int) bool {
		return !isMetadata(line)
	})

	return strings.Join(filtered, "\n")
}

func isMetadata(text string) bool {
	r, _ := regexp.Compile("^\\[_metadata_:\\w+\\]:# \".*\"$")

	return r.MatchString(text)
}
