package event

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
