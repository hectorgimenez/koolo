package action

import (
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (b *Builder) ChickenOnMonsters(monsterIds []npc.ID) *Chain {
	return NewChain(func(d game.Data) []Action {
		for _, enemy := range d.Monsters.Enemies() {
			for _, m := range monsterIds {
				if m == enemy.Name {
					b.Logger.Info("Triggering chicken action")

					return nil
				}
			}
		}

		return []Action{}
	}, AbortOtherActionsIfNil())
}
