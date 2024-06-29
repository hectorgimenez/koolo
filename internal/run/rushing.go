package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type Rushing struct {
	baseRun
}

func (a Rushing) Name() string {
	return string(config.RushingRun)
}

func (a Rushing) BuildActions() []action.Action {
	return []action.Action{
		a.rushAct1(),
		a.rushAct2(),
		a.rushAct3(),
		a.rushAct4(),
		a.rushAct5(),
	}
}

var charactersBeingRushed map[string][]string

func (a Rushing) AddCharacterToRush(rusherSupervisorName string, characterToBeRushed string) {
	// Check if the rusher exists in the map
	if foundCharacters, ok := charactersBeingRushed[rusherSupervisorName]; ok {
		// Check if the value already exists in the slice
		for _, v := range foundCharacters {
			if v == characterToBeRushed {
				return // Do nothing, it's already in there
			}
		}
		// If character is not found in the slice, append it
		charactersBeingRushed[rusherSupervisorName] = append(charactersBeingRushed[rusherSupervisorName], characterToBeRushed)
	} else {
		// If rusher key does not exist, create a new slice with the character name
		charactersBeingRushed[rusherSupervisorName] = []string{characterToBeRushed}
	}
}

func (a Rushing) GetCharactersBeingRushed(rusherSupervisorName string) []string {
	if r, found := charactersBeingRushed[rusherSupervisorName]; found {
		return r
	}
	return []string{}
}

func (a Rushing) RemoveCharacterFromRush(rusherSupervisorName string, characterToBeRushed string) {
	if foundCharacters, ok := charactersBeingRushed[rusherSupervisorName]; ok {
		for i, v := range foundCharacters {
			if v == characterToBeRushed {
				// Remove the value from the slice
				charactersBeingRushed[rusherSupervisorName] = append(foundCharacters[:i], foundCharacters[i+1:]...)
				return
			}
		}
	}
}

func (a Rushing) waitForParty() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		var actions []action.Action
		for {
			data := a.Container.Reader.GetData(false)

			var shouldContinue bool

			for _, c := range data.Roster {
				if c.Area.Area() == d.PlayerUnit.Area.Area() {
					shouldContinue = true
					break
				}
			}
			if shouldContinue == true {
				break
			} else {
				helper.Sleep(1000) // sleep 1
			}
		}

		return actions
	})
}
