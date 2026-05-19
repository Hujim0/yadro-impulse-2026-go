package tests

import (
	"dungeonGameLib/lib/engine/gamestate"
	"strings"
	"testing"
)

func TestDungeonTimeExpiration(t *testing.T) {
	config, err := loadConfig("test_config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	gameInstance, err := gamestate.CreateGameInstance(config.Floors, config.Monsters, config.OpenAt, config.Duration)
	if err != nil {
		t.Fatal("Failed to create game instance:", err)
	}

	go gameInstance.RunGameLoop()

	registerClose := "[16:04:59] 1 1\n"
	enterClose := "[16:04:59] 1 2\n"

	_, err = testProcessReader(strings.NewReader(registerClose+enterClose),
		gameInstance.InputEventChan, gameInstance.OutputEventChan)
	if err != nil {
		t.Fatalf("Failed to process events: %v", err)
	}

	finalReport := compilePlayerDataSafe(gameInstance)
	player1 := finalReport[0].Player
	if player1.State == gamestate.FAIL {
		t.Errorf("Expected player to be IN_GAME before timeout, got: %s", player1.State)
	}
}

func TestDungeonClosureEffects(t *testing.T) {
	config, err := loadConfig("test_config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	gameInstance, err := gamestate.CreateGameInstance(config.Floors, config.Monsters, config.OpenAt, config.Duration)
	if err != nil {
		t.Fatal("Failed to create game instance:", err)
	}

	go gameInstance.RunGameLoop()

	register1 := "[16:04:00] 1 1\n"
	enter1 := "[16:04:01] 1 2\n"
	register2 := "[16:04:00] 2 1\n"
	enter2 := "[16:04:01] 2 2\n"
	afterClose := "[16:05:00] 1 3\n"

	events := register1 + enter1 + register2 + enter2 + afterClose

	_, err = testProcessReader(strings.NewReader(events),
		gameInstance.InputEventChan, gameInstance.OutputEventChan)
	if err != nil {
		t.Fatalf("Failed to process events: %v", err)
	}

	finalReport := compilePlayerDataSafe(gameInstance)
	for _, pair := range finalReport {
		if pair.Player.Id == 1 || pair.Player.Id == 2 {
			if pair.Player.State == gamestate.IN_GAME {
				t.Errorf("Expected player %d to be FAIL after closure, got: IN_GAME", pair.Player.Id)
			}
		}
	}
}
