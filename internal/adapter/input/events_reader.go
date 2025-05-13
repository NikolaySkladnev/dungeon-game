package input

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"dungeon-game/internal/domain"
)

type EventReader struct{}

func NewEventReader() EventReader {
	return EventReader{}
}

func (r EventReader) ReadFile(path string) ([]domain.Event, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open events file: %w", err)
	}
	defer file.Close()

	return r.Read(file)
}

func (r EventReader) Read(source io.Reader) ([]domain.Event, error) {
	scanner := bufio.NewScanner(source)
	events := make([]domain.Event, 0)

	lineNumber := 0
	previousTime := -1
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		event, err := ParseEventLine(line)
		if err != nil {
			return nil, fmt.Errorf("events line %d: %w", lineNumber, err)
		}
		if previousTime > event.At {
			return nil, fmt.Errorf("events line %d: event time is less than previous event time", lineNumber)
		}

		previousTime = event.At
		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read events: %w", err)
	}

	return events, nil
}

func ParseEventLine(line string) (domain.Event, error) {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return domain.Event{}, fmt.Errorf("expected format: [HH:MM:SS] playerID eventID [extra]")
	}

	rawTime := parts[0]
	if len(rawTime) != len("[00:00:00]") || rawTime[0] != '[' || rawTime[len(rawTime)-1] != ']' {
		return domain.Event{}, fmt.Errorf("invalid time token %q", rawTime)
	}

	at, err := domain.ParseClock(rawTime[1 : len(rawTime)-1])
	if err != nil {
		return domain.Event{}, err
	}

	playerID, err := strconv.Atoi(parts[1])
	if err != nil || playerID <= 0 {
		return domain.Event{}, fmt.Errorf("invalid player id %q", parts[1])
	}

	eventID, err := strconv.Atoi(parts[2])
	if err != nil || eventID <= 0 {
		return domain.Event{}, fmt.Errorf("invalid event id %q", parts[2])
	}

	return domain.Event{
		At:       at,
		PlayerID: playerID,
		ID:       domain.EventID(eventID),
		Extra:    strings.Join(parts[3:], " "),
	}, nil
}
