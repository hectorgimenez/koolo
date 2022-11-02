package map_client

type serverResponse struct {
	ID         string        `json:"id"`
	Seed       int           `json:"seed"`
	Difficulty int           `json:"difficulty"`
	Act        int           `json:"act"`
	Levels     []serverLevel `json:"levels"`
}

type serverLevel struct {
	Type   string         `json:"type"`
	ID     int            `json:"id"`
	Name   string         `json:"name"`
	Offset serverPosition `json:"offset"`
	Size   struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"size"`
	Objects []serverObject `json:"objects"`
	Map     [][]int
}

type serverPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type serverObject struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	serverPosition
}
