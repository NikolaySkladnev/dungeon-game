package domain

import "fmt"

type DungeonConfig struct {
	Floors        int
	Monsters      int
	OpenAtSeconds int
	DurationHours int
}

func NewDungeonConfig(floors, monsters, openAtSeconds, durationHours int) (DungeonConfig, error) {
	cfg := DungeonConfig{
		Floors:        floors,
		Monsters:      monsters,
		OpenAtSeconds: openAtSeconds,
		DurationHours: durationHours,
	}
	if err := cfg.Validate(); err != nil {
		return DungeonConfig{}, err
	}
	return cfg, nil
}

func (c DungeonConfig) Validate() error {
	if c.Floors < 2 {
		return fmt.Errorf("floors must be at least 2: got %d", c.Floors)
	}
	if c.Monsters <= 0 {
		return fmt.Errorf("monsters must be greater than 0: got %d", c.Monsters)
	}
	if c.DurationHours <= 0 {
		return fmt.Errorf("duration must be greater than 0: got %d", c.DurationHours)
	}
	if c.OpenAtSeconds < 0 || c.OpenAtSeconds >= SecondsInDay {
		return fmt.Errorf("open time must be within one day: got %d seconds", c.OpenAtSeconds)
	}
	return nil
}

func (c DungeonConfig) RegularFloors() int {
	return c.Floors - 1
}

func (c DungeonConfig) CloseAtSeconds() int {
	return c.OpenAtSeconds + c.DurationHours*SecondsInHour
}
