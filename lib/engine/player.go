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
)

var PlayerStateToString = map[PlayerState]string{
	IN_GAME: "IN_GAME",
	SUCCESS: "SUCCESS",
	FAIL:    "FAIL",
	DISQUAL: "DISQUAL",
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
}

func impossibleMove(player *Player, event *GameEvent, gameInstance *GameInstance) {
	gameInstance.OutputEventChan <- GameEvent{
		OccuredAtSecond: event.OccuredAtSecond,
		PlayerId:        event.PlayerId,
		EventId:         PlayerMakesImposibleMove,
		ExtraParam:      int(event.EventId),
	}
}

type playerEventHandler func(player *Player, event *GameEvent, gameInstance *GameInstance) error

var PlayerEventTypeToPlayerEventHandler = map[GameEventId]playerEventHandler{
	PlayerWentNext: func(player *Player, event *GameEvent, gameInstance *GameInstance) error {
		isLastFloor := len(gameInstance.floors)-1 == player.CurrentFloor

		if isLastFloor {
			impossibleMove(player, event, gameInstance)
			return fmt.Errorf("Cant go next floor: already at last")
		}

		player.CurrentFloor++
		return nil
	},
	PlayerWentPrev: func(player *Player, event *GameEvent, gameInstance *GameInstance) error {
		isFirstFloor := player.CurrentFloor == 0

		if isFirstFloor {
			impossibleMove(player, event, gameInstance)
			return fmt.Errorf("Cant go prev floor: already at first")
		}

		player.CurrentFloor--
		return nil
	},
	PlayerGotHealed: func(player *Player, event *GameEvent, gameInstance *GameInstance) error {
		hpToHeal := event.ExtraParam
		if player.Hp+hpToHeal > 100 {
			player.Hp = 100
		} else {
			player.Hp += hpToHeal
		}
		return nil
	},
	PlayerGotDamaged: func(player *Player, event *GameEvent, gameInstance *GameInstance) error {
		damage := event.ExtraParam

		if player.Hp > damage {
			player.Hp -= damage
			return nil
		}

		player.Hp = 0
		player.State = FAIL

		playerDeadEvent := GameEvent{
			OccuredAtSecond: event.OccuredAtSecond,
			PlayerId:        player.Id,
			EventId:         PlayerDead,
		}

		gameInstance.OutputEventChan <- playerDeadEvent

		return nil
	},

	PlayerKilled: func(player *Player, event *GameEvent, gameInstance *GameInstance) error {
		floorInstance := gameInstance.floors[player.CurrentFloor]

		if floorInstance.monsterCount <= 0 {
			impossibleMove(player, event, gameInstance)
			return fmt.Errorf("Player %d cant kill: no monsters on the floor %d", player.Id, player.CurrentFloor)
		}

		floorInstance.monsterCount--
		gameInstance.updateCompletedGameState()
		return nil
	},

	PlayerKilledBoss: func(player *Player, event *GameEvent, gameInstance *GameInstance) error {
		isLastFloor := len(gameInstance.floors)-1 == player.CurrentFloor
		if player.CurrentFloor >= len(gameInstance.floors) {
			return fmt.Errorf("Current floor %d is out of bounds", player.CurrentFloor)
		}

		isBossFloor := gameInstance.floors[player.CurrentFloor].bossFloor

		if !isLastFloor || !isBossFloor {
			impossibleMove(player, event, gameInstance)
			return fmt.Errorf("Player %d cant kill boss: not on the boss floor %d", player.Id, player.CurrentFloor)
		}

		gameInstance.floors[len(gameInstance.floors)-1].bossDefeated = true
		gameInstance.updateCompletedGameState()

		return nil
	},
	PlayerLeftDungeon: func(player *Player, event *GameEvent, gameInstance *GameInstance) error {
		player.ExitedDungeonAtSeconds = event.OccuredAtSecond

		if gameInstance.IsCompleted() {
			player.State = SUCCESS
		} else {
			player.State = FAIL
		}
		return nil
	},
	PlayerCannotContinue: func(player *Player, event *GameEvent, gameInstance *GameInstance) error {
		player.ExitedDungeonAtSeconds = event.OccuredAtSecond
		player.State = FAIL

		return nil
	},
}

func (player *Player) handleEvent(event *GameEvent, gameInstance *GameInstance) error {
	fun, ok := PlayerEventTypeToPlayerEventHandler[event.EventId]

	if !ok {
		fmt.Printf("Player %d doesnt know how to handle %s\n", player.Id, event.EventId.String())
	}

	return fun(player, event, gameInstance)
}

func createNewPlayer(currentTime int, id int) *Player {
	newPlayer := new(Player)
	newPlayer.Hp = 100
	newPlayer.State = IN_GAME
	newPlayer.EnteredDungeonAtSeconds = currentTime
	newPlayer.Id = id

	return newPlayer
}
