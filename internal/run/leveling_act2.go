package run

import "github.com/hectorgimenez/koolo/internal/action"

func (a Leveling) act2() (actions []action.Action) {
	actions = append(actions,
		a.radament(),
		a.findHoradricCube(),
	)

	return
}

func (a Leveling) radament() action.Action {
	return nil
}

func (a Leveling) findHoradricCube() action.Action {
	return nil
}
