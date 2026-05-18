package formatting

import (
	"fmt"
)

func SecondsToFormattedDate(seconds int) string {
	totalMinutes := seconds / 60
	hours := totalMinutes / 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, totalMinutes%60, seconds%60)
}
