package tests

import (
	"dungeonGameLib/lib/engine/gamestate"
	"strings"
	"testing"
)

func TestPlayerRegistrationBoundary(t *testing.T) {
	config, err := loadConfig("test_config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	gameInstance, err := gamestate.CreateGameInstance(config.Floors, config.Monsters, config.OpenAt, config.Duration)
	if err != nil {
		t.Fatal("Failed to create game instance:", err)
	}

	go gameInstance.RunGameLoop()

	register := "[14:00:00] 1 1\n"

	outputs, err := testProcessReader(strings.NewReader(register), gameInstance.InputEventChan, gameInstance.OutputEventChan)
	if err != nil {
		t.Fatalf("Failed to process events: %v", err)
	}

	if !contains(outputs, "Player [1] registered") {
		t.Errorf("Expected registration message not found")
	}
}

func TestUnregisteredPlayerActions(t *testing.T) {
	config, err := loadConfig("test_config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	gameInstance, err := gamestate.CreateGameInstance(config.Floors, config.Monsters, config.OpenAt, config.Duration)
	if err != nil {
		t.Fatal("Failed to create game instance:", err)
	}

	go gameInstance.RunGameLoop()

	unregisteredEnter := "[14:00:00] 99 2\n"

	outputs, err := testProcessReader(strings.NewReader(unregisteredEnter), gameInstance.InputEventChan, gameInstance.OutputEventChan)
	if err != nil {
		t.Fatalf("Failed to process events: %v", err)
	}

	if !contains(outputs, "Player [99] is disqualified") {
		t.Errorf("Expected disqualification message not found")
	}
}
