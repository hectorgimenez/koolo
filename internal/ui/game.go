package ui

const (
	Normal    Difficulty = "normal"
	Nightmare Difficulty = "nightmare"
	Hell      Difficulty = "hell"
)

type Difficulty string

type GameManager struct {
}

func StartNewGame(difficulty Difficulty) error {
	return nil
}
