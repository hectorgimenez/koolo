package area

type Area int

func (a Area) IsTown() bool {
	switch a {
	case RogueEncampment, LutGholein, KurastDocks, ThePandemoniumFortress, Harrogath:
		return true
	}

	return false
}

const (
	Abaddon                  Area = 125
	AncientTunnels           Area = 65
	ArcaneSanctuary          Area = 74
	ArreatPlateau            Area = 112
	ArreatSummit             Area = 120
	Barracks                 Area = 28
	BlackMarsh               Area = 6
	BloodMoor                Area = 2
	BloodyFoothills          Area = 110
	BurialGrounds            Area = 17
	CanyonOfTheMagi          Area = 46
	CatacombsLevel1          Area = 34
	CatacombsLevel2          Area = 35
	CatacombsLevel3          Area = 36
	CatacombsLevel4          Area = 37
	Cathedral                Area = 33
	CaveLevel1               Area = 9
	CaveLevel2               Area = 13
	ChaosSanctuary           Area = 108
	CityOfTheDamned          Area = 106
	ClawViperTempleLevel1    Area = 58
	ClawViperTempleLevel2    Area = 61
	ColdPlains               Area = 3
	Crypt                    Area = 18
	CrystallinePassage       Area = 113
	DarkWood                 Area = 5
	DenOfEvil                Area = 8
	DisusedFane              Area = 95
	DisusedReliquary         Area = 99
	DrifterCavern            Area = 116
	DryHills                 Area = 42
	DuranceOfHateLevel1      Area = 100
	DuranceOfHateLevel2      Area = 101
	DuranceOfHateLevel3      Area = 102
	DurielsLair              Area = 73
	FarOasis                 Area = 43
	FlayerDungeonLevel1      Area = 88
	FlayerDungeonLevel2      Area = 89
	FlayerDungeonLevel3      Area = 91
	FlayerJungle             Area = 78
	ForgottenReliquary       Area = 96
	ForgottenSands           Area = 134
	ForgottenTemple          Area = 97
	ForgottenTower           Area = 20
	FrigidHighlands          Area = 111
	FrozenRiver              Area = 114
	FrozenTundra             Area = 117
	FurnaceOfPain            Area = 135
	GlacialTrail             Area = 115
	GreatMarsh               Area = 77
	HallsOfAnguish           Area = 122
	HallsOfPain              Area = 123
	HallsOfTheDeadLevel1     Area = 56
	HallsOfTheDeadLevel2     Area = 57
	HallsOfTheDeadLevel3     Area = 60
	HallsOfVaught            Area = 124
	HaremLevel1              Area = 50
	HaremLevel2              Area = 51
	Harrogath                Area = 109
	HoleLevel1               Area = 11
	HoleLevel2               Area = 15
	IcyCellar                Area = 119
	InfernalPit              Area = 127
	InnerCloister            Area = 32
	JailLevel1               Area = 29
	JailLevel2               Area = 30
	JailLevel3               Area = 31
	KurastBazaar             Area = 80
	KurastCauseway           Area = 82
	KurastDocks              Area = 75
	LostCity                 Area = 44
	LowerKurast              Area = 79
	LutGholein               Area = 40
	MaggotLairLevel1         Area = 62
	MaggotLairLevel2         Area = 63
	MaggotLairLevel3         Area = 64
	MatronsDen               Area = 133
	Mausoleum                Area = 19
	MonasteryGate            Area = 26
	MooMooFarm               Area = 39
	NihlathaksTemple         Area = 121
	None                     Area = 0
	OuterCloister            Area = 27
	OuterSteppes             Area = 104
	PalaceCellarLevel1       Area = 52
	PalaceCellarLevel2       Area = 53
	PalaceCellarLevel3       Area = 54
	PitLevel1                Area = 12
	PitLevel2                Area = 16
	PitOfAcheron             Area = 126
	PlainsOfDespair          Area = 105
	RiverOfFlame             Area = 107
	RockyWaste               Area = 41
	RogueEncampment          Area = 1
	RuinedFane               Area = 98
	RuinedTemple             Area = 94
	SewersLevel1Act2         Area = 47
	SewersLevel1Act3         Area = 92
	SewersLevel2Act2         Area = 48
	SewersLevel2Act3         Area = 93
	SewersLevel3Act2         Area = 49
	SpiderCave               Area = 84
	SpiderCavern             Area = 85
	SpiderForest             Area = 76
	StonyField               Area = 4
	StonyTombLevel1          Area = 55
	StonyTombLevel2          Area = 59
	SwampyPitLevel1          Area = 86
	SwampyPitLevel2          Area = 87
	SwampyPitLevel3          Area = 90
	TalRashasTomb1           Area = 66
	TalRashasTomb2           Area = 67
	TalRashasTomb3           Area = 68
	TalRashasTomb4           Area = 69
	TalRashasTomb5           Area = 70
	TalRashasTomb6           Area = 71
	TalRashasTomb7           Area = 72
	TamoeHighland            Area = 7
	TheAncientsWay           Area = 118
	ThePandemoniumFortress   Area = 103
	TheWorldstoneChamber     Area = 132
	TheWorldStoneKeepLevel1  Area = 128
	TheWorldStoneKeepLevel2  Area = 129
	TheWorldStoneKeepLevel3  Area = 130
	ThroneOfDestruction      Area = 131
	TowerCellarLevel1        Area = 21
	TowerCellarLevel2        Area = 22
	TowerCellarLevel3        Area = 23
	TowerCellarLevel4        Area = 24
	TowerCellarLevel5        Area = 25
	Travincal                Area = 83
	Tristram                 Area = 38
	UberTristram             Area = 136
	UndergroundPassageLevel1 Area = 10
	UndergroundPassageLevel2 Area = 14
	UpperKurast              Area = 81
	ValleyOfSnakes           Area = 45
	MapsAncientTemple        Area = 137
	MapsDesecratedTemple     Area = 138
	MapsFrigidPlateau        Area = 139
	MapsInfernalTrial        Area = 140
	MapsRuinedCitadel        Area = 141
)
