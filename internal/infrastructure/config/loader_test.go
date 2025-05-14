package config

import (
	"os"
	"path/filepath"
	"testing"

	"dungeon-game/internal/domain"
)

func TestLoadValidConfig(t *testing.T) {
	path := writeTempConfig(t, `{
		"Floors": 3,
		"Monsters": 2,
		"OpenAt": "14:05:00",
		"Duration": 2
	}`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	wantOpenAt := 14*domain.SecondsInHour + 5*60
	if cfg.Floors != 3 {
		t.Fatalf("Floors = %d, want 3", cfg.Floors)
	}
	if cfg.Monsters != 2 {
		t.Fatalf("Monsters = %d, want 2", cfg.Monsters)
	}
	if cfg.OpenAtSeconds != wantOpenAt {
		t.Fatalf("OpenAtSeconds = %d, want %d", cfg.OpenAtSeconds, wantOpenAt)
	}
	if cfg.DurationHours != 2 {
		t.Fatalf("DurationHours = %d, want 2", cfg.DurationHours)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("Load expected error for missing file")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	path := writeTempConfig(t, `{bad json}`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load expected error for invalid json")
	}
}

func TestLoadInvalidOpenAt(t *testing.T) {
	path := writeTempConfig(t, `{
		"Floors": 2,
		"Monsters": 2,
		"OpenAt": "25:00:00",
		"Duration": 2
	}`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load expected error for invalid OpenAt")
	}
}

func TestLoadInvalidDomainConfig(t *testing.T) {
	path := writeTempConfig(t, `{
		"Floors": 1,
		"Monsters": 2,
		"OpenAt": "14:05:00",
		"Duration": 2
	}`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load expected error for invalid domain config")
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}
