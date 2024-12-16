package character

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/context"
)

func BuildCharacter(ctx *context.Context) (context.Character, error) {
	bc := BaseCharacter{
		Context: ctx,
	}

	if len(ctx.CharacterCfg.Game.Runs) > 0 && ctx.CharacterCfg.Game.Runs[0] == "leveling" {
		switch strings.ToLower(ctx.CharacterCfg.Character.Class) {
		case "sorceress_leveling_lightning":
			return SorceressLevelingLightning{BaseCharacter: bc}, nil
		case "sorceress_leveling":
			return SorceressLeveling{BaseCharacter: bc}, nil
		case "paladin":
			return PaladinLeveling{BaseCharacter: bc}, nil
		}

		return nil, fmt.Errorf("leveling only available for sorceress and paladin")
	}

	switch strings.ToLower(ctx.CharacterCfg.Character.Class) {
	case "sorceress":
		return BlizzardSorceress{BaseCharacter: bc}, nil
	case "nova":
		return NovaSorceress{BaseCharacter: bc}, nil
	case "hydraorb":
		return HydraOrbSorceress{BaseCharacter: bc}, nil
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
		return &Berserker{BaseCharacter: bc}, nil // Return a pointer to Berserker
	}

	return nil, fmt.Errorf("class %s not implemented", ctx.CharacterCfg.Character.Class)
}

type BaseCharacter struct {
	*context.Context
}

func (bc BaseCharacter) preBattleChecks(id data.UnitID) bool {
	monster, found := bc.Data.Monsters.FindByID(id)
	if !found {
		return false
	}
	for _, i := range bc.CharacterCfg.Runtime.ImmunityFilter {
		if monster.IsImmune(i) {
			bc.Logger.Info("Monster is immune! skipping", slog.String("immuneTo", string(i)))
			return false
		}
	}

	return true
}
