package config

type Run string

const (
	CountessRun         Run = "countess"
	AndarielRun         Run = "andariel"
	AncientTunnelsRun   Run = "ancient_tunnels"
	MausoleumRun        Run = "mausoleum"
	SummonerRun         Run = "summoner"
	DurielRun           Run = "duriel"
	MephistoRun         Run = "mephisto"
	TravincalRun        Run = "travincal"
	EldritchRun         Run = "eldritch"
	PindleskinRun       Run = "pindleskin"
	NihlathakRun        Run = "nihlathak"
	TristramRun         Run = "tristram"
	LowerKurastRun      Run = "lower_kurast"
	LowerKurastChestRun Run = "lower_kurast_chest"
	StonyTombRun        Run = "stony_tomb"
	PitRun              Run = "pit"
	ArachnidLairRun     Run = "arachnid_lair"
	TalRashaTombsRun    Run = "tal_rasha_tombs"
	BaalRun             Run = "baal"
	DiabloRun           Run = "diablo"
	CowsRun             Run = "cows"
	LevelingRun         Run = "leveling"
	FollowerRun         Run = "follower"
	QuestsRun           Run = "quests"
	TerrorZoneRun       Run = "terror_zone"
	ThreshsocketRun     Run = "threshsocket"
	DrifterCavernRun    Run = "drifter_cavern"
	SpiderCavernRun     Run = "spider_cavern"
	EnduguRun           Run = "endugu"
)

var AvailableRuns = map[Run]interface{}{
	CountessRun:         nil,
	AndarielRun:         nil,
	AncientTunnelsRun:   nil,
	MausoleumRun:        nil,
	SummonerRun:         nil,
	DurielRun:           nil,
	MephistoRun:         nil,
	TravincalRun:        nil,
	EldritchRun:         nil,
	PindleskinRun:       nil,
	NihlathakRun:        nil,
	TristramRun:         nil,
	LowerKurastRun:      nil,
	LowerKurastChestRun: nil,
	StonyTombRun:        nil,
	PitRun:              nil,
	ArachnidLairRun:     nil,
	TalRashaTombsRun:    nil,
	BaalRun:             nil,
	DiabloRun:           nil,
	CowsRun:             nil,
	LevelingRun:         nil,
	QuestsRun:           nil,
	FollowerRun:         nil,
	TerrorZoneRun:       nil,
	ThreshsocketRun:     nil,
	DrifterCavernRun:    nil,
	SpiderCavernRun:     nil,
	EnduguRun:           nil,
}
