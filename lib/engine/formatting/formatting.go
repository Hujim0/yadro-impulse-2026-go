package formatting

import (
	"fmt"
	"strings"
)

func SecondsToFormattedDate(seconds int) string {
	totalMinutes := seconds / 60
	hours := totalMinutes / 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, totalMinutes%60, seconds%60)
}

func RemoveExtraSprintfParam(input string) string {
	return strings.Split(input+"\n", "\n")[0]
}
