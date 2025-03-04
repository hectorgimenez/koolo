package character

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	BasicMaxAttackLoops = 15
)

type BasicCharacter struct {
	BaseCharacter
}

func (b BasicCharacter) CheckKeyBindings() []skill.ID {
	return []skill.ID{}
}

func (b BasicCharacter) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	previousUnitID := 0
	attackSequenceLoop := 0

	for {
		id, found := monsterSelector(*b.Data)
		if !found {
			return nil
		}

		if !b.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		monster, found := b.Data.Monsters.FindByID(id)
		if !found {
			b.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		// If area is unreachable, or monster is dead, skip.
		if previousUnitID == int(id) {
			if monster.Stats[stat.Life] > 0 && s.Data.AreaData.IsWalkable(monster.Position) {
				b.PathFinder.RandomTeleport() // will walk if can't teleport
				utils.Sleep(400)
			} else {
				continue
			}
		}

		step.PrimaryAttack(
			monster.UnitID,
			6,
			false,
			step.Distance(0, 15),
		)

		if attackSequenceLoop >= BasicMaxAttackLoops {
			return nil
		}
		attackSequenceLoop++
		previousUnitID = int(id)
	}
}

func (b BasicCharacter) KillBossSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {

	for {
		id, found := monsterSelector(*b.Data)
		if !found {
			return nil
		}

		if !b.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		monster, found := b.Data.Monsters.FindByID(id)
		if !found {
			b.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		if monster.Stats[stat.Life] <= 0 {
			return nil
		}

		step.PrimaryAttack(
			monster.UnitID,
			6,
			false,
			step.Distance(0, 15),
		)
		action.Buff()

	}
}

func (b BasicCharacter) killMonsterByName(id npc.ID, monsterType data.MonsterType, skipOnImmunities []stat.Resist) error {
	return s.KillBossSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}

func (b BasicCharacter) BuffSkills() []skill.ID {
	return []skill.ID{}
}

func (b BasicCharacter) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (b BasicCharacter) KillAndariel() error {
	return s.killMonsterByName(npc.Andariel, data.MonsterTypeUnique, nil)
}

func (b BasicCharacter) KillDuriel() error {
	return s.killMonsterByName(npc.Duriel, data.MonsterTypeUnique, nil)
}

func (b BasicCharacter) KillMephisto() error {
	return s.killMonsterByName(npc.Mephisto, data.MonsterTypeUnique, nil)
}

func (b BasicCharacter) KillDiablo() error {
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

		return s.killMonsterByName(npc.Diablo, data.MonsterTypeUnique, nil)
	}
}

func (b BasicCharacter) KillBaal() error {
	return s.killMonsterByName(npc.BaalCrab, data.MonsterTypeUnique, nil)
}

func (b BasicCharacter) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, nil)
}

func (b BasicCharacter) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (b BasicCharacter) KillIzual() error {
	return s.killMonsterByName(npc.Izual, data.MonsterTypeUnique, nil)
}

func (b BasicCharacter) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		for _, m := range d.Monsters.Enemies() {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				return m.UnitID, true
			}
		}
		return 0, false
	}, nil)
}

func (b BasicCharacter) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, s.CharacterCfg.Game.Pindleskin.SkipOnImmunities)
}

func (b BasicCharacter) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, nil)
}
