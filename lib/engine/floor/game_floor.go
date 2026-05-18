package floor

type GameFloor struct {
	MonsterCount           int
	BossFloor              bool
	BossDefeated           bool
	TotalTimeSpentSeconds  int
	LastTimeEnteredSeconds int
}

func (floor GameFloor) IsCompleted() bool {
	if floor.BossFloor {
		return floor.BossDefeated
	} else {
		return floor.MonsterCount == 0
	}
}
