package data

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	genericData = "/get_data"
	hostName    = "http://localhost:1111"
)

func Status() Data {
	r, err := http.Get(hostName + genericData)
	if err != nil {
		fmt.Println(err)
		// TODO: Handle error
	}

	d := gameDataHttpResponse{}
	err = json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		// TODO: Handle error
	}
	if !d.Success {
		// TODO: Handle error
	}

	corpse := Corpse{}
	// Match with current player
	for _, c := range d.Corpses {
		if c.Name == d.PlayerUnit.Name {
			corpse.Found = true
			corpse.IsHovered = c.IsHovered
			corpse.Position = Position{
				X: int(c.Position.X),
				Y: int(c.Position.Y),
			}
		}
	}

	monsters := map[NPCID]Monster{}
	for _, m := range d.Monsters {
		monsters[NPCID(m.Name)] = Monster{
			Name:      m.Name,
			IsHovered: m.IsHovered,
			Position: Position{
				X: int(m.Position.X),
				Y: int(m.Position.Y),
			},
		}
	}

	npcs := map[NPCID]NPC{}
	for _, npc := range d.NPCs {
		var positions []Position
		for _, p := range npc.Positions {
			positions = append(positions, Position{
				X: int(p.X),
				Y: int(p.Y),
			})
		}
		npcs[NPCID(npc.Name)] = NPC{
			Name:      npc.Name,
			Positions: positions,
		}
	}

	stats := map[Stat]int{}
	for _, stat := range d.PlayerUnit.Stats {
		stats[Stat(stat.Stat)] = stat.Value
	}

	var objects []Object
	for _, o := range d.Objects {
		objects = append(objects, Object{
			Name:       o.Name,
			IsHovered:  o.IsHovered,
			Selectable: o.Selectable,
			Position: Position{
				X: int(o.Position.X),
				Y: int(o.Position.Y),
			},
		})
	}

	return Data{
		Health: Health{
			Life:    d.Status.Life,
			MaxLife: d.Status.MaxLife,
			Mana:    d.Status.Mana,
			MaxMana: d.Status.MaxMana,
			Merc: MercStatus{
				Alive:   d.Status.Merc.Alive,
				Life:    d.Status.Merc.Life,
				MaxLife: d.Status.Merc.MaxLife,
			},
		},
		Area: Area(d.Area),
		AreaOrigin: Position{
			X: int(d.AreaOrigin.X),
			Y: int(d.AreaOrigin.Y),
		},
		Corpse:        corpse,
		Monsters:      monsters,
		NPCs:          npcs,
		CollisionGrid: d.CollisionGrid,
		PlayerUnit: PlayerUnit{
			Name: d.PlayerUnit.Name,
			Position: Position{
				X: int(d.PlayerUnit.Position.X),
				Y: int(d.PlayerUnit.Position.Y),
			},
			Stats: stats,
			Class: Class(d.PlayerUnit.PlayerClass),
		},
		OpenMenus: OpenMenus{
			Inventory:   d.MenuOpen.Inventory,
			NPCInteract: d.MenuOpen.NPCInteract,
			NPCShop:     d.MenuOpen.NPCShop,
			Stash:       d.MenuOpen.Stash,
			Waypoint:    d.MenuOpen.Waypoint,
		},
		Items:   parseItems(d),
		Objects: objects,
	}
}

func parseItems(d gameDataHttpResponse) Items {
	var potions []Potion
	for _, i := range d.Items {
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
	for _, i := range d.Items {
		stats := map[Stat]int{}
		for _, s := range i.Stats {
			stats[Stat(s.Stat)] = s.Value
		}
		item := Item{
			Position: Position{
				X: int(i.Position.X),
				Y: int(i.Position.Y),
			},
			Name:      i.Name,
			Quality:   Quality(i.Quality),
			Ethereal:  i.Ethereal,
			IsHovered: i.IsHovered,
			Stats:     stats,
		}
		switch i.Place {
		case "Vendor":
			shop = append(shop, item)
		case "Ground":
			ground = append(ground, item)
		case "Inventory":
			inventory = append(inventory, item)
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
