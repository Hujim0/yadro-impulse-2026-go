package engine

type GameFloor struct {
	monsterCount           int
	bossFloor              bool
	bossDefeated           bool
	totalTimeSpentSeconds  int
	lastTimeEnteredSeconds int
}

func (floor GameFloor) IsCompleted() bool {
	if floor.bossFloor {
		return floor.bossDefeated
	} else {
		return floor.monsterCount == 0
	}
}
