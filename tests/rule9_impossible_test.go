package tests

import (
	"dungeonGameLib/lib/engine/gamestate"
	"strings"
	"testing"
)

func TestImpossibleMoves(t *testing.T) {
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
	enter := "[14:00:01] 1 2\n"
	killMonster1 := "[14:00:02] 1 3\n"
	killMonster2 := "[14:00:03] 1 3\n"
	killMonster3 := "[14:00:04] 1 3\n"

	events := register + enter + killMonster1 + killMonster2 + killMonster3

	outputs, err := testProcessReader(strings.NewReader(events), gameInstance.InputEventChan, gameInstance.OutputEventChan)
	if err != nil {
		t.Fatalf("Failed to process events: %v", err)
	}

	if !contains(outputs, "Player [1] makes imposible move [3]") {
		t.Errorf("Expected impossible move message not found")
	}
}
