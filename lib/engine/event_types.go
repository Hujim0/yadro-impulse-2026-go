package engine

import (
	"fmt"
	"strings"
)

type GameEventId int

const (
	PlayerRegistered GameEventId = iota + 1
	PlayerEntered
	PlayerKilled
	PlayerWentNext
	PlayerWentPrev
	PlayerEnteredBoss
	PlayerKilledBoss
	PlayerLeftDungeon
	PlayerCannotContinue
	PlayerGotHealed
	PlayerGotDamaged

	PlayerDisqualified       = GameEventId(31)
	PlayerDead               = GameEventId(32)
	PlayerMakesImposibleMove = GameEventId(33)
)

func (s GameEventId) String() string {
	str, ok := EventTypeToFormatString[s]

	if ok {
		return str
	}

	return "Unknown"
}

func ParseEventId(id int) (GameEventId, error) {
	eid := GameEventId(id)
	if _, ok := EventTypeToFormatString[eid]; !ok {
		return 0, fmt.Errorf("invalid EventId: %d", id)
	}
	return eid, nil
}

var EventTypeToFormatString = map[GameEventId]string{
	PlayerRegistered:     "Player [%d] registered",
	PlayerEntered:        "Player [%d] entered the dungeon",
	PlayerKilled:         "Player [%d] killed the monster",
	PlayerWentNext:       "Player [%d] went to the next floor",
	PlayerWentPrev:       "Player [%d] went to the previous floor",
	PlayerEnteredBoss:    "Player [%d] entered the boss’s floor",
	PlayerKilledBoss:     "Player [%d] killed the boss",
	PlayerLeftDungeon:    "Player [%d] left the dungeon",
	PlayerCannotContinue: "Player [%d] cannot continue due to [%d]",
	PlayerGotHealed:      "Player [%d] has restored [%d] of health",
	PlayerGotDamaged:     "Player [%d] recieved [%d] of damage",

	PlayerDisqualified:       "Player [%d] is disqualified",
	PlayerDead:               "Player [%d] is dead",
	PlayerMakesImposibleMove: "Player [%d] makes imposible move [%d]",
}

type GameEvent struct {
	OccuredAtSecond int
	PlayerId        int
	EventId         GameEventId
	ExtraParam      int
}

func (gameEvent GameEvent) String() string {
	str, ok := EventTypeToFormatString[gameEvent.EventId]

	if !ok {
		return "Unknown event"
	}

	timeStamp := fmt.Sprintf("[%s] ", secondsToFormattedDate(gameEvent.OccuredAtSecond))

	eventDescription := strings.Split(fmt.Sprintf(str+"\n", gameEvent.PlayerId, gameEvent.ExtraParam), "\n")[0]

	return timeStamp + eventDescription
}

type GameEventMetaType int

var EventTypeToMetaTypeEvent = map[GameEventId]GameEventMetaType{
	PlayerRegistered:     DungeonMetaTypeEvent,
	PlayerEntered:        DungeonMetaTypeEvent,
	PlayerKilled:         PlayerMetaTypeEvent,
	PlayerWentNext:       PlayerMetaTypeEvent,
	PlayerWentPrev:       PlayerMetaTypeEvent,
	PlayerEnteredBoss:    PlayerMetaTypeEvent,
	PlayerKilledBoss:     PlayerMetaTypeEvent,
	PlayerLeftDungeon:    PlayerMetaTypeEvent,
	PlayerCannotContinue: PlayerMetaTypeEvent,
	PlayerGotHealed:      PlayerMetaTypeEvent,
	PlayerGotDamaged:     PlayerMetaTypeEvent,

	PlayerDisqualified:       DungeonMetaTypeEvent,
	PlayerDead:               DungeonMetaTypeEvent,
	PlayerMakesImposibleMove: DungeonMetaTypeEvent,
}

const (
	PlayerMetaTypeEvent GameEventMetaType = iota
	DungeonMetaTypeEvent
)
