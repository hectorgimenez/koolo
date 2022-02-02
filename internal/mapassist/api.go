package mapassist

import (
	"encoding/json"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/health"
	"net/http"
)

const (
	genericData    = "/get_data"
	healthEndpoint = "/get_health"
)

type APIClient struct {
	hostName string
}

func NewAPIClient(hostName string) APIClient {
	return APIClient{hostName: hostName}
}

func (A APIClient) CurrentStatus() health.Status {
	r, err := http.Get(A.hostName + healthEndpoint)
	if err != nil {
		return health.Status{}
	}

	status := statusHttpResponse{}
	err = json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		return health.Status{}
	}
	if !status.Success {
		return health.Status{}
	}

	return health.Status{
		Life:    status.Life,
		MaxLife: status.MaxLife,
		Mana:    status.Mana,
		MaxMana: status.MaxMana,
		Merc: health.MercStatus{
			Alive:   status.Merc.Alive,
			Life:    status.Merc.Life,
			MaxLife: status.Merc.MaxLife,
		},
	}
}

func (A APIClient) GameData() data.Data {
	// TODO: Fix on MapAssist, first request always returns old data
	http.Get(A.hostName + genericData)
	r, err := http.Get(A.hostName + genericData)
	if err != nil {
		fmt.Println(err)
		// TODO: Handle error
	}

	d := gameDataHttpResponse{}
	err = json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		// TODO: Handle error
		return data.Data{}
	}
	if !d.Success {
		// TODO: Handle error
		return data.Data{}
	}

	corpse := data.Corpse{}
	// Match with current player
	for _, c := range d.Corpses {
		if c.Name == d.PlayerUnit.Name {
			corpse.Found = true
			corpse.IsHovered = c.IsHovered
			corpse.Position = data.Position{
				X: int(c.Position.X),
				Y: int(c.Position.Y),
			}
		}
	}

	monsters := map[data.NPCID]data.Monster{}
	for _, m := range d.Monsters {
		monsters[data.NPCID(m.Name)] = data.Monster{
			Name:      m.Name,
			IsHovered: m.IsHovered,
			Position: data.Position{
				X: int(m.Position.X),
				Y: int(m.Position.Y),
			},
		}
	}

	npcs := map[data.NPCID]data.NPC{}
	for _, npc := range d.NPCs {
		var positions []data.Position
		for _, p := range npc.Positions {
			positions = append(positions, data.Position{
				X: int(p.X),
				Y: int(p.Y),
			})
		}
		npcs[data.NPCID(npc.Name)] = data.NPC{
			Name:      npc.Name,
			Positions: positions,
		}
	}
	return data.Data{
		Area: data.Area(d.Area),
		AreaOrigin: data.Position{
			X: int(d.AreaOrigin.X),
			Y: int(d.AreaOrigin.Y),
		},
		Corpse:        corpse,
		Monsters:      monsters,
		NPCs:          npcs,
		CollisionGrid: d.CollisionGrid,
		PlayerUnit: data.PlayerUnit{
			Name: d.PlayerUnit.Name,
			Position: data.Position{
				X: int(d.PlayerUnit.Position.X),
				Y: int(d.PlayerUnit.Position.Y),
			},
		},
		OpenMenus: data.OpenMenus{
			Inventory:   d.MenuOpen.Inventory,
			NPCInteract: d.MenuOpen.NPCInteract,
			NPCShop:     d.MenuOpen.NPCShop,
			Stash:       d.MenuOpen.Stash,
			Waypoint:    d.MenuOpen.Waypoint,
		},
		Items: parseItems(d),
	}
}

func parseItems(d gameDataHttpResponse) data.Items {
	var potions []data.Potion
	for _, i := range d.Items {
		if i.Place == "Belt" {
			potionType := data.HealingPotion
			switch i.Name {
			case "Minor Mana Potion", "Light Mana Potion", "Mana Potion", "Greater Mana Potion", "Super Mana Potion":
				potionType = data.ManaPotion
			case "Rejuvenation Potion", "Full Rejuvenation Potion":
				potionType = data.RejuvenationPotion
			}

			potions = append(potions, data.Potion{
				Item: data.Item{
					Position: data.Position{
						X: int(i.Position.X),
						Y: int(i.Position.Y),
					},
					Name: i.Name,
				},
				Type: potionType,
			})
		}
	}

	var shop []data.Item
	for _, i := range d.Items {
		if i.Place == "Vendor" {
			shop = append(shop, data.Item{
				Position: data.Position{
					X: int(i.Position.X),
					Y: int(i.Position.Y),
				},
				Name: i.Name,
			})
		}
	}

	return data.Items{
		Belt: data.Belt{
			Potions: potions,
		},
		Shop: shop,
	}
}
