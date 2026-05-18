package gamestate

import (
	"dungeonGameLib/lib/engine/event"
	"fmt"
)

func ImpossibleMove(newEvent *event.GameEvent, gameInstance *GameInstance) {
	gameInstance.OutputEventChan <- event.GameOutputEvent{
		Event: event.GameEvent{
			OccuredAtSecond: newEvent.OccuredAtSecond,
			PlayerId:        newEvent.PlayerId,
			EventId:         event.PlayerMakesImposibleMove,
			ExtraParam:      int(newEvent.EventId),
		},
	}
}

type playerEventHandler func(eventPlayer *Player, event *event.GameEvent, gameInstance *GameInstance) (*event.GameEvent, error)

var PlayerEventTypeToPlayerEventHandler = map[event.GameEventId]playerEventHandler{
	event.PlayerWentNext: func(eventPlayer *Player, newEvent *event.GameEvent, gameInstance *GameInstance) (*event.GameEvent, error) {
		isLastFloor := len(eventPlayer.Floors)-1 == eventPlayer.CurrentFloor

		if isLastFloor {
			ImpossibleMove(newEvent, gameInstance)
			return nil, fmt.Errorf("Cant go next floor: already at last")
		}
		currentFloor := eventPlayer.Floors[eventPlayer.CurrentFloor]

		if !currentFloor.IsCompleted() {
			timeSpentOnCurrentFloorSeconds := newEvent.OccuredAtSecond - currentFloor.LastTimeEnteredSeconds
			currentFloor.TotalTimeSpentSeconds += timeSpentOnCurrentFloorSeconds
		}

		eventPlayer.CurrentFloor++

		nextFloor := eventPlayer.Floors[eventPlayer.CurrentFloor]
		nextFloor.LastTimeEnteredSeconds = newEvent.OccuredAtSecond

		return nil, nil
	},
	event.PlayerWentPrev: func(eventPlayer *Player, newEvent *event.GameEvent, gameInstance *GameInstance) (*event.GameEvent, error) {
		isFirstFloor := eventPlayer.CurrentFloor == 0

		if isFirstFloor {
			ImpossibleMove(newEvent, gameInstance)
			return nil, fmt.Errorf("Cant go prev floor: already at first")
		}
		currentFloor := eventPlayer.Floors[eventPlayer.CurrentFloor]

		if !currentFloor.IsCompleted() {
			timeSpentOnCurrentFloorSeconds := newEvent.OccuredAtSecond - currentFloor.LastTimeEnteredSeconds
			currentFloor.TotalTimeSpentSeconds += timeSpentOnCurrentFloorSeconds
		}

		eventPlayer.CurrentFloor--

		prevFloor := eventPlayer.Floors[eventPlayer.CurrentFloor]
		prevFloor.LastTimeEnteredSeconds = newEvent.OccuredAtSecond

		return nil, nil
	},
	event.PlayerGotHealed: func(eventPlayer *Player, newEvent *event.GameEvent, gameInstance *GameInstance) (*event.GameEvent, error) {
		hpToHeal := newEvent.ExtraParam
		if eventPlayer.Hp+hpToHeal > 100 {
			eventPlayer.Hp = 100
		} else {
			eventPlayer.Hp += hpToHeal
		}
		return nil, nil
	},
	event.PlayerGotDamaged: func(eventPlayer *Player, newEvent *event.GameEvent, gameInstance *GameInstance) (*event.GameEvent, error) {
		damage := newEvent.ExtraParam

		if eventPlayer.Hp > damage {
			eventPlayer.Hp -= damage
			return nil, nil
		}

		eventPlayer.Hp = 0
		eventPlayer.State = FAIL

		playerDeadEvent := event.GameEvent{
			OccuredAtSecond: newEvent.OccuredAtSecond,
			PlayerId:        eventPlayer.Id,
			EventId:         event.PlayerDead,
		}
		eventPlayer.ExitedDungeonAtSeconds = newEvent.OccuredAtSecond

		return &playerDeadEvent, nil
	},

	event.PlayerKilled: func(eventPlayer *Player, newEvent *event.GameEvent, gameInstance *GameInstance) (*event.GameEvent, error) {
		floorInstance := eventPlayer.Floors[eventPlayer.CurrentFloor]

		if floorInstance.MonsterCount <= 0 {
			ImpossibleMove(newEvent, gameInstance)
			return nil, fmt.Errorf("Player %d cant kill: no monsters on the floor %d", eventPlayer.Id, eventPlayer.CurrentFloor)
		}

		floorInstance.MonsterCount--

		if floorInstance.MonsterCount == 0 {
			timeSpentOnCurrentFloorSeconds := newEvent.OccuredAtSecond - floorInstance.LastTimeEnteredSeconds
			floorInstance.TotalTimeSpentSeconds += timeSpentOnCurrentFloorSeconds
		}

		eventPlayer.UpdateCompletedGameState()
		return nil, nil
	},

	event.PlayerKilledBoss: func(eventPlayer *Player, newEvent *event.GameEvent, gameInstance *GameInstance) (*event.GameEvent, error) {
		currentFloor := eventPlayer.Floors[eventPlayer.CurrentFloor]
		isBossFloor := currentFloor.BossFloor

		if !isBossFloor {
			ImpossibleMove(newEvent, gameInstance)
			return nil, fmt.Errorf("Player %d cant kill boss: not on the boss floor %d", eventPlayer.Id, eventPlayer.CurrentFloor)
		}

		currentFloor.BossDefeated = true
		currentFloor.TotalTimeSpentSeconds = newEvent.OccuredAtSecond - currentFloor.LastTimeEnteredSeconds
		eventPlayer.BossKillDurationSeconds = currentFloor.TotalTimeSpentSeconds
		eventPlayer.UpdateCompletedGameState()

		return nil, nil
	},
	event.PlayerLeftDungeon: func(eventPlayer *Player, newEvent *event.GameEvent, gameInstance *GameInstance) (*event.GameEvent, error) {

		if eventPlayer.State != SUCCESS {
			eventPlayer.State = FAIL

		}
		eventPlayer.ExitedDungeonAtSeconds = newEvent.OccuredAtSecond

		currentFloor := eventPlayer.Floors[eventPlayer.CurrentFloor]
		currentFloor.TotalTimeSpentSeconds = newEvent.OccuredAtSecond - currentFloor.LastTimeEnteredSeconds
		return nil, nil
	},
	event.PlayerCannotContinue: func(eventPlayer *Player, newEvent *event.GameEvent, gameInstance *GameInstance) (*event.GameEvent, error) {
		eventPlayer.State = FAIL
		eventPlayer.ExitedDungeonAtSeconds = newEvent.OccuredAtSecond

		currentFloor := eventPlayer.Floors[eventPlayer.CurrentFloor]
		currentFloor.TotalTimeSpentSeconds = newEvent.OccuredAtSecond - currentFloor.LastTimeEnteredSeconds
		return nil, nil
	},
	event.PlayerEnteredBoss: func(eventPlayer *Player, newEvent *event.GameEvent, gameInstance *GameInstance) (*event.GameEvent, error) {
		currentFloor := eventPlayer.Floors[eventPlayer.CurrentFloor]
		isBossFloor := currentFloor.BossFloor

		if !isBossFloor {
			ImpossibleMove(newEvent, gameInstance)
			return nil, fmt.Errorf("Player %d cant enter boss: not on the boss floor %d", eventPlayer.Id, eventPlayer.CurrentFloor)
		}

		currentFloor.LastTimeEnteredSeconds = newEvent.OccuredAtSecond

		return nil, nil
	},
}
