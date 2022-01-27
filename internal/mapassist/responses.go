package mapassist

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

type gameDataHttpResponse struct {
	Success bool   `json:"success"`
	Area    string `json:"current_area"`
	Corpses []struct {
		Name     string `json:"name"`
		Position struct {
			X int `json:"X"`
			Y int `json:"Y"`
		} `json:"position"`
	} `json:"corpses"`
	PlayerUnit struct {
		Name string `json:"name"`
	} `json:"player_unit"`
}
