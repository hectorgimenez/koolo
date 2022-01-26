package mapassist

import (
	"encoding/json"
	"errors"
	koolo "github.com/hectorgimenez/koolo/internal"
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

func (A APIClient) CurrentStatus() (koolo.Status, error) {
	r, err := http.Get(A.hostName + healthEndpoint)
	if err != nil {
		return koolo.Status{}, err
	}

	status := statusHttpResponse{}
	err = json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		return koolo.Status{}, err
	}
	if !status.Success {
		return koolo.Status{}, errors.New("error fetching MapAssist data from API")
	}

	return koolo.Status{
		Life:    status.Life,
		MaxLife: status.MaxLife,
		Mana:    status.Mana,
		MaxMana: status.MaxMana,
		Merc: koolo.MercStatus{
			Alive:   status.Merc.Alive,
			Life:    status.Merc.Life,
			MaxLife: status.Merc.MaxLife,
		},
	}, nil
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
