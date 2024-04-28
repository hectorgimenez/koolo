package config

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/hectorgimenez/d2go/pkg/data"

	"os"
	"strings"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	cp "github.com/otiai10/copy"

	"github.com/hectorgimenez/d2go/pkg/nip"

	"gopkg.in/yaml.v3"
)

var (
	Koolo      *KooloCfg
	Characters map[string]*CharacterCfg
	Version    = "dev"
)

type KooloCfg struct {
	Debug struct {
		Log       bool `yaml:"log"`
		RenderMap bool `yaml:"renderMap"`
	} `yaml:"debug"`
	FirstRun              bool   `yaml:"firstRun"`
	UseCustomSettings     bool   `yaml:"useCustomSettings"`
	GameWindowArrangement bool   `yaml:"gameWindowArrangement"`
	LogSaveDirectory      string `yaml:"logSaveDirectory"`
	D2LoDPath             string `yaml:"D2LoDPath"`
	D2RPath               string `yaml:"D2RPath"`
	Discord               struct {
		Enabled   bool   `yaml:"enabled"`
		ChannelID string `yaml:"channelId"`
		Token     string `yaml:"token"`
	} `yaml:"discord"`
	Telegram struct {
		Enabled bool   `yaml:"enabled"`
		ChatID  int64  `yaml:"chatId"`
		Token   string `yaml:"token"`
	}
}

type CharacterCfg struct {
	MaxGameLength int    `yaml:"maxGameLength"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	Realm         string `yaml:"realm"`
	CharacterName string `yaml:"characterName"`
	Health        struct {
		HealingPotionAt     int `yaml:"healingPotionAt"`
		ManaPotionAt        int `yaml:"manaPotionAt"`
		RejuvPotionAtLife   int `yaml:"rejuvPotionAtLife"`
		RejuvPotionAtMana   int `yaml:"rejuvPotionAtMana"`
		MercHealingPotionAt int `yaml:"mercHealingPotionAt"`
		MercRejuvPotionAt   int `yaml:"mercRejuvPotionAt"`
		ChickenAt           int `yaml:"chickenAt"`
		MercChickenAt       int `yaml:"mercChickenAt"`
	} `yaml:"health"`
	Inventory struct {
		InventoryLock [][]int     `yaml:"inventoryLock"`
		BeltColumns   BeltColumns `yaml:"beltColumns"`
	} `yaml:"inventory"`
	Character struct {
		Class         string `yaml:"class"`
		CastingFrames int    `yaml:"castingFrames"`
		UseMerc       bool   `yaml:"useMerc"`
		StashToShared bool   `yaml:"stashToShared"`
		UseTeleport   bool   `yaml:"useTeleport"`
	} `yaml:"character"`
	Game struct {
		MinGoldPickupThreshold int                   `yaml:"minGoldPickupThreshold"`
		ClearTPArea            bool                  `yaml:"clearTPArea"`
		Difficulty             difficulty.Difficulty `yaml:"difficulty"`
		RandomizeRuns          bool                  `yaml:"randomizeRuns"`
		Runs                   []Run                 `yaml:"runs"`
		Pindleskin             struct {
			SkipOnImmunities []stat.Resist `yaml:"skipOnImmunities"`
		} `yaml:"pindleskin"`
		Pit struct {
			MoveThroughBlackMarsh bool `yaml:"moveThroughBlackMarsh"`
			OpenChests            bool `yaml:"openChests"`
			FocusOnElitePacks     bool `yaml:"focusOnElitePacks"`
		} `yaml:"pit"`
		StonyTomb struct {
			OpenChests        bool `yaml:"openChests"`
			FocusOnElitePacks bool `yaml:"focusOnElitePacks"`
		} `yaml:"stony_tomb"`
		AncientTunnels struct {
			OpenChests        bool `yaml:"openChests"`
			FocusOnElitePacks bool `yaml:"focusOnElitePacks"`
		} `yaml:"ancient_tunnels"`
		Mephisto struct {
			KillCouncilMembers bool `yaml:"killCouncilMembers"`
			OpenChests         bool `yaml:"openChests"`
		} `yaml:"mephisto"`
		Tristram struct {
			ClearPortal       bool `yaml:"clearPortal"`
			FocusOnElitePacks bool `yaml:"focusOnElitePacks"`
		} `yaml:"tristram"`
		Nihlathak struct {
			ClearArea bool `yaml:"clearArea"`
		} `yaml:"nihlathak"`
		Diablo struct {
			KillDiablo bool `yaml:"killDiablo"`
			ClearArea  bool `yaml:"clearArea"`
			OnlyElites bool `yaml:"onlyElites"`
		} `yaml:"diablo"`
		Baal struct {
			KillBaal bool `yaml:"killBaal"`
		} `yaml:"baal"`
		Eldritch struct {
			KillShenk bool `yaml:"killShenk"`
		} `yaml:"eldritch"`
		TerrorZone struct {
			FocusOnElitePacks bool          `yaml:"focusOnElitePacks"`
			SkipOnImmunities  []stat.Resist `yaml:"skipOnImmunities"`
			SkipOtherRuns     bool          `yaml:"skipOtherRuns"`
			Areas             []area.ID     `yaml:"areas"`
		} `yaml:"terror_zone"`
		Leveling struct {
			EnsurePointsAllocation bool `yaml:"ensurePointsAllocation"`
			EnsureKeyBinding       bool `yaml:"ensureKeyBinding"`
		} `yaml:"leveling"`
	} `yaml:"game"`
	Companion struct {
		Enabled          bool   `yaml:"enabled"`
		Leader           bool   `yaml:"leader"`
		LeaderName       string `yaml:"leaderName"`
		Attack           bool   `yaml:"attack"`
		GameNameTemplate string `yaml:"gameNameTemplate"`
		GamePassword     string `yaml:"gamePassword"`
	} `yaml:"companion"`
	Gambling struct {
		Enabled bool        `yaml:"enabled"`
		Items   []item.Name `yaml:"items"`
	} `yaml:"gambling"`
	Runtime struct {
		CastDuration time.Duration `yaml:"-"`
		Rules        []nip.Rule    `yaml:"-"`
	} `yaml:"-"`
}

type BeltColumns [4]string

func (bm BeltColumns) Total(potionType data.PotionType) int {
	typeString := ""
	switch potionType {
	case data.HealingPotion:
		typeString = "healing"
	case data.ManaPotion:
		typeString = "mana"
	case data.RejuvenationPotion:
		typeString = "rejuvenation"
	}

	total := 0
	for _, v := range bm {
		if strings.EqualFold(v, typeString) {
			total++
		}
	}

	return total
}

// Load reads the config.ini file and returns a Config struct filled with data from the ini file
func Load() error {
	Characters = make(map[string]*CharacterCfg)
	r, err := os.Open("config/koolo.yaml")
	if err != nil {
		return fmt.Errorf("error loading koolo.yaml: %w", err)
	}

	d := yaml.NewDecoder(r)
	if err = d.Decode(&Koolo); err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	entries, err := os.ReadDir("config")
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		charCfg := CharacterCfg{}
		r, err = os.Open("config/" + entry.Name() + "/config.yaml")
		if err != nil {
			return fmt.Errorf("error loading config.yaml: %w", err)
		}

		d := yaml.NewDecoder(r)
		if err = d.Decode(&charCfg); err != nil {
			return fmt.Errorf("error reading %s character config: %w", entry.Name(), err)
		}

		rules, err := nip.ReadDir("config/" + entry.Name() + "/pickit/")
		if err != nil {
			return err
		}

		if charCfg.Game.Runs[0] == "leveling" {
			levelingRules, err := nip.ReadDir("config/" + entry.Name() + "/pickit_leveling/")
			if err != nil {
				return err
			}
			rules = append(rules, levelingRules...)
		}

		charCfg.Runtime.Rules = rules

		secs := float32(charCfg.Character.CastingFrames)*0.04 + 0.01
		charCfg.Runtime.CastDuration = time.Duration(secs*1000) * time.Millisecond
		Characters[entry.Name()] = &charCfg
	}

	return nil
}

func CreateFromTemplate(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	if _, err := os.Stat("config/" + name); !os.IsNotExist(err) {
		return errors.New("configuration with that name already exists")
	}

	err := cp.Copy("config/template", "config/"+name)
	if err != nil {
		return fmt.Errorf("error copying template: %w", err)
	}

	return Load()
}

func ValidateAndSaveConfig(config KooloCfg) error {
	// Trim executable from the path, just in case
	config.D2LoDPath = strings.ReplaceAll(strings.ToLower(config.D2LoDPath), "game.exe", "")
	config.D2RPath = strings.ReplaceAll(strings.ToLower(config.D2RPath), "d2r.exe", "")

	// Validate paths
	if _, err := os.Stat(config.D2LoDPath + "/d2data.mpq"); os.IsNotExist(err) {
		return errors.New("D2LoDPath is not valid")
	}

	if _, err := os.Stat(config.D2RPath + "/d2r.exe"); os.IsNotExist(err) {
		return errors.New("D2RPath is not valid")
	}

	text, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error parsing koolo config: %w", err)
	}

	err = os.WriteFile("config/koolo.yaml", text, 0644)
	if err != nil {
		return fmt.Errorf("error writing koolo config: %w", err)
	}

	return Load()
}

func SaveSupervisorConfig(supervisorName string, config *CharacterCfg) error {
	filePath := filepath.Join("config", supervisorName, "config.yaml")
	d, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, d, 0644)
	if err != nil {
		return fmt.Errorf("error writing supervisor config: %w", err)
	}

	return Load()
}
