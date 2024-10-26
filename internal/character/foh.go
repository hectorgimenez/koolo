package character

import (
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/state"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	fohMinDistance    = 9
	fohMaxDistance    = 18
	hbMinDistance     = 6
	hbMaxDistance     = 12
	fohMaxAttacksLoop = 20 // Maximum attack attempts before resetting
)

type Foh struct {
	BaseCharacter
}

func (s Foh) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.Conviction, skill.HolyShield, skill.TomeOfTownPortal, skill.FistOfTheHeavens, skill.HolyBolt}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		s.Logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s Foh) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	ctx := context.Get()
	ctx.RefreshGameData()
	lastRefresh := time.Now()
	completedAttackLoops := 0

	for {
		// Limit refresh rate to 10 times per second for state checks
		if time.Since(lastRefresh) > time.Millisecond*100 {
			ctx.RefreshGameData()
			lastRefresh = time.Now()
		}

		id, found := monsterSelector(*s.Data)
		if !found {
			return nil
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		if completedAttackLoops >= fohMaxAttacksLoop {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found || monster.Stats[stat.Life] <= 0 {
			return nil
		}

		// Ensure our Conviction aura is active
		if !s.Data.PlayerUnit.States.HasState(state.Conviction) {
			if convictionKey, found := s.Data.KeyBindings.KeyBindingForSkill(skill.Conviction); found {
				ctx.HID.PressKeyBinding(convictionKey)
			}
		}

		// Setup attack options
		fohOpts := []step.AttackOption{
			step.RangedDistance(fohMinDistance, fohMaxDistance),
			step.EnsureAura(skill.Conviction),
		}

		hbOpts := []step.AttackOption{
			step.RangedDistance(hbMinDistance, hbMaxDistance),
			step.EnsureAura(skill.Conviction),
		}

		// Special handling for bosses and unique monsters
		if monster.Type == data.MonsterTypeUnique || monster.Type == data.MonsterTypeSuperUnique {
			// Cast FOH first
			if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.FistOfTheHeavens); found {
				ctx.HID.PressKeyBinding(kb)
				if err := step.PrimaryAttack(id, 1, true, fohOpts...); err == nil {
					// Then cast Holy Bolt
					if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.HolyBolt); found {
						ctx.HID.PressKeyBinding(kb)
						if err := step.PrimaryAttack(id, 3, true, hbOpts...); err == nil {
							completedAttackLoops++
						}
					}
				}
			}
			continue
		}

		// Check if monster is under the effect of our Conviction and still lightning immune
		monsterHasConviction := monster.States.HasState(state.Conviction)
		isLightningImmune := monster.IsImmune(stat.LightImmune)

		// Choose and select skill based on current state
		if isLightningImmune && monsterHasConviction {
			// Monster is still immune even while under our Conviction effect, use Holy Bolt
			if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.HolyBolt); found {
				ctx.HID.PressKeyBinding(kb)
				if err := step.PrimaryAttack(id, 1, true, hbOpts...); err == nil {
					completedAttackLoops++
				}
			}
		} else {
			// Either monster is not lightning immune or it might have immunity broken by Conviction, use FOH
			if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.FistOfTheHeavens); found {
				ctx.HID.PressKeyBinding(kb)
				if err := step.PrimaryAttack(id, 1, true, fohOpts...); err == nil {
					completedAttackLoops++
				}
			}
		}
	}
}

func (s Foh) BuffSkills() []skill.ID {
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.HolyShield); found {
		return []skill.ID{skill.HolyShield}
	}
	return []skill.ID{}
}

func (s Foh) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s Foh) killBoss(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found || m.Stats[stat.Life] <= 0 {
			return 0, false
		}
		return m.UnitID, true
	}, nil)
}

func (s Foh) KillCountess() error {
	return s.killBoss(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s Foh) KillAndariel() error {
	return s.killBoss(npc.Andariel, data.MonsterTypeUnique)
}

func (s Foh) KillSummoner() error {
	return s.killBoss(npc.Summoner, data.MonsterTypeUnique)
}

func (s Foh) KillDuriel() error {
	return s.killBoss(npc.Duriel, data.MonsterTypeUnique)
}

func (s Foh) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		var councilMembers []data.Monster
		for _, m := range d.Monsters {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				councilMembers = append(councilMembers, m)
			}
		}

		sort.Slice(councilMembers, func(i, j int) bool {
			return s.PathFinder.DistanceFromMe(councilMembers[i].Position) < s.PathFinder.DistanceFromMe(councilMembers[j].Position)
		})

		if len(councilMembers) > 0 {
			return councilMembers[0].UnitID, true
		}

		return 0, false
	}, nil)
}

func (s Foh) KillMephisto() error {
	return s.killBoss(npc.Mephisto, data.MonsterTypeUnique)
}

func (s Foh) KillIzual() error {
	return s.killBoss(npc.Izual, data.MonsterTypeUnique)
}

func (s Foh) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			s.Logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := s.Data.Monsters.FindOne(npc.Diablo, data.MonsterTypeUnique)
		if !found || diablo.Stats[stat.Life] <= 0 {
			if diabloFound {
				return nil
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		diabloFound = true
		s.Logger.Info("Diablo detected, attacking")

		return s.killBoss(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (s Foh) KillPindle() error {
	return s.killBoss(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s Foh) KillNihlathak() error {
	return s.killBoss(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s Foh) KillBaal() error {
	return s.killBoss(npc.BaalCrab, data.MonsterTypeUnique)
}
