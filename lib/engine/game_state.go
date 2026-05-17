package engine

import (
	"fmt"
)

type GameInstance struct {
	InputEventChan    chan GameEvent
	OutputEventChan   chan GameEvent
	players           map[int]*Player
	floors            []*GameFloor
	registeredPlayers map[int]bool
	openAtSeconds     int
	closesAtSeconds   int
	isCompleted       bool
}

func (instance *GameInstance) IsCompleted() bool {
	return instance.isCompleted
}

func (instance *GameInstance) updateCompletedGameState() {
	for _, floor := range instance.floors {
		if !floor.IsCompleted() {
			instance.isCompleted = false
			return
		}
	}

	instance.isCompleted = true
}

func (gameInstance *GameInstance) RunGameLoop() {
	for true {
		newEvent := <-gameInstance.InputEventChan

		eventMetaType, ok := EventTypeToMetaTypeEvent[newEvent.EventId]

		if !ok {
			fmt.Println("Event metadata not found for the event:", newEvent.EventId)
			continue
		}

		switch eventMetaType {
		case PlayerMetaTypeEvent:
			player, ok := gameInstance.players[newEvent.PlayerId]

			if !ok {
				fmt.Println("Player with id", newEvent.PlayerId, "not found.")
				continue
			}

			if player.State != IN_GAME {
				fmt.Println("Player state is wrong:", player.State, "but should be \"IN_GAME\"")
				impossibleMove(player, &newEvent, gameInstance)
				continue
			}

			err := player.handleEvent(&newEvent, gameInstance)
			if err == nil {
				gameInstance.OutputEventChan <- newEvent
			}
		case DungeonMetaTypeEvent:
			handler, ok := DungeonEventTypeToPlayerEventHandler[newEvent.EventId]
			if !ok {
				fmt.Println("Handler not found for event", newEvent.EventId)
			}

			err := handler(&newEvent, gameInstance)
			if err == nil {
				gameInstance.OutputEventChan <- newEvent
			}
		}
	}
}

type DungeonEventHandler func(event *GameEvent, gameInstance *GameInstance) error

var DungeonEventTypeToPlayerEventHandler = map[GameEventId]DungeonEventHandler{
	PlayerRegistered: func(event *GameEvent, gameInstance *GameInstance) error {
		if event.OccuredAtSecond > gameInstance.closesAtSeconds {
			return fmt.Errorf("the dungeon is closed! cant register.")
		}

		gameInstance.registeredPlayers[event.PlayerId] = true
		return nil
	},
	PlayerEntered: func(event *GameEvent, gameInstance *GameInstance) error {
		_, ok := gameInstance.registeredPlayers[event.PlayerId]

		if !ok {
			gameInstance.OutputEventChan <- GameEvent{
				OccuredAtSecond: event.OccuredAtSecond,
				PlayerId:        event.PlayerId,
				EventId:         PlayerDisqualified,
			}
			return fmt.Errorf("Only registered players are allowed to participate in the challenge")
		}

		gameInstance.players[event.PlayerId] = createNewPlayer(event.OccuredAtSecond, event.PlayerId)
		return nil
	},
}

func CreateGameInstance(Floors int, Monsters int, OpenAt string, DurationHours int) *GameInstance {
	if Floors <= 0 {
		fmt.Println("Floor count should be greater than 0")
		return nil
	}
	if Monsters < 0 {
		fmt.Println("Monsters count should be greater or equal to 0")
		return nil
	}
	if DurationHours <= 0 {
		fmt.Println("DurationHours should be greater than 0")
		return nil
	}

	seconds, minutes, hours := 0, 0, 0
	n, err := fmt.Sscanf(OpenAt, "%d:%d:%d", &hours, &minutes, &seconds)

	if err != nil || n != 3 {
		fmt.Println("Failed to parse dungeon open time")
		return nil
	}

	openAtSeconds := seconds + minutes*60 + hours*60*60

	gameInstance := new(GameInstance)
	gameInstance.InputEventChan = make(chan GameEvent)
	gameInstance.OutputEventChan = make(chan GameEvent)

	gameInstance.openAtSeconds = openAtSeconds
	gameInstance.closesAtSeconds = openAtSeconds + DurationHours*60*60
	gameInstance.isCompleted = false
	gameInstance.players = make(map[int]*Player)
	gameInstance.registeredPlayers = make(map[int]bool)
	gameInstance.floors = make([]*GameFloor, Floors)

	for i := range gameInstance.floors {
		newFloor := new(GameFloor)
		isBossFloor := i == Floors-1
		if isBossFloor {
			newFloor.bossFloor = true
		} else {
			newFloor.monsterCount = Monsters
		}

		gameInstance.floors[i] = newFloor
	}

	return gameInstance
}
