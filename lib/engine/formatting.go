package engine

import "fmt"

func secondsToFormattedDate(seconds int) string {
	totalMinutes := seconds / 60
	hours := totalMinutes / 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, totalMinutes%60, seconds%60)
}

func (player *Player) String() string {
	spentInDungeonSeconds := player.ExitedDungeonAtSeconds - player.EnteredDungeonAtSeconds
	averageMonstersFloorClearDurationSeconds := player.getAverageFloorTimeSeconds()

	if spentInDungeonSeconds < 0 {
		spentInDungeonSeconds = 0
	}

	// [SUCCESS] 1 [00:24:00, 00:05:00, 00:11:00] HP:35
	return fmt.Sprintf("[%s] %d [%s, %s, %s] HP:%d",
		player.State,
		player.Id,
		secondsToFormattedDate(spentInDungeonSeconds),
		secondsToFormattedDate(averageMonstersFloorClearDurationSeconds),
		secondsToFormattedDate(player.BossKillDurationSeconds),
		player.Hp,
	)
}
