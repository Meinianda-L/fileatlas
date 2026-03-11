package search

import (
	"testing"
	"time"

	"filecairn/internal/store"
)

func TestFindRanksRecentMatchHigher(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()
	records := []store.FileRecord{
		{
			ID:      "a",
			Path:    "/tmp/notes-old.txt",
			ModUnix: now - (180 * 24 * 3600),
			Tokens:  []string{"notes", "budget"},
			Labels:  []string{"topic:budget"},
		},
		{
			ID:      "b",
			Path:    "/Users/me/Documents/notes-recent.txt",
			ModUnix: now - (2 * 24 * 3600),
			Tokens:  []string{"notes", "budget"},
			Labels:  []string{"folder:documents", "topic:budget"},
		},
	}
	inverted := store.InvertedIndex{
		"budget": {"a", "b"},
	}

	results := Find(records, inverted, "budget", 10)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Record.ID != "b" {
		t.Fatalf("expected recent record first, got %s", results[0].Record.ID)
	}
}

func TestFindHandlesEmptyOrBadInputs(t *testing.T) {
	t.Parallel()

	records := []store.FileRecord{}
	inverted := store.InvertedIndex{}

	if got := Find(records, inverted, "   ", 10); len(got) != 0 {
		t.Fatalf("expected 0 results for empty query, got %d", len(got))
	}
	if got := Find(records, inverted, "notes", 0); len(got) != 0 {
		t.Fatalf("expected 0 results on empty index, got %d", len(got))
	}
}
