package parser

import (
	"dungeonGameLib/lib/engine"
	"fmt"
)

func ParseLine(line string) (engine.GameEvent, error) {
	hours, minutes, seconds, eventTypeInt, playerId, eventParam := 0, 0, 0, 0, 0, 0

	n, err := fmt.Sscanf(line, "[%d:%d:%d] %d %d %d", &hours, &minutes, &seconds, &playerId, &eventTypeInt, &eventParam)

	if n < 5 {
		return engine.GameEvent{}, fmt.Errorf("invalid log format (too few arguments): %q", line)
	}

	eventType, err := engine.ParseEventId(eventTypeInt)

	if err != nil {
		return engine.GameEvent{}, err
	}

	totalSeconds := hours*60*60 + minutes*60 + seconds

	return engine.GameEvent{
		OccuredAtSecond: totalSeconds,
		PlayerId:        playerId,
		EventId:         eventType,
		ExtraParam:      eventParam,
	}, nil
}
