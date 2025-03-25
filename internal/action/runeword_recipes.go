package action

import (
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
)

type ItemBase struct {
	Name       string
	NumSockets int
}

type RunewordStatRolls struct {
	Min    float64
	Max    float64
	StatID stat.ID
	Layer  int
}

type Runeword struct {
	Enabled bool

	// Required
	Name          item.RunewordName
	Runes         []string // rune order is important! Always has to be the same
	BaseItemTypes []string
	Rolls         []RunewordStatRolls

	// Options
	AllowEth      bool
	AllowReroll   bool
	BaseSortOrder []stat.ID
	// BaseItems will override BaseItemTypes
	BaseItems        []item.Name
	PickNoSocketBase bool
}

var Runewords = []Runeword{
	{
		Name:          item.RunewordAncientsPledge,
		Runes:         []string{"RalRune", "OrtRune", "TalRune"},
		BaseItemTypes: []string{item.TypeShield, item.TypeAuricShields},
		AllowEth:      false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordBeast,
		Runes:         []string{"BerRune", "TirRune", "UmRune", "MalRune", "LumRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypeHammer, item.TypeScepter},
		Rolls: []RunewordStatRolls{
			{Min: 240, Max: 270, StatID: stat.EnhancedDamageMin},
			{Min: 25, Max: 40, StatID: stat.Strength},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordBlack,
		Runes:         []string{"ThulRune", "IoRune", "NefRune"},
		BaseItemTypes: []string{item.TypeClub, item.TypeHammer, item.TypeMace},
		BaseSortOrder: []stat.ID{stat.MinDamage},
	},
	{
		Name:          item.RunewordBone,
		Runes:         []string{"SolRune", "UmRune", "UmRune"},
		BaseItemTypes: []string{item.TypeArmor},
		Rolls: []RunewordStatRolls{
			{Min: 100, Max: 150, StatID: stat.MaxMana},
		},
		AllowReroll:   true,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordBramble,
		Runes:         []string{"RalRune", "OhmRune", "SurRune", "EthRune"},
		BaseItemTypes: []string{item.TypeArmor},
		Rolls: []RunewordStatRolls{
			{Min: 15, Max: 21, StatID: stat.Aura}, // 103
			{Min: 25, Max: 50, StatID: stat.PoisonSkillDamage},
		},
		AllowReroll:   true,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordBrand,
		Runes:         []string{"JahRune", "LoRune", "MalRune", "GulRune"},
		BaseItemTypes: []string{item.TypeAmazonBow, item.TypeBow, item.TypeCrossbow},
		Rolls: []RunewordStatRolls{
			{Min: 260, Max: 340, StatID: stat.EnhancedDamageMin},
			{Min: 280, Max: 330, StatID: stat.DemonDamagePercent},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.MinDamage},
	},
	{
		Name:          item.RunewordBreathOfTheDying,
		Runes:         []string{"VexRune", "HelRune", "ElRune", "EldRune", "ZodRune", "EthRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypeWand, item.TypeClub, item.TypeScepter, item.TypeMace, item.TypeHammer, item.TypeSword, item.TypeKnife, item.TypeSpear, item.TypePolearm, item.TypeStaff, item.TypeHandtoHand, item.TypeHandtoHand2},
		Rolls: []RunewordStatRolls{
			{Min: 350, Max: 400, StatID: stat.EnhancedDamageMin},
			{Min: 12, Max: 15, StatID: stat.LifeSteal},
		},
		AllowEth:      true,
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.MinDamage},
	},
	{
		Name:          item.RunewordBulwark,
		Runes:         []string{"ShaelRune", "IoRune", "SolRune"},
		BaseItemTypes: []string{item.TypeHelm, item.TypePelt, item.TypePrimalHelm, item.TypeCirclet},
		Rolls: []RunewordStatRolls{
			{Min: 4, Max: 6, StatID: stat.LifeSteal},
			{Min: 75, Max: 100, StatID: stat.EnhancedDefense},
			{Min: 10, Max: 15, StatID: stat.DamageReduced},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordCallToArms,
		Runes:         []string{"AmnRune", "RalRune", "MalRune", "IstRune", "OhmRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypeWand, item.TypeClub, item.TypeScepter, item.TypeMace, item.TypeHammer, item.TypeSword, item.TypeKnife, item.TypeSpear, item.TypePolearm, item.TypeStaff, item.TypeHandtoHand, item.TypeHandtoHand2},
		Rolls: []RunewordStatRolls{
			//	{Min: 240, Max: 290, StatID: stat.EnhancedDamageMin}, // Excluding temporarily
			{Min: 2, Max: 6, StatID: stat.NonClassSkill, Layer: 155}, // 155
			{Min: 1, Max: 6, StatID: stat.NonClassSkill, Layer: 149}, // 149
			{Min: 1, Max: 4, StatID: stat.NonClassSkill, Layer: 146}, // 146
		},
		AllowReroll: true,
		AllowEth:    true,
	},
	{
		Name:          item.RunewordChainsOfHonor,
		Runes:         []string{"DolRune", "UmRune", "BerRune", "IstRune"},
		BaseItemTypes: []string{item.TypeArmor},
		AllowEth:      true,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordChaos,
		Runes:         []string{"FalRune", "OhmRune", "UmRune"},
		BaseItemTypes: []string{item.TypeHandtoHand, item.TypeHandtoHand2},
		Rolls: []RunewordStatRolls{
			{Min: 240, Max: 290, StatID: stat.EnhancedDamageMin},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.SingleSkill, stat.MinDamage},
	},
	{
		Name:          item.RunewordCrescentMoon,
		Runes:         []string{"ShaelRune", "UmRune", "TirRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypeSword, item.TypePolearm},
		Rolls: []RunewordStatRolls{
			{Min: 180, Max: 220, StatID: stat.EnhancedDamageMin},
			{Min: 9, Max: 11, StatID: stat.AbsorbMagic},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordCure,
		Runes:         []string{"ShaelRune", "IoRune", "TalRune"},
		BaseItemTypes: []string{item.TypeHelm, item.TypePelt, item.TypePrimalHelm, item.TypeCirclet},
		Rolls: []RunewordStatRolls{
			{Min: 75, Max: 100, StatID: stat.EnhancedDefense},
			{Min: 40, Max: 60, StatID: stat.PoisonResist},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordDeath,
		Runes:         []string{"HelRune", "ElRune", "VexRune", "OrtRune", "GulRune"},
		BaseItemTypes: []string{item.TypeSword, item.TypeAxe},
		Rolls: []RunewordStatRolls{
			{Min: 300, Max: 385, StatID: stat.EnhancedDamageMin},
		},
		AllowEth:    true,
		AllowReroll: false,
	},
	{
		Name:          item.RunewordDelerium,
		Runes:         []string{"LemRune", "IstRune", "IoRune"},
		BaseItemTypes: []string{item.TypeHelm, item.TypePelt, item.TypePrimalHelm, item.TypeCirclet},
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordDestruction,
		Runes:         []string{"VexRune", "LoRune", "BerRune", "JahRune", "KoRune"},
		BaseItemTypes: []string{item.TypePolearm, item.TypeSword},
		BaseSortOrder: []stat.ID{stat.SingleSkill, stat.MinDamage},
	},
	{
		Name:          item.RunewordDoom,
		Runes:         []string{"HelRune", "OhmRune", "UmRune", "LoRune", "ChamRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypePolearm, item.TypeHammer},
		Rolls: []RunewordStatRolls{
			{Min: 330, Max: 370, StatID: stat.EnhancedDamageMin},
			{Min: 40, Max: 60, StatID: stat.PierceCold},
		},
		AllowEth:    true,
		AllowReroll: false,
	},
	{
		Name:          item.RunewordDragon,
		Runes:         []string{"SurRune", "LoRune", "SolRune"},
		BaseItemTypes: []string{item.TypeArmor, item.TypeShield, item.TypeAuricShields},
		Rolls:         []RunewordStatRolls{
			//	{Min: 3, Max: 5, StatID: "To All Attributes"},
		},
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordDream,
		Runes:         []string{"IoRune", "JahRune", "PulRune"},
		BaseItemTypes: []string{item.TypeHelm, item.TypePelt, item.TypePrimalHelm, item.TypeCirclet, item.TypeShield, item.TypeAuricShields},
		Rolls: []RunewordStatRolls{
			{Min: 20, Max: 30, StatID: stat.FasterHitRecovery},
			{Min: 150, Max: 220, StatID: stat.Defense},
			//	{Min: 5, Max: 20, StatID: "All Resistances"},
			{Min: 12, Max: 25, StatID: stat.MagicFind},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordDuress,
		Runes:         []string{"ShaelRune", "UmRune", "ThulRune"},
		BaseItemTypes: []string{item.TypeArmor},
		Rolls: []RunewordStatRolls{
			{Min: 10, Max: 20, StatID: stat.EnhancedDamageMin},
			{Min: 150, Max: 200, StatID: stat.EnhancedDefense},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordEdge,
		Runes:         []string{"TirRune", "TalRune", "AmnRune"},
		BaseItemTypes: []string{item.TypeAmazonBow, item.TypeBow, item.TypeCrossbow},
		Rolls: []RunewordStatRolls{
			{Min: 320, Max: 380, StatID: stat.DemonDamagePercent},
			//	{Min: 5, Max: 10, StatID: "To All Attributes"},
		},
		AllowReroll:   true,
		BaseSortOrder: []stat.ID{stat.MinDamage},
	},
	{
		Name:          item.RunewordEnigma,
		Runes:         []string{"JahRune", "IthRune", "BerRune"},
		BaseItemTypes: []string{item.TypeArmor},
		Rolls: []RunewordStatRolls{
			{Min: 750, Max: 775, StatID: stat.EnhancedDefense},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},

	{
		Name:          item.RunewordEnlightenment,
		Runes:         []string{"PulRune", "RalRune", "SolRune"},
		BaseItemTypes: []string{item.TypeArmor},
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordEternity,
		Runes:         []string{"AmnRune", "BerRune", "IstRune", "SolRune", "SurRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypeWand, item.TypeClub, item.TypeScepter, item.TypeMace, item.TypeHammer, item.TypeSword, item.TypeKnife, item.TypeSpear, item.TypePolearm, item.TypeStaff, item.TypeHandtoHand, item.TypeHandtoHand2},
		Rolls: []RunewordStatRolls{
			{Min: 260, Max: 310, StatID: stat.EnhancedDamageMin},
		},
		AllowEth:      true,
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.MinDamage},
	},
	{
		Name:          item.RunewordExile,
		Runes:         []string{"VexRune", "OhmRune", "IstRune", "DolRune"},
		BaseItemTypes: []string{item.TypeAuricShields},
		Rolls: []RunewordStatRolls{
			{Min: 13, Max: 16, StatID: stat.Aura, Layer: 104}, // 104
			{Min: 220, Max: 260, StatID: stat.EnhancedDefense},
		},
		AllowEth:      true,
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.PoisonResist},
	},
	{
		Name:          item.RunewordFaith,
		Runes:         []string{"OhmRune", "JahRune", "LemRune", "EldRune"},
		BaseItemTypes: []string{item.TypeAmazonBow, item.TypeBow, item.TypeCrossbow},
		Rolls: []RunewordStatRolls{
			{Min: 12, Max: 15, StatID: stat.Aura, Layer: 122}, // 122
			{Min: 1, Max: 2, StatID: stat.AllSkills},
		},
		AllowReroll:   true,
		BaseSortOrder: []stat.ID{stat.MinDamage},
	},
	{
		Name:          item.RunewordFamine,
		Runes:         []string{"FalRune", "OhmRune", "OrtRune", "JahRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypeHammer},
		Rolls: []RunewordStatRolls{
			{Min: 320, Max: 370, StatID: stat.EnhancedDamageMin},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.MinDamage},
	},
	{
		Name:          item.RunewordFlickeringFlame,
		Runes:         []string{"NefRune", "PulRune", "VexRune"},
		BaseItemTypes: []string{item.TypeHelm, item.TypePelt, item.TypePrimalHelm, item.TypeCirclet},
		Rolls: []RunewordStatRolls{
			{Min: 4, Max: 8, StatID: stat.Aura}, // Resist fire unmapped
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordFortitude,
		Runes:         []string{"ElRune", "SolRune", "DolRune", "LoRune"},
		BaseItemTypes: []string{item.TypeArmor},
		Rolls: []RunewordStatRolls{
			{Min: 8, Max: 12, StatID: stat.LifePerLevel},
			//	{Min: 25, Max: 30, StatID: "All Resistances"},
		},
		AllowReroll:   false,
		AllowEth:      true,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordFury,
		Runes:         []string{"JahRune", "GulRune", "EthRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypeWand, item.TypeClub, item.TypeScepter, item.TypeMace, item.TypeHammer, item.TypeSword, item.TypeKnife, item.TypeSpear, item.TypePolearm, item.TypeStaff, item.TypeHandtoHand, item.TypeHandtoHand2},
		BaseSortOrder: []stat.ID{stat.MinDamage},
	},
	{
		Name:          item.RunewordGloom,
		Runes:         []string{"FalRune", "UmRune", "PulRune"},
		BaseItemTypes: []string{item.TypeArmor},
		Rolls: []RunewordStatRolls{
			{Min: 200, Max: 260, StatID: stat.EnhancedDefense},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordGrief,
		Runes:         []string{"EthRune", "TirRune", "LoRune", "MalRune", "RalRune"},
		BaseItemTypes: []string{item.TypeSword, item.TypeAxe},
		Rolls: []RunewordStatRolls{
			{Min: 30, Max: 40, StatID: stat.IncreasedAttackSpeed},
			{Min: 340, Max: 400, StatID: stat.MinDamage}, // 0
			{Min: 20, Max: 25, StatID: stat.EnemyPoisonResist},
			{Min: 10, Max: 15, StatID: stat.LifeAfterEachKill},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.MinDamage},
	},
	{
		Name:          item.RunewordGround,
		Runes:         []string{"ShaelRune", "IoRune", "OrtRune"},
		BaseItemTypes: []string{item.TypeHelm, item.TypePelt, item.TypePrimalHelm, item.TypeCirclet},
		Rolls: []RunewordStatRolls{
			{Min: 75, Max: 100, StatID: stat.EnhancedDefense},
			{Min: 40, Max: 60, StatID: stat.LightningResist},
			{Min: 10, Max: 15, StatID: stat.AbsorbLightning},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordHandOfJustice,
		Runes:         []string{"SurRune", "ChamRune", "AmnRune", "LoRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypeWand, item.TypeClub, item.TypeScepter, item.TypeMace, item.TypeHammer, item.TypeSword, item.TypeKnife, item.TypeSpear, item.TypePolearm, item.TypeStaff, item.TypeHandtoHand, item.TypeHandtoHand2},
		Rolls: []RunewordStatRolls{
			{Min: 280, Max: 330, StatID: stat.EnhancedDamageMin},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.MinDamage},
	},
	{
		Name:          item.RunewordHarmony,
		Runes:         []string{"TirRune", "IthRune", "SolRune", "KoRune"},
		BaseItemTypes: []string{item.TypeAmazonBow, item.TypeBow, item.TypeCrossbow},
		Rolls: []RunewordStatRolls{
			{Min: 200, Max: 275, StatID: stat.EnhancedDamageMin},
			{Min: 2, Max: 6, StatID: stat.SingleSkill}, // 107
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.MinDamage},
	},
	{
		Name:          item.RunewordHeartOfTheOak,
		Runes:         []string{"KoRune", "VexRune", "PulRune", "ThulRune"},
		BaseItemTypes: []string{item.TypeMace},
		Rolls:         []RunewordStatRolls{
			//	{Min: 30, Max: 40, StatID: "All Resistances"},
		},
		AllowReroll: true,
		AllowEth:    true,
	},
	{
		Name:          item.RunewordHearth,
		Runes:         []string{"ShaelRune", "IoRune", "ThulRune"},
		BaseItemTypes: []string{item.TypeHelm, item.TypePelt, item.TypePrimalHelm, item.TypeCirclet},
		Rolls: []RunewordStatRolls{
			{Min: 75, Max: 100, StatID: stat.EnhancedDefense},
			{Min: 40, Max: 60, StatID: stat.ColdResist},
			{Min: 10, Max: 15, StatID: stat.AbsorbCold},
		},
		AllowReroll:   true,
		AllowEth:      true,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordHolyThunder,
		Runes:         []string{"EthRune", "RalRune", "OrtRune", "TalRune"},
		BaseItemTypes: []string{item.TypeScepter},
		BaseSortOrder: []stat.ID{stat.SingleSkill},
	},
	{
		Name:          item.RunewordHonor,
		Runes:         []string{"AmnRune", "ElRune", "IthRune", "TirRune", "SolRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypeWand, item.TypeClub, item.TypeScepter, item.TypeMace, item.TypeHammer, item.TypeSword, item.TypeKnife, item.TypeSpear, item.TypePolearm, item.TypeStaff, item.TypeHandtoHand, item.TypeHandtoHand2},
		BaseSortOrder: []stat.ID{stat.MinDamage},
	},
	{
		Name:          item.RunewordHustle,
		Runes:         []string{"ShaelRune", "KoRune", "EldRune"},
		BaseItemTypes: []string{item.TypeArmor, item.TypeAxe, item.TypeWand, item.TypeClub, item.TypeScepter, item.TypeMace, item.TypeHammer, item.TypeSword, item.TypeKnife, item.TypeSpear, item.TypePolearm, item.TypeStaff, item.TypeHandtoHand, item.TypeHandtoHand2},
		Rolls: []RunewordStatRolls{
			{Min: 180, Max: 200, StatID: stat.EnhancedDamage},
		},
	},
	{
		Name:          item.RunewordIce,
		Runes:         []string{"AmnRune", "ShaelRune", "JahRune", "LoRune"},
		BaseItemTypes: []string{item.TypeAmazonBow, item.TypeBow, item.TypeCrossbow},
		Rolls: []RunewordStatRolls{
			{Min: 140, Max: 210, StatID: stat.EnhancedDamageMin},
			{Min: 25, Max: 30, StatID: stat.ColdSkillDamage},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.MinDamage},
	},
	{
		Name:          item.RunewordInfinity,
		Runes:         []string{"BerRune", "MalRune", "BerRune", "IstRune"},
		BaseItemTypes: []string{item.TypePolearm, item.TypeSpear},
		Rolls: []RunewordStatRolls{
			{Min: 255, Max: 325, StatID: stat.EnhancedDamageMin},
			{Min: 45, Max: 55, StatID: stat.EnemyLightningResist},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordInsight,
		Runes:         []string{"RalRune", "TirRune", "TalRune", "SolRune"},
		BaseItemTypes: []string{item.TypePolearm, item.TypeAmazonBow, item.TypeBow, item.TypeCrossbow},
		Rolls: []RunewordStatRolls{
			{Min: 12, Max: 17, StatID: stat.Aura, Layer: 120}, // 120
			//	{Min: 200, Max: 260, StatID: stat.EnhancedDamageMin},
			{Min: 180, Max: 250, StatID: stat.AttackRatingPercent},
			{Min: 1, Max: 6, StatID: stat.NonClassSkill, Layer: 120}, // 9
		},
		AllowReroll:   true,
		AllowEth:      true,
		BaseSortOrder: []stat.ID{stat.TwoHandedMaxDamage},
	},
	{
		Name:          item.RunewordKingslayer,
		Runes:         []string{"MalRune", "UmRune", "GulRune", "FalRune"},
		BaseItemTypes: []string{item.TypeSword, item.TypeAxe},
		Rolls: []RunewordStatRolls{
			{Min: 230, Max: 270, StatID: stat.EnhancedDamage},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordKingsGrace,
		Runes:         []string{"AmnRune", "RalRune", "ThulRune"},
		BaseItemTypes: []string{item.TypeSword, item.TypeScepter},
		Rolls: []RunewordStatRolls{
			{Min: 100, Max: 150, StatID: stat.EnhancedDamageMin},
			{Min: 50, Max: 100, StatID: stat.DemonDamagePercent},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordLastWish,
		Runes:         []string{"JahRune", "MalRune", "JahRune", "SurRune", "JahRune", "BerRune"},
		BaseItemTypes: []string{item.TypeSword, item.TypeHammer, item.TypeAxe},
		Rolls: []RunewordStatRolls{
			{Min: 330, Max: 375, StatID: stat.EnhancedDamage},
			{Min: 60, Max: 70, StatID: stat.CrushingBlow},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.MinDamage, stat.TwoHandedMaxDamage},
	},
	{
		Name:          item.RunewordLawbringer,
		Runes:         []string{"AmnRune", "LemRune", "KoRune"},
		BaseItemTypes: []string{item.TypeSword, item.TypeHammer, item.TypeScepter},
		Rolls: []RunewordStatRolls{
			{Min: 16, Max: 18, StatID: stat.Aura, Layer: 119}, // 119
			{Min: 200, Max: 250, StatID: stat.DefenseVsMissiles},
		},
		AllowReroll: true,
	},
	{
		Name:          item.RunewordLeaf,
		Runes:         []string{"TirRune", "RalRune"},
		BaseItemTypes: []string{item.TypeStaff},
		Rolls: []RunewordStatRolls{
			{Min: 50, Max: 70, StatID: stat.FireSkillDamage},
		},
		AllowReroll: true,
		AllowEth:    true,
	},
	{
		Name:          item.RunewordLionheart,
		Runes:         []string{"HelRune", "LumRune", "FalRune"},
		BaseItemTypes: []string{item.TypeArmor},
		Rolls: []RunewordStatRolls{
			//	{Min: 10, Max: 15, StatID: "To All Attributes"},
			{Min: 15, Max: 20, StatID: stat.Vitality},
			{Min: 10, Max: 15, StatID: stat.Dexterity},
			{Min: 50, Max: 75, StatID: stat.MaxLife},
		},
		AllowReroll: true,
	},
	{
		Name:          item.RunewordLore,
		Runes:         []string{"OrtRune", "SolRune"},
		BaseItemTypes: []string{item.TypeHelm, item.TypePelt, item.TypePrimalHelm, item.TypeCirclet},
		AllowEth:      false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordMalice,
		Runes:         []string{"IthRune", "ElRune", "EthRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypeWand, item.TypeClub, item.TypeScepter, item.TypeMace, item.TypeHammer, item.TypeSword, item.TypeKnife, item.TypeSpear, item.TypePolearm, item.TypeStaff, item.TypeHandtoHand, item.TypeHandtoHand2},
	},
	{
		Name:          item.RunewordMelody,
		Runes:         []string{"ShaelRune", "KoRune", "NefRune"},
		BaseItemTypes: []string{item.TypeAmazonBow, item.TypeBow, item.TypeCrossbow},
	},
	{
		Name:          item.RunewordMemory,
		Runes:         []string{"LumRune", "IoRune", "SolRune", "EthRune"},
		BaseItemTypes: []string{item.TypeStaff},
	},
	{
		Name:          item.RunewordMetamorphosis,
		Runes:         []string{"IoRune", "ChamRune", "FalRune"},
		BaseItemTypes: []string{item.TypeHelm, item.TypePelt, item.TypePrimalHelm, item.TypeCirclet},
	},
	{
		Name:          item.RunewordMist,
		Runes:         []string{"ChamRune", "ShaelRune", "GulRune", "ThulRune", "IthRune"},
		BaseItemTypes: []string{item.TypeAmazonBow, item.TypeBow, item.TypeCrossbow},
		Rolls: []RunewordStatRolls{
			{Min: 8, Max: 12, StatID: stat.Aura, Layer: 113}, // 113
			{Min: 325, Max: 375, StatID: stat.EnhancedDamage},
		},
	},
	{
		Name:          item.RunewordMosaic,
		Runes:         []string{"MalRune", "GulRune", "AmnRune"},
		BaseItemTypes: []string{item.TypeHandtoHand, item.TypeHandtoHand2},
		Rolls: []RunewordStatRolls{
			{Min: 200, Max: 250, StatID: stat.EnhancedDamage},
			{Min: 8, Max: 15, StatID: stat.ColdSkillDamage},
			{Min: 8, Max: 15, StatID: stat.LightningSkillDamage},
			{Min: 8, Max: 15, StatID: stat.FireSkillDamage},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordMyth,
		Runes:         []string{"HelRune", "AmnRune", "NefRune"},
		BaseItemTypes: []string{item.TypeArmor},
	},
	{
		Name:          item.RunewordNadir,
		Runes:         []string{"NefRune", "TirRune"},
		BaseItemTypes: []string{item.TypeHelm, item.TypePelt, item.TypePrimalHelm, item.TypeCirclet},
	},
	{
		Name:          item.RunewordOath,
		Runes:         []string{"ShaelRune", "PulRune", "MalRune", "LumRune"},
		BaseItemTypes: []string{item.TypeSword, item.TypeAxe, item.TypeMace},
		Rolls: []RunewordStatRolls{
			{Min: 10, Max: 15, StatID: stat.AbsorbMagic},
			{Min: 210, Max: 340, StatID: stat.EnhancedDamageMin},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordObsession,
		Runes:         []string{"ZodRune", "IstRune", "LemRune", "LumRune", "IoRune", "NefRune"},
		BaseItemTypes: []string{item.TypeStaff},
		Rolls: []RunewordStatRolls{
			{Min: 15, Max: 25, StatID: stat.MaxLife},
			{Min: 15, Max: 30, StatID: stat.ManaRecoveryBonus},
			//	{Min: 60, Max: 70, StatID: "All Resistances"},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordObedience,
		Runes:         []string{"HelRune", "KoRune", "ThulRune", "EthRune", "FalRune"},
		BaseItemTypes: []string{item.TypePolearm, item.TypeSpear},
		Rolls: []RunewordStatRolls{
			{Min: 200, Max: 300, StatID: stat.Defense},
			//	{Min: 20, Max: 30, StatID: "All Resistances"},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.TwoHandedMaxDamage},
	},
	{
		Name:          item.RunewordPassion,
		Runes:         []string{"DolRune", "OrtRune", "EldRune", "LemRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypeWand, item.TypeClub, item.TypeScepter, item.TypeMace, item.TypeHammer, item.TypeSword, item.TypeKnife, item.TypeSpear, item.TypePolearm, item.TypeStaff, item.TypeHandtoHand, item.TypeHandtoHand2},
		Rolls: []RunewordStatRolls{
			{Min: 160, Max: 210, StatID: stat.EnhancedDamage},
			{Min: 50, Max: 80, StatID: stat.AttackRatingPercent},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordPattern,
		Runes:         []string{"TalRune", "OrtRune", "ThulRune"},
		BaseItemTypes: []string{item.TypeHandtoHand, item.TypeHandtoHand2},
		Rolls: []RunewordStatRolls{
			{Min: 40, Max: 80, StatID: stat.EnhancedDamage},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordPeace,
		Runes:         []string{"ShaelRune", "ThulRune", "AmnRune"},
		BaseItemTypes: []string{item.TypeArmor},
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordPhoenix,
		Runes:         []string{"VexRune", "VexRune", "LoRune", "JahRune"},
		BaseItemTypes: []string{item.TypeShield, item.TypeAuricShields, item.TypeAxe, item.TypeWand, item.TypeClub, item.TypeScepter, item.TypeMace, item.TypeHammer, item.TypeSword, item.TypeKnife, item.TypeSpear, item.TypePolearm, item.TypeStaff, item.TypeHandtoHand, item.TypeHandtoHand2},
		Rolls: []RunewordStatRolls{
			{Min: 10, Max: 15, StatID: stat.Aura, Layer: 124}, // 124
			{Min: 350, Max: 400, StatID: stat.EnhancedDamage},
			{Min: 350, Max: 400, StatID: stat.DefenseVsMissiles},
			{Min: 15, Max: 21, StatID: stat.AbsorbFire},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.PoisonResist},
	},
	{
		Name:          item.RunewordPlague,
		Runes:         []string{"ChamRune", "ShaelRune", "UmRune"},
		BaseItemTypes: []string{item.TypeSword, item.TypeKnife, item.TypeHandtoHand, item.TypeHandtoHand2},
		Rolls: []RunewordStatRolls{
			{Min: 260, Max: 320, StatID: stat.Aura}, // Cleansing unmapped
			{Min: 1, Max: 2, StatID: stat.AllSkills},
			{Min: 220, Max: 320, StatID: stat.EnhancedDamage},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordPride,
		Runes:         []string{"ChamRune", "SurRune", "IoRune", "LoRune"},
		BaseItemTypes: []string{item.TypePolearm, item.TypeSpear},
		Rolls: []RunewordStatRolls{
			{Min: 16, Max: 20, StatID: stat.Aura, Layer: 113}, // 113
			{Min: 260, Max: 300, StatID: stat.AttackRatingPercent},
		},
		AllowReroll:   true,
		BaseSortOrder: []stat.ID{stat.TwoHandedMaxDamage},
	},
	{
		Name:          item.RunewordPrinciple,
		Runes:         []string{"RalRune", "GulRune", "EldRune"},
		BaseItemTypes: []string{item.TypeArmor},
		Rolls: []RunewordStatRolls{
			{Min: 100, Max: 150, StatID: stat.MaxLife},
		},
		AllowReroll:   true,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordPrudence,
		Runes:         []string{"MalRune", "TirRune"},
		BaseItemTypes: []string{item.TypeArmor},
		Rolls: []RunewordStatRolls{
			{Min: 140, Max: 170, StatID: stat.EnhancedDefense},
			//	{Min: 25, Max: 35, StatID: "All Resistances"},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordRadiance,
		Runes:         []string{"NefRune", "SolRune", "IthRune"},
		BaseItemTypes: []string{item.TypeHelm, item.TypePelt, item.TypePrimalHelm, item.TypeCirclet},
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordRain,
		Runes:         []string{"OrtRune", "MalRune", "IthRune"},
		BaseItemTypes: []string{item.TypeArmor},
		Rolls: []RunewordStatRolls{
			{Min: 100, Max: 150, StatID: stat.MaxMana},
		},
		AllowReroll:   true,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordRhyme,
		Runes:         []string{"ShaelRune", "EthRune"},
		BaseItemTypes: []string{item.TypeShield, item.TypeAuricShields},
		AllowEth:      false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordRift,
		Runes:         []string{"HelRune", "KoRune", "LemRune", "GulRune"},
		BaseItemTypes: []string{item.TypePolearm, item.TypeScepter},
		Rolls:         []RunewordStatRolls{
			//	{Min: 5, Max: 10, StatID: "To All Attributes"},
		},
		AllowReroll: true,
	},
	{
		Name:          item.RunewordSanctuary,
		Runes:         []string{"KoRune", "KoRune", "MalRune"},
		BaseItemTypes: []string{item.TypeShield, item.TypeAuricShields},
		Rolls: []RunewordStatRolls{
			{Min: 130, Max: 160, StatID: stat.EnhancedDefense},
			//	{Min: 50, Max: 70, StatID: "All Resistances"},
		},
		AllowReroll: true,
	},
	{
		Name:          item.RunewordSilence,
		Runes:         []string{"DolRune", "EldRune", "HelRune", "IstRune", "TirRune", "VexRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypeWand, item.TypeClub, item.TypeScepter, item.TypeMace, item.TypeHammer, item.TypeSword, item.TypeKnife, item.TypeSpear, item.TypePolearm, item.TypeStaff, item.TypeHandtoHand, item.TypeHandtoHand2},
	},
	{
		Name:          item.RunewordSmoke,
		Runes:         []string{"NefRune", "LumRune"},
		BaseItemTypes: []string{item.TypeArmor},
		AllowEth:      false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordSpirit,
		Runes:         []string{"TalRune", "ThulRune", "OrtRune", "AmnRune"},
		BaseItemTypes: []string{item.TypeSword, item.TypeShield, item.TypeAuricShields},
		Rolls: []RunewordStatRolls{
			{Min: 25, Max: 35, StatID: stat.FasterCastRate},
			{Min: 89, Max: 112, StatID: stat.MaxMana},
			{Min: 3, Max: 8, StatID: stat.AbsorbMagic},
		},
		AllowReroll:   true,
		AllowEth:      false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordSplendor,
		Runes:         []string{"EthRune", "LumRune"},
		BaseItemTypes: []string{item.TypeShield, item.TypeAuricShields},
		Rolls: []RunewordStatRolls{
			{Min: 60, Max: 100, StatID: stat.EnhancedDefense},
		},
		AllowReroll:   false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordStealth,
		Runes:         []string{"TalRune", "EthRune"},
		BaseItemTypes: []string{item.TypeArmor},
		AllowEth:      false,
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordSteel,
		BaseItemTypes: []string{item.TypeSword, item.TypeAxe, item.TypeMace},
		Runes:         []string{"TirRune", "ElRune"},
	},
	{
		Name:          item.RunewordStone,
		Runes:         []string{"ShaelRune", "UmRune", "PulRune", "LumRune"},
		BaseItemTypes: []string{item.TypeArmor},
		Rolls: []RunewordStatRolls{
			{Min: 250, Max: 290, StatID: stat.EnhancedDefense},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordStrength,
		BaseItemTypes: []string{item.TypeAxe, item.TypeWand, item.TypeClub, item.TypeScepter, item.TypeMace, item.TypeHammer, item.TypeSword, item.TypeKnife, item.TypeSpear, item.TypePolearm, item.TypeStaff, item.TypeHandtoHand, item.TypeHandtoHand2},
		Runes:         []string{"AmnRune", "TirRune"},
	},
	{
		Name:          item.RunewordTemper,
		Runes:         []string{"ShaelRune", "IoRune", "RalRune"},
		BaseItemTypes: []string{item.TypeHelm, item.TypePelt, item.TypePrimalHelm, item.TypeCirclet},
		Rolls: []RunewordStatRolls{
			{Min: 75, Max: 100, StatID: stat.EnhancedDefense},
			{Min: 40, Max: 60, StatID: stat.FireResist},
			{Min: 10, Max: 15, StatID: stat.AbsorbFire},
		},
	},
	{
		Name:          item.RunewordTreachery,
		Runes:         []string{"ShaelRune", "ThulRune", "LemRune"},
		BaseItemTypes: []string{item.TypeArmor},
		BaseSortOrder: []stat.ID{stat.Defense},
	},
	{
		Name:          item.RunewordUnbendingWill,
		Runes:         []string{"FalRune", "IoRune", "IthRune", "EldRune", "ElRune", "HelRune"},
		BaseItemTypes: []string{item.TypeSword},
		Rolls: []RunewordStatRolls{
			{Min: 20, Max: 30, StatID: stat.IncreasedAttackSpeed},
			{Min: 300, Max: 350, StatID: stat.EnhancedDamage},
			{Min: 8, Max: 10, StatID: stat.LifeSteal},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordVenom,
		Runes:         []string{"TalRune", "DolRune", "MalRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypeWand, item.TypeClub, item.TypeScepter, item.TypeMace, item.TypeHammer, item.TypeSword, item.TypeKnife, item.TypeSpear, item.TypePolearm, item.TypeStaff, item.TypeHandtoHand, item.TypeHandtoHand2},
	},
	{
		Name:          item.RunewordVoiceOfReason,
		Runes:         []string{"LemRune", "KoRune", "ElRune", "EldRune"},
		BaseItemTypes: []string{item.TypeSword, item.TypeMace},
		Rolls: []RunewordStatRolls{
			{Min: 220, Max: 350, StatID: stat.DemonDamagePercent},
			{Min: 355, Max: 375, StatID: stat.UndeadDamagePercent},
		},
		AllowReroll: true,
	},
	{
		Name:          item.RunewordWealth,
		Runes:         []string{"LemRune", "KoRune", "TirRune"},
		BaseItemTypes: []string{item.TypeArmor},
	},
	{
		Name:          item.RunewordWhite,
		Runes:         []string{"DolRune", "IoRune"},
		BaseItemTypes: []string{item.TypeWand},
	},
	{
		Name:          item.RunewordWind,
		Runes:         []string{"SurRune", "ElRune"},
		BaseItemTypes: []string{item.TypeAxe, item.TypeWand, item.TypeClub, item.TypeScepter, item.TypeMace, item.TypeHammer, item.TypeSword, item.TypeKnife, item.TypeSpear, item.TypePolearm, item.TypeStaff, item.TypeHandtoHand, item.TypeHandtoHand2},
		Rolls: []RunewordStatRolls{
			{Min: 120, Max: 160, StatID: stat.EnhancedDamageMin},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordWisdom,
		Runes:         []string{"PulRune", "IthRune", "EldRune"},
		BaseItemTypes: []string{item.TypeHelm, item.TypePelt, item.TypePrimalHelm, item.TypeCirclet},
		Rolls: []RunewordStatRolls{
			{Min: 15, Max: 25, StatID: stat.AttackRatingPercent},
			{Min: 4, Max: 8, StatID: stat.ManaSteal},
		},
		AllowReroll: false,
	},
	{
		Name:          item.RunewordWrath,
		Runes:         []string{"PulRune", "LumRune", "BerRune", "MalRune"},
		BaseItemTypes: []string{item.TypeAmazonBow, item.TypeBow, item.TypeCrossbow},
		Rolls: []RunewordStatRolls{
			{Min: 250, Max: 300, StatID: stat.UndeadDamagePercent},
		},
		AllowReroll: true,
	},
	{
		Name:          item.RunewordZephyr,
		Runes:         []string{"OrtRune", "EthRune"},
		BaseItemTypes: []string{item.TypeAmazonBow, item.TypeBow, item.TypeCrossbow},
	},
}
