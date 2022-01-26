package mapassist

import (
	"encoding/json"
	"errors"
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

func (A APIClient) Inventory() inventory.Inventory {
	return inventory.Inventory{}
}

type statusHttpResponse struct {
	Success bool `json:"success"`
	Life    int  `json:"life"`
	MaxLife int  `json:"max_life"`
	Mana    int  `json:"mana"`
	MaxMana int  `json:"max_mana"`
	Merc    struct {
		Alive   bool `json:"alive"`
		Life    int  `json:"life"`
		MaxLife int  `json:"max_life"`
	} `json:"merc"`
}
