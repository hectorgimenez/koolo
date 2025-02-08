package character

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/context"
)

func BuildCharacter(ctx *context.Context) (context.Character, error) {
	bc := BaseCharacter{
		Context: ctx,
	}
	characterBuild := CharacterBuild{BaseCharacter: bc}

	if len(ctx.CharacterCfg.Game.Runs) > 0 && ctx.CharacterCfg.Game.Runs[0] == "leveling" {
		switch strings.ToLower(ctx.CharacterCfg.Character.Class) {
		case "sorceress_leveling_lightning":
			return SorceressLevelingLightning{CharacterBuild: characterBuild}, nil
		case "sorceress_leveling":
			return SorceressLeveling{CharacterBuild: characterBuild}, nil
		case "paladin":
			return PaladinLeveling{CharacterBuild: characterBuild}, nil
		}

		return nil, fmt.Errorf("leveling only available for sorceress and paladin")
	}

	switch strings.ToLower(ctx.CharacterCfg.Character.Class) {
	case "sorceress":
		return BlizzardSorceress{CharacterBuild: characterBuild}, nil
	case "nova":
		return NovaSorceress{CharacterBuild: characterBuild}, nil
	case "hydraorb":
		return HydraOrbSorceress{CharacterBuild: characterBuild}, nil
	case "lightsorc":
		return LightningSorceress{CharacterBuild: characterBuild}, nil
	case "hammerdin":
		return Hammerdin{CharacterBuild: characterBuild}, nil
	case "foh":
		return Foh{CharacterBuild: characterBuild}, nil
	case "trapsin":
		return Trapsin{CharacterBuild: characterBuild}, nil
	case "mosaic":
		return MosaicSin{CharacterBuild: characterBuild}, nil
	case "winddruid":
		return WindDruid{CharacterBuild: characterBuild}, nil
	case "javazon":
		return Javazon{CharacterBuild: characterBuild}, nil
	case "berserker":
		return &Berserker{CharacterBuild: characterBuild}, nil // Return a pointer to Berserker
	}

	return nil, fmt.Errorf("class %s not implemented", ctx.CharacterCfg.Character.Class)
}

type BaseCharacter struct {
	*context.Context
}

func (bc BaseCharacter) preBattleChecks(id data.UnitID, skipOnImmunities []stat.Resist) bool {
	monster, found := bc.Data.Monsters.FindByID(id)
	if !found {
		return false
	}
	for _, i := range skipOnImmunities {
		if monster.IsImmune(i) {
			bc.Logger.Info("Monster is immune! skipping", slog.String("immuneTo", string(i)))
			return false
		}
	}

	return true
}
