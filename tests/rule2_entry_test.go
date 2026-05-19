package tests

import (
	"dungeonGameLib/lib/engine/gamestate"
	"strings"
	"testing"
)

func TestDungeonEntryTiming(t *testing.T) {
	config, err := loadConfig("test_config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	gameInstance, err := gamestate.CreateGameInstance(config.Floors, config.Monsters, config.OpenAt, config.Duration)
	if err != nil {
		t.Fatal("Failed to create game instance:", err)
	}

	go gameInstance.RunGameLoop()

	register := "[13:59:59] 1 1\n"
	enter := "[13:59:59] 1 2\n"

	outputs, err := testProcessReader(strings.NewReader(register+enter), gameInstance.InputEventChan, gameInstance.OutputEventChan)
	if err != nil {
		t.Fatalf("Failed to process events: %v", err)
	}

	if !contains(outputs, "Player [1] registered") {
		t.Errorf("Expected registration message not found")
	}
}
