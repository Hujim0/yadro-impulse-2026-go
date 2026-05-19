package tests

import (
	"dungeonGameLib/lib/engine/gamestate"
	"strings"
	"testing"
)

func TestHealthBoundaries(t *testing.T) {
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
	healAboveMax := "[14:00:02] 1 10 200\n"

	_, err = testProcessReader(strings.NewReader(register+enter+healAboveMax), gameInstance.InputEventChan, gameInstance.OutputEventChan)
	if err != nil {
		t.Fatalf("Failed to process events: %v", err)
	}

	finalReport := compilePlayerDataSafe(gameInstance)
	player1 := finalReport[0].Player
	if player1.Hp != 100 {
		t.Errorf("Expected HP capped at 100, got: %d", player1.Hp)
	}
}

func TestDamageTiming(t *testing.T) {
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
	lethalDamage := "[14:00:02] 1 11 100\n"

	outputs, err := testProcessReader(strings.NewReader(register+enter+lethalDamage), gameInstance.InputEventChan, gameInstance.OutputEventChan)
	if err != nil {
		t.Fatalf("Failed to process events: %v", err)
	}

	if !contains(outputs, "Player [1] is dead") {
		t.Errorf("Expected death message not found")
	}

	finalReport := compilePlayerDataSafe(gameInstance)
	player1 := finalReport[0].Player
	if player1.State != gamestate.FAIL {
		t.Errorf("Expected FAIL state, got: %s", player1.State)
	}
}
