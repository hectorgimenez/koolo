package character

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
)

func BuildCharacter(logger *slog.Logger, container container.Container) (action.Character, error) {
	bc := BaseCharacter{
		logger:    logger,
		container: container,
	}

	if container.CharacterCfg.Game.Runs[0] == "leveling" {
		switch strings.ToLower(container.CharacterCfg.Character.Class) {
		case "sorceress_leveling_lightning":
			return SorceressLevelingLightning{BaseCharacter: bc}, nil
		case "sorceress_leveling":
			return SorceressLeveling{BaseCharacter: bc}, nil
		case "paladin":
			return PaladinLeveling{BaseCharacter: bc}, nil
		}

		return nil, fmt.Errorf("leveling only available for sorceress and paladin")
	}

	switch strings.ToLower(container.CharacterCfg.Character.Class) {
	case "sorceress":
		return BlizzardSorceress{BaseCharacter: bc}, nil
	case "lightning":
		return LightningSorceress{BaseCharacter: bc}, nil
	case "hammerdin":
		return Hammerdin{BaseCharacter: bc}, nil
	case "foh":
		return Foh{BaseCharacter: bc}, nil
	case "trapsin":
		return Trapsin{BaseCharacter: bc}, nil
	case "mosaic":
		return MosaicSin{BaseCharacter: bc}, nil
	case "winddruid":
		return WindDruid{BaseCharacter: bc}, nil
	case "javazon":
		return Javazon{BaseCharacter: bc}, nil
	case "berserker":
		return Berserker{BaseCharacter: bc}, nil
	}

	return nil, fmt.Errorf("class %s not implemented", container.CharacterCfg.Character.Class)
}

type BaseCharacter struct {
	logger    *slog.Logger
	container container.Container
}

func (bc BaseCharacter) preBattleChecks(d game.Data, id data.UnitID, skipOnImmunities []stat.Resist) bool {
	monster, found := d.Monsters.FindByID(id)
	if !found {
		return false
	}
	for _, i := range skipOnImmunities {
		if monster.IsImmune(i) {
			bc.logger.Info("Monster is immune! skipping", slog.String("immuneTo", string(i)))
			return false
		}
	}

	return true
}
