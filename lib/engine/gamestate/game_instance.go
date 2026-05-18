package gamestate

import (
	"cmp"
	"dungeonGameLib/lib/engine/event"
	"fmt"
	"slices"
)

type GameInstance struct {
	InputEventChan    chan event.GameEvent
	OutputEventChan   chan event.GameOutputEvent
	players           map[int]*Player
	registeredPlayers map[int]PlayerState
	openAtSeconds     int
	closesAtSeconds   int
	floors            int
	monsters          int
}

func (gameInstance *GameInstance) RunGameLoop() {
	for true {
		newEvent := <-gameInstance.InputEventChan

		if newEvent.OccuredAtSecond >= gameInstance.closesAtSeconds {
			gameInstance.CloseTheDungeon(newEvent.OccuredAtSecond)
			gameInstance.ImpossibleMove(&newEvent)
			break
		}

		gameInstance.handleNewEvent(newEvent)
	}
}

func (gameInstance *GameInstance) CloseTheDungeon(currentTimeSeconds int) {
	for _, player := range gameInstance.players {
		player.LeaveDungeon(currentTimeSeconds)

		if player.State == IN_GAME {
			player.State = FAIL
		}
	}
}

func (gameInstance *GameInstance) handleNewEvent(newEvent event.GameEvent) {

	eventMetaType, ok := event.EventTypeToMetaTypeEvent[newEvent.EventId]

	if !ok {
		gameInstance.OutputEventChan <- event.GameOutputEvent{
			Error: fmt.Errorf("Event metadata not found for the event: %d", newEvent.EventId),
		}
		return
	}

	switch eventMetaType {
	case event.PlayerMetaTypeEvent:
		player, ok := gameInstance.players[newEvent.PlayerId]

		if !ok {
			if _, disqualified := gameInstance.registeredPlayers[newEvent.PlayerId]; disqualified {
				gameInstance.OutputEventChan <- event.GameOutputEvent{
					Error: fmt.Errorf("Player %d is disqualified and cannot perform actions", newEvent.PlayerId),
				}
				return
			}

			gameInstance.OutputEventChan <- event.GameOutputEvent{
				Error: fmt.Errorf("Player with id %d not found", newEvent.PlayerId),
			}
			return
		}

		if player.State != IN_GAME && player.State != SUCCESS {
			gameInstance.OutputEventChan <- event.GameOutputEvent{
				Error: fmt.Errorf("Player state is wrong: %s but should be \"IN_GAME\"", player.State),
			}
			gameInstance.ImpossibleMove(&newEvent)
			return
		}

		outputEvent, err := player.HandleEvent(&newEvent, gameInstance)
		if err == nil {
			readOneMore := outputEvent != nil
			gameInstance.OutputEventChan <- event.GameOutputEvent{Event: newEvent, ReadOneMore: readOneMore}
		}

		if outputEvent != nil {
			gameInstance.OutputEventChan <- event.GameOutputEvent{Event: *outputEvent}
		}
	case event.DungeonMetaTypeEvent:
		handler, ok := DungeonEventTypeToPlayerEventHandler[newEvent.EventId]
		if !ok {
			gameInstance.OutputEventChan <- event.GameOutputEvent{
				Error: fmt.Errorf("Handler not found for event %d", newEvent.EventId),
			}
		}

		outputEvent, err := handler(&newEvent, gameInstance)
		if err == nil {
			readOneMore := outputEvent != nil
			gameInstance.OutputEventChan <- event.GameOutputEvent{Event: newEvent, ReadOneMore: readOneMore}
		}

		if outputEvent != nil {
			gameInstance.OutputEventChan <- event.GameOutputEvent{Event: *outputEvent}
		}
	}
}

func (gameInstance *GameInstance) ImpossibleMove(newEvent *event.GameEvent) {
	gameInstance.OutputEventChan <- event.GameOutputEvent{
		Event: event.GameEvent{
			OccuredAtSecond: newEvent.OccuredAtSecond,
			PlayerId:        newEvent.PlayerId,
			EventId:         event.PlayerMakesImposibleMove,
			ExtraParam:      int(newEvent.EventId),
		},
	}
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
		if state == DISQUAL {
			newDisqualifiedPlayer := CreateNewPlayer(0, id, 0, 0)
			newDisqualifiedPlayer.State = DISQUAL
			playersSlice[i] = PlayerAndIdPair{id, newDisqualifiedPlayer}
			i++
		} else {
			player, ok := gameInstance.players[id]

			if !ok {
				gameInstance.OutputEventChan <- event.GameOutputEvent{
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
	gameInstance.InputEventChan = make(chan event.GameEvent)
	gameInstance.OutputEventChan = make(chan event.GameOutputEvent)

	gameInstance.openAtSeconds = openAtSeconds
	gameInstance.closesAtSeconds = openAtSeconds + DurationHours*60*60
	gameInstance.players = make(map[int]*Player)
	gameInstance.registeredPlayers = make(map[int]PlayerState)
	gameInstance.floors = Floors
	gameInstance.monsters = Monsters

	return gameInstance, nil
}
