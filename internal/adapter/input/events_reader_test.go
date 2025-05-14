package input

import (
	"strings"
	"testing"

	"dungeon-game/internal/domain"
)

func TestParseEventLineWithExtraMultipleWords(t *testing.T) {
	event, err := ParseEventLine("[14:10:05] 7 9 some serious reason")
	if err != nil {
		t.Fatalf("ParseEventLine returned error: %v", err)
	}

	wantAt := 14*domain.SecondsInHour + 10*60 + 5
	if event.At != wantAt {
		t.Fatalf("At = %d, want %d", event.At, wantAt)
	}
	if event.PlayerID != 7 {
		t.Fatalf("PlayerID = %d, want 7", event.PlayerID)
	}
	if event.ID != domain.EventCannotContinue {
		t.Fatalf("ID = %d, want %d", event.ID, domain.EventCannotContinue)
	}
	if event.Extra != "some serious reason" {
		t.Fatalf("Extra = %q, want %q", event.Extra, "some serious reason")
	}
}

func TestParseEventLineWithoutExtra(t *testing.T) {
	event, err := ParseEventLine("[00:00:00] 1 1")
	if err != nil {
		t.Fatalf("ParseEventLine returned error: %v", err)
	}
	if event.Extra != "" {
		t.Fatalf("Extra = %q, want empty", event.Extra)
	}
}

func TestParseEventLineInvalid(t *testing.T) {
	tests := []string{
		"",
		"[14:10:00] 1",
		"14:10:00 1 1",
		"[14:10] 1 1",
		"[14:10:00] bad 1",
		"[14:10:00] 1 bad",
		"[14:10:00] 0 1",
		"[14:10:00] 1 0",
	}

	for _, tc := range tests {
		t.Run(tc, func(t *testing.T) {
			if _, err := ParseEventLine(tc); err == nil {
				t.Fatalf("ParseEventLine(%q) expected error", tc)
			}
		})
	}
}

func TestEventReaderReadValid(t *testing.T) {
	reader := NewEventReader()
	events, err := reader.Read(strings.NewReader(`[14:00:00] 1 1
[14:05:00] 1 2
[14:05:00] 1 3`))
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("len(events) = %d, want 3", len(events))
	}
	if events[2].ID != domain.EventKillMonster {
		t.Fatalf("events[2].ID = %d, want %d", events[2].ID, domain.EventKillMonster)
	}
}

func TestEventReaderReadRejectsDecreasingTime(t *testing.T) {
	reader := NewEventReader()
	_, err := reader.Read(strings.NewReader(`[14:05:00] 1 1
[14:04:59] 1 2`))
	if err == nil {
		t.Fatal("Read expected error for decreasing event time")
	}
}

func TestEventReaderReadFileMissing(t *testing.T) {
	reader := NewEventReader()
	_, err := reader.ReadFile("/path/to/not-existing-events-file")
	if err == nil {
		t.Fatal("ReadFile expected error for missing file")
	}
}
