package action

import (
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (b *Builder) CheckKeyBindings(d game.Data) []skill.ID {
	return b.ch.CheckKeyBindings(d)
}
