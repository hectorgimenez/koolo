package character

import (
	"fmt"
	"strings"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"go.uber.org/zap"
)

func BuildCharacter(logger *zap.Logger) (action.Character, error) {
	bc := BaseCharacter{
		logger: logger,
	}

	// TODO: Refactor this, using a constant maybe
	if config.Config.Game.Runs[0] == "leveling" {
		return SorceressLeveling{BaseCharacter: bc}, nil
	}

	switch strings.ToLower(config.Config.Character.Class) {
	case "sorceress":
		return BlizzardSorceress{BaseCharacter: bc}, nil
	case "hammerdin":
		return Hammerdin{BaseCharacter: bc}, nil
	}

	return nil, fmt.Errorf("class %s not implemented", config.Config.Character.Class)
}

type BaseCharacter struct {
	logger *zap.Logger
}

func (bc BaseCharacter) buffCTA() (steps []step.Step) {
	if config.Config.Character.UseCTA {
		steps = append(steps,
			step.SwapWeapon(),
			step.SyncStep(func(d data.Data) error {
				hid.PressKey(config.Config.Bindings.CTABattleCommand)
				helper.Sleep(100)
				hid.Click(hid.RightButton)
				helper.Sleep(500)
				hid.PressKey(config.Config.Bindings.CTABattleOrders)
				helper.Sleep(100)
				hid.Click(hid.RightButton)
				helper.Sleep(100)

				return nil
			}),
			step.SwapWeapon(),
		)
	}

	return
}

func (bc BaseCharacter) preBattleChecks(d data.Data, id data.UnitID, skipOnImmunities []stat.Resist) bool {
	monster, found := d.Monsters.FindByID(id)
	if !found {
		return false
	}
	for _, i := range skipOnImmunities {
		if monster.IsImmune(i) {
			bc.logger.Info("Monster is immune! skipping", zap.String("immuneTo", string(i)))
			return false
		}
	}

	return true
}
