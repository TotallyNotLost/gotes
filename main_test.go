package main

import "reflect"
import "testing"

func TestSplitNotes(t *testing.T) {
	got := SplitNotes("Hello---World")
	want := []string{"Hello", "World"}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Expected %q, got %q", want, got)
	}
}

func TestSplitNotesTrimsWhitespace(t *testing.T) {
	got := SplitNotes("\n \nHello, there \n --- \n World\n \n")
	want := []string{"Hello, there", "World"}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Expected %q, got %q", want, got)
	}
}
