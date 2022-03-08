package game

import (
	"context"
	"github.com/hectorgimenez/koolo/api"
)

var GRPCClient api.MapAssistApiClient

func Status(ctx context.Context) Data {
	d, err := GRPCClient.GetData(ctx, &api.R{})
	if err != nil {
		panic(err)
	}

	corpse := Corpse{}
	// Match with current player
	for _, c := range d.GetCorpses() {
		if c.GetName() == d.GetPlayerUnit().GetName() {
			corpse.Found = true
			corpse.IsHovered = c.GetHovered()
			corpse.Position = Position{
				X: int(c.GetPosition().GetX()),
				Y: int(c.GetPosition().GetY()),
			}
		}
	}

	stats := map[Stat]int{}
	for _, stat := range d.GetPlayerUnit().GetStats() {
		stats[Stat(stat.GetName())] = int(stat.GetValue())
	}

	skills := map[Skill]int{}
	for _, skill := range d.GetPlayerUnit().GetSkills() {
		skills[Skill(skill.GetName())] = int(skill.GetPoints())
	}

	var monsters []Monster
	for _, m := range d.GetMonsters() {
		var immunities []Resist
		for _, resist := range m.GetImmunities() {
			immunities = append(immunities, Resist(resist))
		}
		monsters = append(monsters, Monster{
			Name:      m.GetName(),
			IsHovered: m.GetHovered(),
			Position: Position{
				X: int(m.GetPosition().GetX()),
				Y: int(m.GetPosition().GetY()),
			},
			Immunities: immunities,
		})
	}

	npcs := map[NPCID]NPC{}
	for _, npc := range d.GetNpcs() {
		var positions []Position
		for _, p := range npc.GetPositions() {
			positions = append(positions, Position{
				X: int(p.GetX()),
				Y: int(p.GetY()),
			})
		}
		npcs[NPCID(npc.GetName())] = NPC{
			Name:      npc.GetName(),
			Positions: positions,
		}
	}

	var objects []Object
	for _, o := range d.GetObjects() {
		objects = append(objects, Object{
			Name:       o.GetName(),
			IsHovered:  o.GetHovered(),
			Selectable: o.GetSelectable(),
			Position: Position{
				X: int(o.GetPosition().GetX()),
				Y: int(o.GetPosition().GetY()),
			},
		})
	}

	var levels []Level
	for _, lv := range d.GetAdjacentLevels() {
		levels = append(levels, Level{
			Area: Area(lv.GetArea()),
			Position: Position{
				X: int(lv.GetPositions()[0].GetX()),
				Y: int(lv.GetPositions()[0].GetY()),
			},
		})
	}

	var pois []PointOfInterest
	for _, poi := range d.GetPointsOfInterest() {
		pois = append(pois, PointOfInterest{
			Name: poi.GetName(),
			Position: Position{
				X: int(poi.GetPosition().GetX()),
				Y: int(poi.GetPosition().GetY()),
			},
		})
	}

	var cg [][]bool
	for _, g := range d.GetCollisionGrid() {
		var c []bool
		for _, walkable := range g.GetWalkable() {
			c = append(c, walkable)
		}
		cg = append(cg, c)
	}

	return Data{
		Health: Health{
			Life:    int(d.GetStatus().GetLife()),
			MaxLife: int(d.GetStatus().GetMaxLife()),
			Mana:    int(d.GetStatus().GetMana()),
			MaxMana: int(d.GetStatus().GetMaxMana()),
			Merc: MercStatus{
				Alive:   d.GetStatus().GetMercAlive(),
				Life:    int(d.GetStatus().GetMercLife()),
				MaxLife: int(d.GetStatus().GetMercMaxLife()),
			},
		},
		Area: Area(d.GetArea()),
		AreaOrigin: Position{
			X: int(d.AreaOrigin.GetX()),
			Y: int(d.AreaOrigin.GetY()),
		},
		Corpse:        corpse,
		Monsters:      monsters,
		CollisionGrid: cg,
		PlayerUnit: PlayerUnit{
			Name: d.GetPlayerUnit().GetName(),
			Position: Position{
				X: int(d.GetPlayerUnit().GetPosition().GetX()),
				Y: int(d.GetPlayerUnit().GetPosition().GetY()),
			},
			Stats:  stats,
			Skills: skills,
			Class:  Class(d.GetPlayerUnit().GetClass()),
		},
		NPCs:             npcs,
		Items:            organizeItem(d.GetItems()),
		Objects:          objects,
		AdjacentLevels:   levels,
		PointsOfInterest: pois,
		OpenMenus: OpenMenus{
			Inventory:   d.MenuOpen.GetInventory(),
			NPCInteract: d.MenuOpen.GetNpcInteract(),
			NPCShop:     d.MenuOpen.GetNpcShop(),
			Stash:       d.MenuOpen.GetStash(),
			Waypoint:    d.MenuOpen.GetWaypoint(),
		},
	}
}

func organizeItem(items []*api.Item) Items {
	var potions []Potion
	for _, i := range items {
		if i.Place == "Belt" {
			potionType := HealingPotion
			switch i.Name {
			case "MinorManaPotion", "LightManaPotion", "ManaPotion", "GreaterManaPotion", "SuperManaPotion":
				potionType = ManaPotion
			case "RejuvenationPotion", "FullRejuvenationPotion":
				potionType = RejuvenationPotion
			}

			potions = append(potions, Potion{
				Item: Item{
					Position: Position{
						X: int(i.Position.X),
						Y: int(i.Position.Y),
					},
					Name: i.Name,
				},
				Type: potionType,
			})
		}
	}

	var shop []Item
	var ground []Item
	var inventory []Item
	var stash []Item
	for _, i := range items {
		stats := map[Stat]int{}
		for _, s := range i.Stats {
			stats[Stat(s.GetName())] = int(s.Value)
		}
		item := Item{
			ID: int(i.GetId()),
			Position: Position{
				X: int(i.GetPosition().GetX()),
				Y: int(i.GetPosition().GetY()),
			},
			Name:       i.GetName(),
			Quality:    Quality(i.GetQuality()),
			Ethereal:   i.GetEthereal(),
			IsHovered:  i.GetHovered(),
			Stats:      stats,
			Identified: i.GetIdentified(),
		}
		switch i.Place {
		case "Vendor":
			shop = append(shop, item)
		case "Ground":
			ground = append(ground, item)
		case "Inventory":
			inventory = append(inventory, item)
		case "Stash":
			stash = append(stash, item)
		}
	}

	return Items{
		Belt: Belt{
			Potions: potions,
		},
		Shop:      shop,
		Ground:    ground,
		Inventory: inventory,
	}
}
