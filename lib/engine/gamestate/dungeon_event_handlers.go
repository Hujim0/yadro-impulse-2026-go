package gamestate

import (
	"dungeonGameLib/lib/engine/event"
	"fmt"
)

type DungeonEventHandler func(event *event.GameEvent, gameInstance *GameInstance) (*event.GameEvent, error)

var DungeonEventTypeToPlayerEventHandler = map[event.GameEventId]DungeonEventHandler{
	event.PlayerRegistered: func(event *event.GameEvent, gameInstance *GameInstance) (*event.GameEvent, error) {
		if event.OccuredAtSecond > gameInstance.closesAtSeconds {
			return nil, fmt.Errorf("the dungeon is closed! cant register.")
		}

		gameInstance.registeredPlayers[event.PlayerId] = REGISTERED
		return nil, nil
	},
	event.PlayerEntered: func(newEvent *event.GameEvent, gameInstance *GameInstance) (*event.GameEvent, error) {
		_, ok := gameInstance.registeredPlayers[newEvent.PlayerId]

		if !ok {
			gameInstance.registeredPlayers[newEvent.PlayerId] = DISQUAL
			disqualifiedEvent := event.GameEvent{
				OccuredAtSecond: newEvent.OccuredAtSecond,
				PlayerId:        newEvent.PlayerId,
				EventId:         event.PlayerDisqualified,
			}
			return &disqualifiedEvent, fmt.Errorf("Only registered players are allowed to participate in the challenge")
		}

		newPlayerInstance := CreateNewPlayer(newEvent.OccuredAtSecond, newEvent.PlayerId, gameInstance.floors, gameInstance.monsters)
		newPlayerInstance.Floors[0].LastTimeEnteredSeconds = newEvent.OccuredAtSecond

		gameInstance.players[newEvent.PlayerId] = newPlayerInstance
		gameInstance.registeredPlayers[newEvent.PlayerId] = IN_GAME

		return nil, nil
	},
}
