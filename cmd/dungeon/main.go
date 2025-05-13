package main

import (
	"log"
	"os"

	"dungeon-game/internal/adapter/input"
	"dungeon-game/internal/adapter/output"
	configloader "dungeon-game/internal/infrastructure/config"
	"dungeon-game/internal/usecase"
)

func main() {
	configPath := "config.json"
	eventsPath := "events"

	if len(os.Args) == 3 {
		configPath = os.Args[1]
		eventsPath = os.Args[2]
	} else if len(os.Args) != 1 {
		log.Fatalf("usage: %s [config.json events]", os.Args[0])
	}

	cfg, err := configloader.Load(configPath)
	if err != nil {
		log.Fatal(err)
	}

	eventsReader := input.NewEventReader()
	events, err := eventsReader.ReadFile(eventsPath)
	if err != nil {
		log.Fatal(err)
	}

	processor := usecase.NewProcessor(cfg)
	logs, report := processor.Run(events)

	presenter := output.NewPresenter()
	if err := presenter.Write(os.Stdout, logs, report); err != nil {
		log.Fatal(err)
	}
}
