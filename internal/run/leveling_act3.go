package run

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/ui"
)

func (a Leveling) act3() action.Action {
	running := false
	return action.NewChain(func(d data.Data) []action.Action {
		if running || d.PlayerUnit.Area != area.KurastDocks {
			return nil
		}

		quests := a.builder.GetCompletedQuests(3)

		running = true
		_, willFound := d.Items.Find("KhalimsWill", item.LocationInventory, item.LocationStash)
		if willFound {
			return a.openMephistoStairs()
		}

		if quests[2] {
			actions := Mephisto{baseRun: a.baseRun}.BuildActions()
			return append(actions, a.builder.ItemPickup(true, 25),
				a.builder.InteractObject(object.HellGate, func(d data.Data) bool {
					return d.PlayerUnit.Area == area.ThePandemoniumFortress
				}),
			)
		}

		// Find KhalimsEye
		_, found := d.Items.Find("KhalimsEye", item.LocationInventory, item.LocationStash)
		if found {
			a.logger.Info("KhalimsEye found, skipping quest")
		} else {
			a.logger.Info("KhalimsEye not found, starting quest")
			return a.findKhalimsEye()
		}

		// Find KhalimsBrain
		_, found = d.Items.Find("KhalimsBrain", item.LocationInventory, item.LocationStash)
		if found {
			a.logger.Info("KhalimsBrain found, skipping quest")
		} else {
			a.logger.Info("KhalimsBrain not found, starting quest")
			return a.findKhalimsBrain()
		}

		// Find KhalimsHeart
		_, found = d.Items.Find("KhalimsHeart", item.LocationInventory, item.LocationStash)
		if found {
			a.logger.Info("KhalimsHeart found, skipping quest")
		} else {
			a.logger.Info("KhalimsHeart not found, starting quest")
			return a.findKhalimsHeart()
		}

		// Trav
		return a.openMephistoStairs()
	})
}

func (a Leveling) findKhalimsEye() []action.Action {
	return []action.Action{
		a.builder.WayPoint(area.SpiderForest),
		a.builder.Buff(),
		a.builder.MoveToArea(area.SpiderCavern),
		a.builder.Buff(),
		a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
			a.logger.Info("Khalm Chest found, moving to that room")
			chest, found := d.Objects.FindOne(object.KhalimChest3)

			return chest.Position, found
		}),
		a.builder.ClearAreaAroundPlayer(15),
		a.builder.InteractObject(object.KhalimChest3, func(d data.Data) bool {
			chest, _ := d.Objects.FindOne(object.KhalimChest3)
			return !chest.Selectable
		}),
		a.builder.ItemPickup(true, 10),
	}
}

func (a Leveling) findKhalimsBrain() []action.Action {
	return []action.Action{
		a.builder.WayPoint(area.FlayerJungle),
		a.builder.Buff(),
		a.builder.MoveToArea(area.FlayerDungeonLevel1),
		a.builder.Buff(),
		a.builder.MoveToArea(area.FlayerDungeonLevel2),
		a.builder.Buff(),
		a.builder.MoveToArea(area.FlayerDungeonLevel3),
		a.builder.Buff(),
		a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
			a.logger.Info("Khalm Chest found, moving to that room")
			chest, found := d.Objects.FindOne(object.KhalimChest2)

			return chest.Position, found
		}),
		//a.builder.ClearAreaAroundPlayer(15),
		a.builder.InteractObject(object.KhalimChest2, func(d data.Data) bool {
			chest, _ := d.Objects.FindOne(object.KhalimChest2)
			return !chest.Selectable
		}),
		a.builder.ItemPickup(true, 10),
	}
}

func (a Leveling) findKhalimsHeart() []action.Action {
	return []action.Action{
		a.builder.WayPoint(area.KurastBazaar),
		a.builder.Buff(),
		a.builder.MoveToArea(area.SewersLevel1Act3),
		a.builder.Buff(),
		a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
			for _, l := range d.AdjacentLevels {
				if l.Area == area.SewersLevel2Act3 {
					return l.Position, true
				}
			}
			return data.Position{}, false
		}),
		a.builder.ClearAreaAroundPlayer(10),
		a.builder.InteractObject(object.Act3SewerStairsToLevel3, func(d data.Data) bool {
			o, _ := d.Objects.FindOne(object.Act3SewerStairsToLevel3)

			return !o.Selectable
		}),
		a.builder.Wait(time.Second * 3),
		a.builder.InteractObject(object.Act3SewerStairs, func(d data.Data) bool {
			return d.PlayerUnit.Area == area.SewersLevel2Act3
		}),
		a.builder.Buff(),
		a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
			a.logger.Info("Khalm Chest found, moving to that room")
			chest, found := d.Objects.FindOne(object.KhalimChest1)

			return chest.Position, found
		}),
		a.builder.ClearAreaAroundPlayer(15),
		a.builder.InteractObject(object.KhalimChest1, func(d data.Data) bool {
			chest, _ := d.Objects.FindOne(object.KhalimChest1)
			return !chest.Selectable
		}),
		a.builder.ItemPickup(true, 10),
	}
}

func (a Leveling) openMephistoStairs() []action.Action {
	actions := Council{baseRun: a.baseRun}.BuildActions()

	return append(actions,
		a.builder.ItemPickup(true, 40),
		a.builder.ReturnTown(),
		action.NewChain(func(d data.Data) []action.Action {
			eye, _ := d.Items.Find("KhalimsEye", item.LocationInventory, item.LocationStash)
			brain, _ := d.Items.Find("KhalimsBrain", item.LocationInventory, item.LocationStash)
			heart, _ := d.Items.Find("KhalimsHeart", item.LocationInventory, item.LocationStash)
			flail, _ := d.Items.Find("KhalimsFlail", item.LocationInventory, item.LocationStash)

			return []action.Action{
				a.builder.CubeAddItems(eye, brain, heart, flail),
				a.builder.CubeTransmute(),
			}
		}),

		a.builder.UsePortalInTown(),
		action.NewStepChain(func(d data.Data) []step.Step {
			return []step.Step{
				// Let's asume we don't have secondary weapon, so we swap to it and equip Khalim's Will
				step.SyncStep(func(d data.Data) error {
					khalimsWill, found := d.Items.Find("KhalimsWill")
					if !found {
						return nil
					}

					hid.PressKey(config.Config.Bindings.SwapWeapon)
					helper.Sleep(1000)
					hid.PressKey(config.Config.Bindings.OpenInventory)
					screenPos := ui.GetScreenCoordsForItem(khalimsWill)
					hid.MovePointer(screenPos.X, screenPos.Y)
					hid.KeyDown("shift")
					helper.Sleep(500)
					hid.Click(hid.LeftButton)
					helper.Sleep(500)
					hid.KeyUp("shift")
					helper.Sleep(300)
					hid.PressKey("esc")
					return nil
				}),
			}
		}),
		a.builder.InteractObject(object.CompellingOrb,
			func(d data.Data) bool {
				o, _ := d.Objects.FindOne(object.CompellingOrb)

				return !o.Selectable
			},
			step.SyncStep(func(d data.Data) error {
				helper.Sleep(1000)
				hid.PressKey(config.Config.Bindings.SwapWeapon)
				return nil
			})),
		a.builder.Wait(time.Second*12),
		a.builder.InteractObject(object.StairSR, func(d data.Data) bool {
			return d.PlayerUnit.Area == area.DuranceOfHateLevel1
		}),
		a.builder.MoveToArea(area.DuranceOfHateLevel2),
		a.builder.DiscoverWaypoint(),
	)
}
