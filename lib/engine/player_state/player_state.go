package playerstate

type PlayerState int

const (
	IN_GAME PlayerState = iota
	SUCCESS
	FAIL
	DISQUAL
	REGISTERED
)

var PlayerStateToString = map[PlayerState]string{
	IN_GAME:    "IN_GAME",
	SUCCESS:    "SUCCESS",
	FAIL:       "FAIL",
	DISQUAL:    "DISQUAL",
	REGISTERED: "REGISTERED",
}

func (s PlayerState) String() string {
	str, ok := PlayerStateToString[s]

	if ok {
		return str
	}

	return "Unknown"
}
