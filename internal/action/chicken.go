package action

import (
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

func (b *Builder) ChickenOnMonsters(distance int, monsterIds []npc.ID) *Chain {
	return NewChain(func(d game.Data) []Action {
		for _, enemy := range d.Monsters.Enemies() {
			for _, m := range monsterIds {
				if m == enemy.Name && pather.DistanceFromMe(d, enemy.Position) <= distance {
					b.Logger.Info("Triggering chicken action")

					return nil
				}
			}
		}

		return []Action{}
	}, AbortOtherActionsIfNil(ReasonChicken))
}
