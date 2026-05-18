package engine

import playerstate "dungeonGameLib/lib/engine/player_state"

type Player struct {
	State                   playerstate.PlayerState
	Id                      int
	Hp                      int
	CurrentFloor            int
	BossKillDurationSeconds int
	EnteredDungeonAtSeconds int
	ExitedDungeonAtSeconds  int
	floors                  []*GameFloor
}

func createNewPlayer(currentTime int, id int, Floors int, Monsters int) *Player {
	newPlayer := new(Player)
	newPlayer.Hp = 100
	newPlayer.State = playerstate.IN_GAME
	newPlayer.EnteredDungeonAtSeconds = currentTime
	newPlayer.Id = id

	newPlayer.floors = make([]*GameFloor, Floors)

	for i := range newPlayer.floors {
		newFloor := new(GameFloor)
		isBossFloor := i == Floors-1
		if isBossFloor {
			newFloor.bossFloor = true
		} else {
			newFloor.monsterCount = Monsters
		}

		newPlayer.floors[i] = newFloor
	}

	return newPlayer
}

func (player *Player) updateCompletedGameState() {
	for _, floor := range player.floors {
		if !floor.IsCompleted() {
			return
		}
	}

	player.State = playerstate.SUCCESS
}

func (player *Player) getAverageFloorTimeSeconds() int {
	sum := 0

	for _, floor := range player.floors {
		if !floor.bossFloor {
			sum += floor.totalTimeSpentSeconds
		}
	}

	floorsCountWithoutBossFloor := len(player.floors) - 1

	return sum / floorsCountWithoutBossFloor
}
