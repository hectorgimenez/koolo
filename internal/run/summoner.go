package run

//import (
//	"errors"
//	"github.com/hectorgimenez/koolo/internal/game"
//)
//
//type Summoner struct {
//	BaseRun
//}
//
//func NewSummoner(run BaseRun) Summoner {
//	return Summoner{
//		BaseRun: run,
//	}
//}
//
//func (p Summoner) Name() string {
//	return "Summoner"
//}
//
//func (p Summoner) Kill() error {
//	err := p.char.KillSummoner()
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (p Summoner) MoveToStartingPoint() error {
//	if game.Status().Area != game.ArcaneSanctuary {
//		if err := p.tm.WPTo(2, 8); err != nil {
//			return err
//		}
//	}
//
//	if game.Status().Area != game.ArcaneSanctuary {
//		return errors.New("error moving to Arcane Sanctuary")
//	}
//
//	p.char.Buff()
//	return nil
//}
//
//func (p Summoner) TravelToDestination() error {
//	npc, found := game.Status().NPCs[game.TheSummoner]
//	if !found {
//		return errors.New("The Summoner not found")
//	}
//
//	p.pf.MoveTo(npc.Positions[0].X, npc.Positions[0].Y, true)
//
//	return nil
//}
