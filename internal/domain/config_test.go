package domain

import "testing"

func TestNewDungeonConfigValid(t *testing.T) {
	cfg, err := NewDungeonConfig(3, 2, 14*SecondsInHour, 2)
	if err != nil {
		t.Fatalf("NewDungeonConfig returned error: %v", err)
	}

	if cfg.RegularFloors() != 2 {
		t.Fatalf("RegularFloors() = %d, want 2", cfg.RegularFloors())
	}
	if cfg.CloseAtSeconds() != 16*SecondsInHour {
		t.Fatalf("CloseAtSeconds() = %d, want %d", cfg.CloseAtSeconds(), 16*SecondsInHour)
	}
}

func TestNewDungeonConfigInvalid(t *testing.T) {
	tests := []struct {
		name     string
		floors   int
		monsters int
		openAt   int
		duration int
	}{
		{name: "no boss floor", floors: 1, monsters: 1, openAt: 0, duration: 1},
		{name: "no monsters", floors: 2, monsters: 0, openAt: 0, duration: 1},
		{name: "bad duration", floors: 2, monsters: 1, openAt: 0, duration: 0},
		{name: "negative open time", floors: 2, monsters: 1, openAt: -1, duration: 1},
		{name: "open time outside day", floors: 2, monsters: 1, openAt: SecondsInDay, duration: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewDungeonConfig(tc.floors, tc.monsters, tc.openAt, tc.duration); err == nil {
				t.Fatalf("NewDungeonConfig expected error")
			}
		})
	}
}
