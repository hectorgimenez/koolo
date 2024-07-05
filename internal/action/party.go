package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"time"
)

var charactersInParty map[string][]string

func (b Builder) AddCharacterToParty(charName string, characterToBeAdded string) {
	if foundCharacters, ok := charactersInParty[charName]; ok {
		// Check if the value already exists in the slice
		for _, v := range foundCharacters {
			if v == characterToBeAdded {
				return // Do nothing, it's already in there
			}
		}
		// If character is not found in the slice, append it
		charactersInParty[charName] = append(charactersInParty[charName], characterToBeAdded)
	} else {
		charactersInParty[charName] = []string{characterToBeAdded}
	}
}

func (b Builder) GetCharactersInParty(charName string) []string {
	if r, found := charactersInParty[charName]; found {
		return r
	}
	return []string{}
}

func (b Builder) RemoveCharacterFromParty(charName string, characterToBeRemoved string) {
	if foundCharacters, ok := charactersInParty[charName]; ok {
		for i, v := range foundCharacters {
			if v == characterToBeRemoved {
				// Remove the value from the slice
				charactersInParty[charName] = append(foundCharacters[:i], foundCharacters[i+1:]...)
				return
			}
		}
	}
}

func (b *Builder) WaitForParty(charName string) *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		partyMembers := b.GetCharactersInParty(charName)

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

func (b Builder) WaitForPartyToEnterPortal(charName string) *Chain {
	partyMembers := b.GetCharactersInParty(charName)
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
