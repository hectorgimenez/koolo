package mapassist

import (
	"encoding/json"
	"errors"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/inventory"
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

func (A APIClient) GameData() game.Data {
	// TODO: Fix on MapAssist, first request always returns old data
	http.Get(A.hostName + genericData)
	r, _ := http.Get(A.hostName + genericData)

	data := gameDataHttpResponse{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		// TODO: Handle error
		return game.Data{}
	}
	if !data.Success {
		// TODO: Handle error
		return game.Data{}
	}

	corpse := game.Corpse{}
	for _, c := range data.Corpses {
		if c.Name == data.PlayerUnit.Name {
			corpse.Found = true
			corpse.X = c.Position.X
			corpse.Y = c.Position.Y
		}
	}
	return game.Data{
		Area:   game.Area(data.Area),
		Corpse: corpse,
	}
}

func (A APIClient) Inventory() inventory.Inventory {
	return inventory.Inventory{}
}
