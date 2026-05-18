package engine

import (
	playerstate "dungeonGameLib/lib/engine/player_state"
	"fmt"
)

func impossibleMove(event *GameEvent, gameInstance *GameInstance) {
	gameInstance.OutputEventChan <- GameOutput{
		Event: GameEvent{
			OccuredAtSecond: event.OccuredAtSecond,
			PlayerId:        event.PlayerId,
			EventId:         PlayerMakesImposibleMove,
			ExtraParam:      int(event.EventId),
		},
	}
}

type playerEventHandler func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error)

var PlayerEventTypeToPlayerEventHandler = map[GameEventId]playerEventHandler{
	PlayerWentNext: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		isLastFloor := len(player.floors)-1 == player.CurrentFloor

		if isLastFloor {
			impossibleMove(event, gameInstance)
			return nil, fmt.Errorf("Cant go next floor: already at last")
		}
		currentFloor := player.floors[player.CurrentFloor]

		if !currentFloor.IsCompleted() {
			timeSpentOnCurrentFloorSeconds := event.OccuredAtSecond - currentFloor.lastTimeEnteredSeconds
			currentFloor.totalTimeSpentSeconds += timeSpentOnCurrentFloorSeconds
		}

		player.CurrentFloor++

		nextFloor := player.floors[player.CurrentFloor]
		nextFloor.lastTimeEnteredSeconds = event.OccuredAtSecond

		return nil, nil
	},
	PlayerWentPrev: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		isFirstFloor := player.CurrentFloor == 0

		if isFirstFloor {
			impossibleMove(event, gameInstance)
			return nil, fmt.Errorf("Cant go prev floor: already at first")
		}
		currentFloor := player.floors[player.CurrentFloor]

		if !currentFloor.IsCompleted() {
			timeSpentOnCurrentFloorSeconds := event.OccuredAtSecond - currentFloor.lastTimeEnteredSeconds
			currentFloor.totalTimeSpentSeconds += timeSpentOnCurrentFloorSeconds
		}

		player.CurrentFloor--

		prevFloor := player.floors[player.CurrentFloor]
		prevFloor.lastTimeEnteredSeconds = event.OccuredAtSecond

		return nil, nil
	},
	PlayerGotHealed: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		hpToHeal := event.ExtraParam
		if player.Hp+hpToHeal > 100 {
			player.Hp = 100
		} else {
			player.Hp += hpToHeal
		}
		return nil, nil
	},
	PlayerGotDamaged: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		damage := event.ExtraParam

		if player.Hp > damage {
			player.Hp -= damage
			return nil, nil
		}

		player.Hp = 0
		player.State = playerstate.FAIL

		playerDeadEvent := GameEvent{
			OccuredAtSecond: event.OccuredAtSecond,
			PlayerId:        player.Id,
			EventId:         PlayerDead,
		}
		player.ExitedDungeonAtSeconds = event.OccuredAtSecond

		return &playerDeadEvent, nil
	},

	PlayerKilled: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		floorInstance := player.floors[player.CurrentFloor]

		if floorInstance.monsterCount <= 0 {
			impossibleMove(event, gameInstance)
			return nil, fmt.Errorf("Player %d cant kill: no monsters on the floor %d", player.Id, player.CurrentFloor)
		}

		floorInstance.monsterCount--

		if floorInstance.monsterCount == 0 {
			timeSpentOnCurrentFloorSeconds := event.OccuredAtSecond - floorInstance.lastTimeEnteredSeconds
			floorInstance.totalTimeSpentSeconds += timeSpentOnCurrentFloorSeconds
		}

		player.updateCompletedGameState()
		return nil, nil
	},

	PlayerKilledBoss: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		currentFloor := player.floors[player.CurrentFloor]
		isBossFloor := currentFloor.bossFloor

		if !isBossFloor {
			impossibleMove(event, gameInstance)
			return nil, fmt.Errorf("Player %d cant kill boss: not on the boss floor %d", player.Id, player.CurrentFloor)
		}

		currentFloor.bossDefeated = true
		currentFloor.totalTimeSpentSeconds = event.OccuredAtSecond - currentFloor.lastTimeEnteredSeconds
		player.BossKillDurationSeconds = currentFloor.totalTimeSpentSeconds
		player.updateCompletedGameState()

		return nil, nil
	},
	PlayerLeftDungeon: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {

		if player.State != playerstate.SUCCESS {
			player.State = playerstate.FAIL

		}
		player.ExitedDungeonAtSeconds = event.OccuredAtSecond

		currentFloor := player.floors[player.CurrentFloor]
		currentFloor.totalTimeSpentSeconds = event.OccuredAtSecond - currentFloor.lastTimeEnteredSeconds
		return nil, nil
	},
	PlayerCannotContinue: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		player.State = playerstate.FAIL
		player.ExitedDungeonAtSeconds = event.OccuredAtSecond

		currentFloor := player.floors[player.CurrentFloor]
		currentFloor.totalTimeSpentSeconds = event.OccuredAtSecond - currentFloor.lastTimeEnteredSeconds
		return nil, nil
	},
	PlayerEnteredBoss: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		currentFloor := player.floors[player.CurrentFloor]
		isBossFloor := currentFloor.bossFloor

		if !isBossFloor {
			impossibleMove(event, gameInstance)
			return nil, fmt.Errorf("Player %d cant enter boss: not on the boss floor %d", player.Id, player.CurrentFloor)
		}

		currentFloor.lastTimeEnteredSeconds = event.OccuredAtSecond

		return nil, nil
	},
}

func (player *Player) handleEvent(event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
	fun, ok := PlayerEventTypeToPlayerEventHandler[event.EventId]

	if !ok {
		return nil, fmt.Errorf("Player %d handler not found for event %s", player.Id, event.EventId.String())
	}

	return fun(player, event, gameInstance)
}
