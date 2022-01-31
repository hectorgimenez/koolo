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
	Success    bool     `json:"success"`
	Area       string   `json:"area"`
	AreaOrigin position `json:"area_origin"`
	Corpses    []struct {
		Name     string   `json:"name"`
		Position position `json:"position"`
	} `json:"corpses"`
	PlayerUnit struct {
		Name     string   `json:"name"`
		Position position `json:"position"`
	} `json:"player_unit"`
	Items []struct {
		Position position `json:"position"`
		Name     string   `json:"name"`
		Place    string   `json:"place"`
	} `json:"items"`
	Monsters []struct {
		Position position `json:"position"`
		Name     string   `json:"name"`
	} `json:"monsters"`
	NPCs []struct {
		Name      string     `json:"name"`
		Positions []position `json:"positions"`
	}
	CollisionGrid [][]int `json:"collision_grid"`
}

type position struct {
	X float32 `json:"X"`
	Y float32 `json:"Y"`
}
