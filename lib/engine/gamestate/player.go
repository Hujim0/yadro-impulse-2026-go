package gamestate

import (
	"dungeonGameLib/lib/engine/event"
	"dungeonGameLib/lib/engine/floor"
	"dungeonGameLib/lib/engine/formatting"
	"fmt"
)

type Player struct {
	State                   PlayerState
	Id                      int
	Hp                      int
	CurrentFloor            int
	BossKillDurationSeconds int
	EnteredDungeonAtSeconds int
	ExitedDungeonAtSeconds  int
	Floors                  []*floor.GameFloor
}

func CreateNewPlayer(currentTime int, id int, Floors int, Monsters int) *Player {
	newPlayer := new(Player)
	newPlayer.Hp = 100
	newPlayer.State = IN_GAME
	newPlayer.EnteredDungeonAtSeconds = currentTime
	newPlayer.Id = id

	newPlayer.Floors = make([]*floor.GameFloor, Floors)

	for i := range newPlayer.Floors {
		newFloor := new(floor.GameFloor)
		isBossFloor := i == Floors-1
		if isBossFloor {
			newFloor.BossFloor = true
		} else {
			newFloor.MonsterCount = Monsters
		}

		newPlayer.Floors[i] = newFloor
	}

	return newPlayer
}

func (currentPlayer *Player) UpdateCompletedGameState() {
	for _, floor := range currentPlayer.Floors {
		if !floor.IsCompleted() {
			return
		}
	}

	currentPlayer.State = SUCCESS
}

func (currentPlayer *Player) GetAverageFloorTimeSeconds() int {
	sum := 0

	for _, floor := range currentPlayer.Floors {
		if !floor.BossFloor {
			sum += floor.TotalTimeSpentSeconds
		}
	}

	floorsCountWithoutBossFloor := len(currentPlayer.Floors) - 1

	return sum / floorsCountWithoutBossFloor
}

func (currentPlayer *Player) HandleEvent(event *event.GameEvent, gameInstance *GameInstance) (*event.GameEvent, error) {
	fun, ok := PlayerEventTypeToPlayerEventHandler[event.EventId]

	if !ok {
		return nil, fmt.Errorf("Player %d handler not found for event %s", currentPlayer.Id, event.EventId.String())
	}

	return fun(currentPlayer, event, gameInstance)
}

func (player *Player) String() string {
	spentInDungeonSeconds := player.ExitedDungeonAtSeconds - player.EnteredDungeonAtSeconds
	averageMonstersFloorClearDurationSeconds := player.GetAverageFloorTimeSeconds()

	if spentInDungeonSeconds < 0 {
		spentInDungeonSeconds = 0
	}

	// [SUCCESS] 1 [00:24:00, 00:05:00, 00:11:00] HP:35
	return fmt.Sprintf("[%s] %d [%s, %s, %s] HP:%d",
		player.State,
		player.Id,
		formatting.SecondsToFormattedDate(spentInDungeonSeconds),
		formatting.SecondsToFormattedDate(averageMonstersFloorClearDurationSeconds),
		formatting.SecondsToFormattedDate(player.BossKillDurationSeconds),
		player.Hp,
	)
}
