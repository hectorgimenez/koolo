package mapassist

import (
	"encoding/json"
	"errors"
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

func (A APIClient) CurrentStatus() (health.Status, error) {
	r, err := http.Get(A.hostName + healthEndpoint)
	if err != nil {
		return health.Status{}, err
	}

	status := statusHttpResponse{}
	err = json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		return health.Status{}, err
	}
	if !status.Success {
		return health.Status{}, errors.New("error fetching MapAssist data from API")
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
	}, nil
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
			corpse.Position = data.Position{
				X: int(c.Position.X),
				Y: int(c.Position.Y),
			}
		}
	}

	monsters := map[data.NPCID]data.Monster{}
	for _, m := range d.Monsters {
		monsters[data.NPCID(m.Name)] = data.Monster{
			Name: m.Name,
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
	}
}

func (A APIClient) Inventory() data.Inventory {
	http.Get(A.hostName + genericData)
	r, _ := http.Get(A.hostName + genericData)

	d := gameDataHttpResponse{}
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		// TODO: Handle error
		return data.Inventory{}
	}
	if !d.Success {
		// TODO: Handle error
		return data.Inventory{}
	}

	var potions []data.Potion
	for _, i := range d.Items {
		if i.Place == "INBELT" {
			potionType := data.HealingPotion
			switch i.Name {
			case "Minor Mana Potion", "Light Mana Potion", "Mana Potion", "Greater Mana Potion", "Super Mana Potion":
				potionType = data.ManaPotion
			case "Rejuvenation Potion", "Full Rejuvenation Potion":
				potionType = data.RejuvenationPotion
			}

			potions = append(potions, data.Potion{
				Row:    int(i.Position.Y),
				Column: int(i.Position.X),
				Type:   potionType,
			})
		}
	}
	return data.Inventory{
		Belt: data.Belt{
			Potions: potions,
		},
	}
}
