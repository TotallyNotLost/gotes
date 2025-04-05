package markdown

import "reflect"
import "testing"

func TestSplitEntries(t *testing.T) {
	got := SplitEntries("Hello---World")
	want := []string{"Hello", "World"}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Expected %q, got %q", want, got)
	}
}

func TestSplitEntriesTrimsWhitespace(t *testing.T) {
	got := SplitEntries("\n \nHello, there \n --- \n World\n \n")
	want := []string{"Hello, there", "World"}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Expected %q, got %q", want, got)
	}
}
