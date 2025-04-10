package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/charmbracelet/log"
	"github.com/samber/lo"
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
	entries := []Entry{}

	notes := splitEntries(readFile(file))

	for _, text := range notes {
		metadata := GetMetadata(text)

		id, ok := metadata["id"]
		if !ok {
			h := sha1.New()
			h.Write([]byte(text))
			id = hex.EncodeToString(h.Sum(nil))
		}

		var tags []string
		if t, ok := metadata["tags"]; ok {
			tags = strings.Split(t, ",")
		}

		relatedIdentifier := metadata["related"]
		removeHashtagPrefix := func(identifier string, index int) string {
			return strings.TrimLeft(identifier, "#")
		}
		relatedTags := lo.Map(strings.Split(relatedIdentifier, ","), removeHashtagPrefix)

		entry := NewEntry(id, file, 0, 0, text, tags, relatedTags)
		entries = append(entries, entry)
	}

	return entries
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
	return strings.Split(text, "\n---\n")
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
	entries := s.FindEntriesWithTags(entry.RelatedTags())

	return lo.Filter(entries, func(e Entry, index int) bool {
		return e.Id() != entry.Id()
	})
}
