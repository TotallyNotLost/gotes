package markdown

import (
	"github.com/samber/lo"
	"regexp"
	"strings"
)

func SplitEntries(text string) []string {
	notes := strings.Split(text, "---")

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

func isMetadata(text string) bool {
	r, _ := regexp.Compile("^\\[_metadata_:\\w+\\]:# \".*\"$")

	return r.MatchString(text)
}
