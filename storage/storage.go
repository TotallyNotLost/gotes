package storage

import (
	"github.com/charmbracelet/log"
	"github.com/samber/lo"
	"os"
	"regexp"
	"sort"
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

	for index, text := range notes {
		entry := NewEntry(file, text, 0, 0, index)
		entries = append(entries, entry)
	}

	return entries
}

func GetMetadata(text string) map[string][]string {
	lines := strings.Split(text, "\n")

	metaLines := lo.Filter(lines, func(line string, index int) bool {
		return isMetadata(line)
	})

	o := make(map[string][]string)

	for _, ml := range metaLines {
		r, _ := regexp.Compile("^\\[_metadata_:*(\\w+)\\]:# \"(.*)\"$")
		key := r.FindStringSubmatch(ml)[1]
		value := r.FindStringSubmatch(ml)[2]
		o[key] = append(o[key], value)
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

func (s *Storage) loadFromFiles() {
	for _, file := range s.sourceFiles {
		entries := loadEntries(file)

		for _, n := range entries {
			s.AddEntry(n)
		}
	}
}

func (s *Storage) AddEntry(entry Entry) {
	id := entry.Id()
	list, _ := (*s.storage)[id]
	(*s.storage)[id] = append(list, entry)
}

func New(sourceFiles []string) *Storage {
	var s = make(map[string][]Entry)
	store := &Storage{
		sourceFiles: sourceFiles,
		storage:     &s,
	}
	store.loadFromFiles()
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

type ByIndex []Entry

func (i ByIndex) Len() int           { return len(i) }
func (a ByIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByIndex) Less(i, j int) bool { return a[i].index < a[j].index }

func (s *Storage) GetLatestEntries() []Entry {
	entries := lo.Map(lo.Values(*s.storage), func(es []Entry, index int) Entry {
		return lo.LastOrEmpty(es)
	})

	sort.Sort(sort.Reverse(ByIndex(entries)))

	return entries
}

func (s *Storage) FindEntriesRelatedTo(e Entry) []Entry {
	return lo.Filter(s.GetLatestEntries(), func(e2 Entry, index int) bool {
		return e.IsRelated(e2)
	})
}

func (s *Storage) GetRelatedTo(entry Entry) []Entry {
	entries := s.FindEntriesRelatedTo(entry)

	return lo.Filter(entries, func(e Entry, index int) bool {
		return e.Id() != entry.Id()
	})
}
