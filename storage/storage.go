package storage

import (
	"github.com/samber/lo"
	"log"
	"os"
	"regexp"
	"strings"
)

type Storage struct {
	sourceFiles []string
	storage     *map[string][]Entry
}

func readFile(file string) string {
	b, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	return string(b)
}

func loadEntries(file string) []Entry {
	items := []Entry{}

	notes := splitEntries(readFile(file))

	for _, text := range notes {
		metadata := GetMetadata(text)

		var tags []string
		if t, ok := metadata["tags"]; ok {
			tags = strings.Split(t, ",")
		}

		id := metadata["id"]
		relatedIdentifier := metadata["related"]

		itm := NewEntry(id, file, 0, 0, text, tags, relatedIdentifier)
		items = append(items, itm)
	}

	return items
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

func splitEntries(text string) []string {
	notes := strings.Split(text, "\n---\n")

	for index, note := range notes {
		notes[index] = strings.TrimSpace(note)
	}

	return notes
}

func (s *Storage) LoadFromFiles() {
	for _, file := range s.sourceFiles {
		entries := loadEntries(file)

		for _, n := range entries {
			if _, ok := (*s.storage)[n.Id()]; !ok {
				(*s.storage)[n.Id()] = []Entry{}
			}
			(*s.storage)[n.Id()] = append((*s.storage)[n.Id()], n)
		}
	}
}

func New(sourceFiles []string) *Storage {
	var s = make(map[string][]Entry)
	store := &Storage{
		sourceFiles: sourceFiles,
		storage:     &s,
	}
	store.LoadFromFiles()
	return store
}

func (s *Storage) Get(id string) ([]Entry, bool) {
	entry, ok := (*s.storage)[id]
	return entry, ok
}

func (s *Storage) GetLatest(id string) (Entry, bool) {
	entries, ok := s.Get(id)

	if !ok {
		return Entry{}, false
	}

	return lo.Last(entries)
}

func (s *Storage) GetLatestEntries() []Entry {
	return lo.Map(lo.Values(*s.storage), func(entries []Entry, index int) Entry {
		return lo.LastOrEmpty(entries)
	})
}

func (s *Storage) FindEntriesWithTags(tags []string) []Entry {
	return lo.Filter(s.GetLatestEntries(), func(entry Entry, index int) bool {
		return lo.Some(entry.Tags(), tags)
	})
}

func (s *Storage) GetRelatedTo(entry Entry) []Entry {
	relatedIdentifier := entry.RelatedIdentifier()

	entries := s.FindEntriesWithTags([]string{strings.TrimLeft(relatedIdentifier, "#")})
	return lo.Filter(entries, func(e Entry, index int) bool {
		return e.Id() != entry.Id()
	})
}
