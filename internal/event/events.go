package event

const (
	ExitedGame        Event = "ExitedGame"
	SafeAreaAbandoned Event = "SafeAreaAbandoned"
	SafeAreaEntered   Event = "SafeAreaEntered"
)

type Event string
