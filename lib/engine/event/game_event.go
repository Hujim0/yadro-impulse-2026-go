package event

import (
	"dungeonGameLib/lib/engine/formatting"
	"fmt"
	"strings"
)

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

	timeStamp := fmt.Sprintf("[%s] ", formatting.SecondsToFormattedDate(gameEvent.OccuredAtSecond))

	eventDescription := strings.Split(fmt.Sprintf(str+"\n", gameEvent.PlayerId, gameEvent.ExtraParam), "\n")[0]

	return timeStamp + formatting.RemoveExtraSprintfParam(eventDescription)
}
