package engine

import (
	"fmt"
)

type PlayerState int

const (
	IN_GAME PlayerState = iota
	SUCCESS
	FAIL
	DISQUAL
	REGISTERED
)

var PlayerStateToString = map[PlayerState]string{
	IN_GAME:    "IN_GAME",
	SUCCESS:    "SUCCESS",
	FAIL:       "FAIL",
	DISQUAL:    "DISQUAL",
	REGISTERED: "REGISTERED",
}

func (s PlayerState) String() string {
	str, ok := PlayerStateToString[s]

	if ok {
		return str
	}

	return "Unknown"
}

type Player struct {
	State                                    PlayerState
	Id                                       int
	Hp                                       int
	CurrentFloor                             int
	SpentInDungeonSeconds                    int
	AverageMonstersFloorClearDurationSeconds int
	BossKillDurationSeconds                  int
	EnteredDungeonAtSeconds                  int
	ExitedDungeonAtSeconds                   int
	floors                                   []*GameFloor
}

func impossibleMove(player *Player, event *GameEvent, gameInstance *GameInstance) {
	gameInstance.OutputEventChan <- GameEvent{
		OccuredAtSecond: event.OccuredAtSecond,
		PlayerId:        event.PlayerId,
		EventId:         PlayerMakesImposibleMove,
		ExtraParam:      int(event.EventId),
	}
}

type playerEventHandler func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error)

var PlayerEventTypeToPlayerEventHandler = map[GameEventId]playerEventHandler{
	PlayerWentNext: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		isLastFloor := len(player.floors)-1 == player.CurrentFloor

		if isLastFloor {
			impossibleMove(player, event, gameInstance)
			return nil, fmt.Errorf("Cant go next floor: already at last")
		}

		player.CurrentFloor++
		return nil, nil
	},
	PlayerWentPrev: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		isFirstFloor := player.CurrentFloor == 0

		if isFirstFloor {
			impossibleMove(player, event, gameInstance)
			return nil, fmt.Errorf("Cant go prev floor: already at first")
		}

		player.CurrentFloor--
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
		player.State = FAIL

		playerDeadEvent := GameEvent{
			OccuredAtSecond: event.OccuredAtSecond,
			PlayerId:        player.Id,
			EventId:         PlayerDead,
		}

		return &playerDeadEvent, nil
	},

	PlayerKilled: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		floorInstance := player.floors[player.CurrentFloor]

		if floorInstance.monsterCount <= 0 {
			impossibleMove(player, event, gameInstance)
			return nil, fmt.Errorf("Player %d cant kill: no monsters on the floor %d", player.Id, player.CurrentFloor)
		}

		floorInstance.monsterCount--
		player.updateCompletedGameState()
		return nil, nil
	},

	PlayerKilledBoss: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		isLastFloor := len(player.floors)-1 == player.CurrentFloor
		if player.CurrentFloor >= len(player.floors) {
			return nil, fmt.Errorf("Current floor %d is out of bounds", player.CurrentFloor)
		}

		isBossFloor := player.floors[player.CurrentFloor].bossFloor

		if !isLastFloor || !isBossFloor {
			impossibleMove(player, event, gameInstance)
			return nil, fmt.Errorf("Player %d cant kill boss: not on the boss floor %d", player.Id, player.CurrentFloor)
		}

		player.floors[len(player.floors)-1].bossDefeated = true
		player.updateCompletedGameState()

		return nil, nil
	},
	PlayerLeftDungeon: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		player.ExitedDungeonAtSeconds = event.OccuredAtSecond

		if player.State != SUCCESS {
			player.State = FAIL
		}
		return nil, nil
	},
	PlayerCannotContinue: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		player.ExitedDungeonAtSeconds = event.OccuredAtSecond
		player.State = FAIL

		return nil, nil
	},
	PlayerEnteredBoss: func(player *Player, event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
		isLastFloor := len(player.floors)-1 == player.CurrentFloor
		if player.CurrentFloor >= len(player.floors) {
			return nil, fmt.Errorf("Current floor %d is out of bounds", player.CurrentFloor)
		}

		isBossFloor := player.floors[player.CurrentFloor].bossFloor

		if !isLastFloor || !isBossFloor {
			impossibleMove(player, event, gameInstance)
			return nil, fmt.Errorf("Player %d cant enter boss: not on the boss floor %d", player.Id, player.CurrentFloor)
		}

		return nil, nil
	},
}

func (player *Player) handleEvent(event *GameEvent, gameInstance *GameInstance) (*GameEvent, error) {
	fun, ok := PlayerEventTypeToPlayerEventHandler[event.EventId]

	if !ok {
		fmt.Printf("Player %d doesnt know how to handle %s\n", player.Id, event.EventId.String())
		return nil, fmt.Errorf("handler not found for event %s", event.EventId.String())
	}

	return fun(player, event, gameInstance)
}

func createNewPlayer(currentTime int, id int, Floors int, Monsters int) *Player {
	newPlayer := new(Player)
	newPlayer.Hp = 100
	newPlayer.State = IN_GAME
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

	player.State = SUCCESS
}
