package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"time"
)

var charactersInParty map[string][]string

func (b Builder) AddCharacterToParty(supervisorName string, characterToBeRushed string) {
	// Check if the rusher exists in the map
	if foundCharacters, ok := charactersInParty[supervisorName]; ok {
		// Check if the value already exists in the slice
		for _, v := range foundCharacters {
			if v == characterToBeRushed {
				return // Do nothing, it's already in there
			}
		}
		// If character is not found in the slice, append it
		charactersInParty[supervisorName] = append(charactersInParty[supervisorName], characterToBeRushed)
	} else {
		// If rusher key does not exist, create a new slice with the character name
		charactersInParty[supervisorName] = []string{characterToBeRushed}
	}
}

func (b Builder) GetCharactersInParty(supervisorName string) []string {
	if r, found := charactersInParty[supervisorName]; found {
		return r
	}
	return []string{}
}

func (b Builder) RemoveCharacterFromParty(supervisorName string, characterToBeRushed string) {
	if foundCharacters, ok := charactersInParty[supervisorName]; ok {
		for i, v := range foundCharacters {
			if v == characterToBeRushed {
				// Remove the value from the slice
				charactersInParty[supervisorName] = append(foundCharacters[:i], foundCharacters[i+1:]...)
				return
			}
		}
	}
}

func (b *Builder) WaitForParty(supervisorName string) *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		partyMembers := b.GetCharactersInParty(supervisorName)

		actions = append(actions, b.OpenTP())
		everyonePresent := true
		for _, member := range partyMembers {
			if player, found := d.Roster.FindByName(member); !found {
				everyonePresent = false
			} else {
				if player.Area != d.PlayerUnit.Area {
					everyonePresent = false
				}
			}
		}

		if everyonePresent {
			return nil
		} else {
			actions = append(actions, NewStepChain(func(d game.Data) []step.Step {
				return []step.Step{step.Wait(time.Second)}
			}))

			return actions
		}
	}, RepeatUntilNoSteps())
}

func (b Builder) WaitForPartyToEnterPortal(supervisorName string) *Chain {
	partyMembers := b.GetCharactersInParty(supervisorName)
	var charactersEntered = make(map[string]bool)

	for _, member := range partyMembers {
		charactersEntered[member] = false
	}

	return NewChain(func(d game.Data) (actions []Action) {
		for _, member := range partyMembers {
			if _, found := charactersEntered[member]; found {
				if player, playerFound := d.Roster.FindByName(member); playerFound {
					charactersEntered[player.Name] = true
				}
			}
		}
		allMembersEntered := true

		for _, value := range charactersEntered {
			if !value {
				allMembersEntered = false
				break
			}
		}

		if allMembersEntered {
			return nil
		} else {
			actions = append(actions, NewStepChain(func(d game.Data) []step.Step {
				return []step.Step{step.Wait(time.Millisecond * 100)}
			}))

			return actions
		}
	}, RepeatUntilNoSteps())
}
