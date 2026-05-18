package event

type GameOutputEvent struct {
	Event       GameEvent
	Error       error
	ReadOneMore bool
}
