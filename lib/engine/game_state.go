package engine

import (
	"cmp"
	playerstate "dungeonGameLib/lib/engine/player_state"
	"fmt"
	"slices"
)

type GameInstance struct {
	InputEventChan    chan GameEvent
	OutputEventChan   chan GameOutput
	players           map[int]*Player
	registeredPlayers map[int]playerstate.PlayerState
	openAtSeconds     int
	closesAtSeconds   int
	floors            int
	monsters          int
}

func (gameInstance *GameInstance) RunGameLoop() {
	for true {
		newEvent := <-gameInstance.InputEventChan

		eventMetaType, ok := EventTypeToMetaTypeEvent[newEvent.EventId]

		if !ok {
			gameInstance.OutputEventChan <- GameOutput{
				Error: fmt.Errorf("Event metadata not found for the event: %d", newEvent.EventId),
			}
			continue
		}

		switch eventMetaType {
		case PlayerMetaTypeEvent:
			player, ok := gameInstance.players[newEvent.PlayerId]

			if !ok {
				if _, disqualified := gameInstance.registeredPlayers[newEvent.PlayerId]; disqualified {
					gameInstance.OutputEventChan <- GameOutput{
						Error: fmt.Errorf("Player %d is disqualified and cannot perform actions", newEvent.PlayerId),
					}
					continue
				}

				gameInstance.OutputEventChan <- GameOutput{
					Error: fmt.Errorf("Player with id %d not found", newEvent.PlayerId),
				}
				continue
			}

			if player.State != playerstate.IN_GAME && player.State != playerstate.SUCCESS {
				gameInstance.OutputEventChan <- GameOutput{
					Error: fmt.Errorf("Player state is wrong: %s but should be \"IN_GAME\"", player.State),
				}
				impossibleMove(&newEvent, gameInstance)
				continue
			}

			outputEvent, err := player.handleEvent(&newEvent, gameInstance)
			if err == nil {
				readOneMore := outputEvent != nil
				gameInstance.OutputEventChan <- GameOutput{Event: newEvent, ReadOneMore: readOneMore}
			}

			if outputEvent != nil {
				gameInstance.OutputEventChan <- GameOutput{Event: *outputEvent}
			}
		case DungeonMetaTypeEvent:
			handler, ok := DungeonEventTypeToPlayerEventHandler[newEvent.EventId]
			if !ok {
				gameInstance.OutputEventChan <- GameOutput{
					Error: fmt.Errorf("Handler not found for event %d", newEvent.EventId),
				}
			}

			outputEvent, err := handler(&newEvent, gameInstance)
			if err == nil {
				readOneMore := outputEvent != nil
				gameInstance.OutputEventChan <- GameOutput{Event: newEvent, ReadOneMore: readOneMore}
			}

			if outputEvent != nil {
				gameInstance.OutputEventChan <- GameOutput{Event: *outputEvent}
			}
		}
	}
}

type DungeonEventHandler func(event *GameEvent, gameInstance *GameInstance) (*GameEvent, error)

var DungeonEventTypeToPlayerEventHandler = map[GameEventId]DungeonEventHandler{
	PlayerRegistered: func(event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		if event.OccuredAtSecond > gameInstance.closesAtSeconds {
			return nil, fmt.Errorf("the dungeon is closed! cant register.")
		}

		gameInstance.registeredPlayers[event.PlayerId] = playerstate.REGISTERED
		return nil, nil
	},
	PlayerEntered: func(event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		_, ok := gameInstance.registeredPlayers[event.PlayerId]

		if !ok {
			gameInstance.registeredPlayers[event.PlayerId] = playerstate.DISQUAL
			disqualifiedEvent := GameEvent{
				OccuredAtSecond: event.OccuredAtSecond,
				PlayerId:        event.PlayerId,
				EventId:         PlayerDisqualified,
			}
			return &disqualifiedEvent, fmt.Errorf("Only registered players are allowed to participate in the challenge")
		}

		newPlayerInstance := createNewPlayer(event.OccuredAtSecond, event.PlayerId, gameInstance.floors, gameInstance.monsters)
		newPlayerInstance.floors[0].lastTimeEnteredSeconds = event.OccuredAtSecond

		gameInstance.players[event.PlayerId] = newPlayerInstance
		gameInstance.registeredPlayers[event.PlayerId] = playerstate.IN_GAME

		return nil, nil
	},
}

func CreateGameInstance(Floors int, Monsters int, OpenAt string, DurationHours int) (*GameInstance, error) {
	if Floors <= 0 {
		return nil, fmt.Errorf("Floor count should be greater than 0")
	}
	if Monsters < 0 {
		return nil, fmt.Errorf("Monsters count should be greater or equal to 0")
	}
	if DurationHours <= 0 {
		return nil, fmt.Errorf("DurationHours should be greater than 0")
	}

	seconds, minutes, hours := 0, 0, 0
	n, err := fmt.Sscanf(OpenAt, "%d:%d:%d", &hours, &minutes, &seconds)

	if err != nil || n != 3 {
		return nil, fmt.Errorf("Failed to parse dungeon open time")
	}

	openAtSeconds := seconds + minutes*60 + hours*60*60

	gameInstance := new(GameInstance)
	gameInstance.InputEventChan = make(chan GameEvent)
	gameInstance.OutputEventChan = make(chan GameOutput)

	gameInstance.openAtSeconds = openAtSeconds
	gameInstance.closesAtSeconds = openAtSeconds + DurationHours*60*60
	gameInstance.players = make(map[int]*Player)
	gameInstance.registeredPlayers = make(map[int]playerstate.PlayerState)
	gameInstance.floors = Floors
	gameInstance.monsters = Monsters

	return gameInstance, nil
}

type PlayerAndIdPair = struct {
	Id     int
	Player *Player
}

func (gameInstance *GameInstance) CompilePlayerData() []PlayerAndIdPair {
	playerCount := len(gameInstance.registeredPlayers)

	playersSlice := make([]PlayerAndIdPair, playerCount)
	i := 0
	for id, state := range gameInstance.registeredPlayers {
		if state == playerstate.DISQUAL {
			newDisqualifiedPlayer := createNewPlayer(0, id, 0, 0)
			newDisqualifiedPlayer.State = playerstate.DISQUAL
			playersSlice[i] = PlayerAndIdPair{id, newDisqualifiedPlayer}
			i++
		} else {
			player, ok := gameInstance.players[id]

			if !ok {
				gameInstance.OutputEventChan <- GameOutput{
					Error: fmt.Errorf("Registered player with id %d not found in player instances!", id),
				}
			}
			playersSlice[i] = PlayerAndIdPair{id, player}
			i++
		}
	}

	slices.SortFunc(playersSlice, func(a, b PlayerAndIdPair) int {
		return cmp.Compare(a.Id, b.Id)
	})

	return playersSlice
}
