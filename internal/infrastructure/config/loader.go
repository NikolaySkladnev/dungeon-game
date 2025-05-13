package config

import (
	"encoding/json"
	"fmt"
	"os"

	"dungeon-game/internal/domain"
)

type fileConfig struct {
	Floors   int    `json:"Floors"`
	Monsters int    `json:"Monsters"`
	OpenAt   string `json:"OpenAt"`
	Duration int    `json:"Duration"`
}

func Load(path string) (domain.DungeonConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.DungeonConfig{}, fmt.Errorf("read config: %w", err)
	}

	var dto fileConfig
	if err := json.Unmarshal(data, &dto); err != nil {
		return domain.DungeonConfig{}, fmt.Errorf("parse config: %w", err)
	}

	openAtSeconds, err := domain.ParseClock(dto.OpenAt)
	if err != nil {
		return domain.DungeonConfig{}, fmt.Errorf("parse OpenAt: %w", err)
	}

	cfg, err := domain.NewDungeonConfig(dto.Floors, dto.Monsters, openAtSeconds, dto.Duration)
	if err != nil {
		return domain.DungeonConfig{}, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}
