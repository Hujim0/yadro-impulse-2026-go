package tests

import (
	"dungeonGameLib/lib/engine/gamestate"
	"strings"
	"testing"
)

func TestFloorMovementBoundaries(t *testing.T) {
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
	goPrevFloor := "[14:00:02] 1 5\n"

	outputs, err := testProcessReader(strings.NewReader(register+enter+goPrevFloor), gameInstance.InputEventChan, gameInstance.OutputEventChan)
	if err != nil {
		t.Fatalf("Failed to process events: %v", err)
	}

	if !contains(outputs, "Player [1] makes imposible move [5]") {
		t.Errorf("Expected impossible move message not found")
	}
}

func TestFloorTimeTracking(t *testing.T) {
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
	nextFloor := "[14:00:04] 1 4\n"
	enterBoss := "[14:00:05] 1 6\n"
	killBoss := "[14:00:06] 1 7\n"
	leave := "[14:00:07] 1 8\n"

	events := register + enter + killMonster1 + killMonster2 + nextFloor + enterBoss + killBoss + leave

	_, err = testProcessReader(strings.NewReader(events), gameInstance.InputEventChan, gameInstance.OutputEventChan)
	if err != nil {
		t.Fatalf("Failed to process events: %v", err)
	}

	finalReport := compilePlayerDataSafe(gameInstance)
	player1 := finalReport[0].Player
	if player1.State != gamestate.SUCCESS {
		t.Errorf("Expected SUCCESS state, got: %s", player1.State)
	}
}
