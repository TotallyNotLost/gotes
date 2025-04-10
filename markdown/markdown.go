package markdown

import (
	"github.com/samber/lo"
	"log"
	"os"
	"regexp"
	"strings"
)

func LoadEntries(file string, filter func(Entry) bool) []Entry {
	items := []Entry{}

	notes := SplitEntries(ReadFile(file))

	for _, text := range notes {
		metadata := GetMetadata(text)

		var tags []string
		if t, ok := metadata["tags"]; ok {
			tags = strings.Split(t, ",")
		}

		id := metadata["id"]
		relatedIdentifier := metadata["related"]

		itm := NewEntry(id, file, 0, 0, text, tags, relatedIdentifier)

		if filter(itm) {
			items = append(items, itm)
		}
	}

	return items
}

func SplitEntries(text string) []string {
	notes := strings.Split(text, "\n---\n")

	for index, note := range notes {
		notes[index] = strings.TrimSpace(note)
	}

	return notes
}

func GetEntry(content string, id string) string {
	var note string

	notes := SplitEntries(content)

	for _, n := range notes {
		metadata := GetMetadata(n)
		if i, ok := metadata["id"]; ok {
			if i == id {
				note = n
			}
		}
	}

	return note
}

// Returns all entries that have at least one of the provided tags.
func GetEntriesWithTags(content string, tags []string) []Entry {
	return LoadEntries(content, func(entry Entry) bool {
		return lo.Some(entry.Tags(), tags)
	})
}

func GetMetadata(text string) map[string]string {
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

func RemoveMetadata(md string, key string) string {
	r, _ := regexp.Compile("\\[_metadata_:" + key + "\\]:# \"[^\"]*\"")

	return r.ReplaceAllString(md, "")
}

func isMetadata(text string) bool {
	r, _ := regexp.Compile("^\\[_metadata_:\\w+\\]:# \".*\"$")

	return r.MatchString(text)
}

func ReadFile(file string) string {
	b, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	return string(b)
}

func AllEntriesFilter(entry Entry) bool {
	return true
}
