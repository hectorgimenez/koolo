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
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	fohMaxAttacksLoop = 20
	fohMinDistance    = 15
	fohMaxDistance    = 20
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
	completedAttackLoops := 0
	ctx := context.Get()

	for {
		ctx.PauseIfNotPriority()

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

		hbKey, holyBoltFound := s.Data.KeyBindings.KeyBindingForSkill(skill.HolyBolt)
		fohKey, fohFound := s.Data.KeyBindings.KeyBindingForSkill(skill.FistOfTheHeavens)
		convictionKey, convictionFound := s.Data.KeyBindings.KeyBindingForSkill(skill.Conviction)
		// Ensure Conviction is active
		if convictionFound {
			ctx.HID.PressKeyBinding(convictionKey)
			utils.Sleep(50)
		}

		if monster.Type == data.MonsterTypeUnique {
			s.attackBoss(monster.UnitID, hbKey, fohKey)
		} else {

			isLightningImmune := false
			if resistance, ok := monster.Stats[stat.LightningResist]; ok && resistance >= 100 {
				isLightningImmune = true
			}

			if monster.States.HasState(state.Conviction) && isLightningImmune && holyBoltFound {
				ctx.HID.PressKeyBinding(hbKey)
			} else if fohFound {
				ctx.HID.PressKeyBinding(fohKey)
			} else if holyBoltFound {
				ctx.HID.PressKeyBinding(hbKey)
			}

			step.PrimaryAttack(
				id,
				3,
				true,
				step.Distance(fohMinDistance, fohMaxDistance),
			)
		}

		completedAttackLoops++
		utils.Sleep(50)
	}
}
func (s Foh) attackBoss(bossID data.UnitID, hbKey, fohKey data.KeyBinding) {
	ctx := context.Get()
	ctx.PauseIfNotPriority()

	// Cast 1 FoH
	ctx.HID.PressKeyBinding(fohKey)
	utils.Sleep(100)
	step.PrimaryAttack(
		bossID,
		1,
		true,
		step.Distance(fohMinDistance, fohMaxDistance),
		step.EnsureAura(skill.Conviction),
	)

	// Cast 3 Holy Bolt
	ctx.HID.PressKeyBinding(hbKey)
	utils.Sleep(150)
	step.PrimaryAttack(
		bossID,
		3,
		true,
		step.Distance(fohMinDistance, fohMaxDistance),
		step.EnsureAura(skill.Conviction),
	)
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
