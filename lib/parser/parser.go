package parser

import (
	"dungeonGameLib/lib/engine/event"
	"fmt"
)

func ParseLine(line string) (event.GameEvent, error) {
	hours, minutes, seconds, eventTypeInt, playerId, eventParam := 0, 0, 0, 0, 0, 0

	n, err := fmt.Sscanf(line, "[%d:%d:%d] %d %d %d", &hours, &minutes, &seconds, &playerId, &eventTypeInt, &eventParam)

	if n < 5 {
		return event.GameEvent{}, fmt.Errorf("invalid log format (too few arguments): %q", line)
	}

	eventType, err := event.ParseEventId(eventTypeInt)

	if err != nil {
		return event.GameEvent{}, err
	}

	totalSeconds := hours*60*60 + minutes*60 + seconds

	return event.GameEvent{
		OccuredAtSecond: totalSeconds,
		PlayerId:        playerId,
		EventId:         eventType,
		ExtraParam:      eventParam,
	}, nil
}
