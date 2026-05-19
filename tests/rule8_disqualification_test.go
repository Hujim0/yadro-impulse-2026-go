package tests

import (
	"dungeonGameLib/lib/engine/gamestate"
	"strings"
	"testing"
)

func TestDisqualificationScenarios(t *testing.T) {
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
	enterRegistered := "[14:00:01] 1 2\n"
	enterUnregistered1 := "[14:00:02] 2 2\n"
	enterUnregistered2 := "[14:00:03] 3 2\n"

	events := register + enterRegistered + enterUnregistered1 + enterUnregistered2

	outputs, err := testProcessReader(strings.NewReader(events), gameInstance.InputEventChan, gameInstance.OutputEventChan)
	if err != nil {
		t.Fatalf("Failed to process events: %v", err)
	}

	disqualified := 0
	for _, output := range outputs {
		if strings.Contains(output, "is disqualified") {
			disqualified++
		}
	}

	if disqualified != 2 {
		t.Errorf("Expected 2 disqualifications, got: %d", disqualified)
	}
}
