package character

import (
	"fmt"
	"strings"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
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

	if config.Config.Game.Runs[0] == "leveling" {
		switch strings.ToLower(config.Config.Character.Class) {
		case "sorceress":
			return SorceressLeveling{BaseCharacter: bc}, nil
		case "paladin":
			return PaladinLeveling{BaseCharacter: bc}, nil
		}

		return nil, fmt.Errorf("leveling only available for sorceress and paladin")
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

func (bc BaseCharacter) buffCTA(d data.Data) (steps []step.Step) {
	ctaFound := false
	for _, itm := range d.Items.ByLocation(item.LocationEquipped) {
		if itm.Stats[stat.NumSockets].Value == 5 && itm.Stats[stat.ReplenishLife].Value == 12 && itm.Stats[stat.NonClassSkill].Value > 0 && itm.Stats[stat.PreventMonsterHeal].Value > 0 {
			ctaFound = true
			bc.logger.Debug("CTA found: swapping weapon and casting Battle Command / Battle Orders")
			break
		}
	}

	if ctaFound && config.Config.Bindings.CTABattleCommand != "" && config.Config.Bindings.CTABattleOrders != "" {
		steps = append(steps,
			step.SwapWeapon(),
			step.SyncStep(func(d data.Data) error {
				hid.PressKey(config.Config.Bindings.CTABattleCommand)
				helper.Sleep(100)
				hid.Click(hid.RightButton)
				helper.Sleep(300)
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
