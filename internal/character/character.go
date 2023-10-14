package character

import (
	"fmt"
	"strings"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
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
	case "lightning":
		return LightningSorceress{BaseCharacter: bc}, nil
	case "hammerdin":
		return Hammerdin{BaseCharacter: bc}, nil
	}

	return nil, fmt.Errorf("class %s not implemented", config.Config.Character.Class)
}

type BaseCharacter struct {
	logger *zap.Logger
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
