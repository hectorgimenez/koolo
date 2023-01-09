package character

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"go.uber.org/zap"
	"strings"
)

func BuildCharacter(logger *zap.Logger) (action.Character, error) {
	bc := BaseCharacter{
		logger: logger,
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
			step.SyncStep(func(data game.Data) error {
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
	} else {
		// Add some delay when CTA is not enabled, because we can not control if weapon has been switched or not
		// so the game can still be on loading screen.
		steps = append(steps, step.SyncStep(func(data game.Data) error {
			helper.Sleep(4000)

			return nil
		}))
	}

	return
}

func (bc BaseCharacter) preBattleChecks(data game.Data, id game.UnitID, skipOnImmunities []stat.Resist) bool {
	monster, found := data.Monsters.FindByID(id)
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
