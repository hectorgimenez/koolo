package character

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

func BuildCharacter(logger *slog.Logger, cfg *config.CharacterCfg, data *game.Data, pf *pather.PathFinder) (context.Character, error) {
	bc := BaseCharacter{
		logger: logger,
		data:   data,
		cfg:    cfg,
		pf:     pf,
	}

	ctx := context.Get()

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

	switch strings.ToLower(cfg.Character.Class) {
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
		return Berserker{BaseCharacter: bc}, nil
	}

	return nil, fmt.Errorf("class %s not implemented", cfg.Character.Class)
}

type BaseCharacter struct {
	logger *slog.Logger
	data   *game.Data
	cfg    *config.CharacterCfg
	pf     *pather.PathFinder
}

func (bc BaseCharacter) preBattleChecks(id data.UnitID, skipOnImmunities []stat.Resist) bool {
	monster, found := bc.data.Monsters.FindByID(id)
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
