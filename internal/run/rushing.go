package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
)

type Rushing struct {
	baseRun
}

func (a Rushing) Name() string {
	return string(config.RushingRun)
}

func (a Rushing) BuildActions() []action.Action {
	if a.CharacterCfg.Companion.Enabled {
		if !a.CharacterCfg.Companion.Leader {
			a.builder.AddCharacterToParty(a.CharacterCfg.Companion.LeaderName, a.CharacterCfg.CharacterName)
			return []action.Action{
				a.getRushedAct1(),
				a.getRushedAct2(),
				a.getRushedAct3(),
				a.getRushedAct4(),
				a.getRushedAct5(),
			}
		} else {
			return []action.Action{
				a.rushAct1(),
				a.rushAct2(),
				a.rushAct3(),
				a.rushAct4(),
				a.rushAct5(),
			}
		}
	}

	return []action.Action{}
}

const (
	Moving RusherStatus = iota
	Waiting
	None
	GivingWPs
	ClearingDen
	FreeingCain
	RetrievingHammer
	KillingAndy
)

type RusherStatus int

var rusherStatuses = make(map[string]RusherStatus)

func (r Rushing) getRusherStatus(rusherName string) RusherStatus {
	if s, found := rusherStatuses[rusherName]; found {
		return s
	}
	return None
}

func (r Rushing) setRusherStatus(rusherName string, status RusherStatus) {
	rusherStatuses[rusherName] = status
}
