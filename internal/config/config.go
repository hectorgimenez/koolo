package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hectorgimenez/koolo/internal/game/difficulty"
	"github.com/hectorgimenez/koolo/internal/game/item"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"gopkg.in/yaml.v3"
)

var (
	Config StructConfig
	Pickit StructPickit
)

type StructConfig struct {
	Display int `yaml:"display"`
	Debug   struct {
		Log       bool `yaml:"log"`
		RenderMap bool `yaml:"renderMap"`
	} `yaml:"debug"`
	LogFilePath   string `yaml:"logFilePath"`
	MaxGameLength int    `yaml:"maxGameLength"`
	D2LoDPath     string `yaml:"D2LoDPath"`
	Discord       struct {
		Enabled   bool   `yaml:"enabled"`
		ChannelID string `yaml:"channelId"`
		Token     string `yaml:"token"`
	} `yaml:"discord"`
	Telegram struct {
		Enabled bool   `yaml:"enabled"`
		ChatID  int64  `yaml:"chatId"`
		Token   string `yaml:"token"`
	}
	Controller struct {
		Webserver bool `yaml:"webserver"`
		Port      int  `yaml:"port"`
		Webview   bool `yaml:"webview"`
	} `yaml:"controller"`
	Health struct {
		HealingPotionAt     int `yaml:"healingPotionAt"`
		ManaPotionAt        int `yaml:"manaPotionAt"`
		RejuvPotionAtLife   int `yaml:"rejuvPotionAtLife"`
		RejuvPotionAtMana   int `yaml:"rejuvPotionAtMana"`
		MercHealingPotionAt int `yaml:"mercHealingPotionAt"`
		MercRejuvPotionAt   int `yaml:"mercRejuvPotionAt"`
		ChickenAt           int `yaml:"chickenAt"`
		MercChickenAt       int `yaml:"mercChickenAt"`
	} `yaml:"health"`
	Bindings struct {
		OpenInventory    string `yaml:"openInventory"`
		Potion1          string `yaml:"potion1"`
		Potion2          string `yaml:"potion2"`
		Potion3          string `yaml:"potion3"`
		Potion4          string `yaml:"potion4"`
		ForceMove        string `yaml:"forceMove"`
		StandStill       string `yaml:"standStill"`
		SwapWeapon       string `yaml:"swapWeapon"`
		Teleport         string `yaml:"teleport"`
		TP               string `yaml:"tp"`
		CTABattleCommand string `yaml:"CTABattleCommand"`
		CTABattleOrders  string `yaml:"CTABattleOrders"`

		// Class Specific bindings
		Sorceress struct {
			Blizzard    string `yaml:"blizzard"`
			StaticField string `yaml:"staticField"`
			FrozenArmor string `yaml:"frozenArmor"`
		} `yaml:"sorceress"`
		Hammerdin struct {
			Concentration string `yaml:"concentration"`
			HolyShield    string `yaml:"holyShield"`
			Vigor         string `yaml:"vigor"`
			Redemption    string `yaml:"redemption"`
			Cleansing     string `yaml:"cleansing"`
		} `yaml:"hammerdin"`
	} `yaml:"bindings"`
	Inventory struct {
		InventoryLock [][]int `yaml:"inventoryLock"`
		BeltColumns   struct {
			Healing      int `yaml:"healing"`
			Mana         int `yaml:"mana"`
			Rejuvenation int `yaml:"rejuvenation"`
		} `yaml:"beltColumns"`
		BeltRows int `yaml:"beltRows"`
	} `yaml:"inventory"`
	Character struct {
		Class         string `yaml:"class"`
		CastingFrames int    `yaml:"castingFrames"`
		UseMerc       bool   `yaml:"useMerc"`
		UseCTA        bool   `yaml:"useCTA"`
	} `yaml:"character"`
	Game struct {
		ClearTPArea   bool                  `yaml:"clearTPArea"`
		Difficulty    difficulty.Difficulty `yaml:"difficulty"`
		RandomizeRuns bool                  `yaml:"randomizeRuns"`
		Runs          []string              `yaml:"runs"`
		Pindleskin    struct {
			SkipOnImmunities []stat.Resist `yaml:"skipOnImmunities"`
		} `yaml:"pindleskin"`
		Tristram struct {
			ClearPortal       bool `yaml:"clearPortal"`
			FocusOnElitePacks bool `yaml:"focusOnElitePacks"`
		} `yaml:"tristram"`
		Nihlathak struct {
			ClearArea bool `yaml:"clearArea"`
		}
	} `yaml:"game"`
	Companion struct {
		Enabled          bool   `yaml:"enabled"`
		Leader           bool   `yaml:"leader"`
		LeaderName       string `yaml:"leaderName"`
		Remote           string `yaml:"remote"`
		GameNameTemplate string `yaml:"gameNameTemplate"`
	} `yaml:"companion"`
	Runtime struct {
		CastDuration time.Duration
	}
}

type StructPickit struct {
	PickupGold          bool `yaml:"pickupGold"`
	MinimumGoldToPickup int  `yaml:"minimumGoldToPickup"`
	Items               []ItemPickit
}

type ItemPickit struct {
	Name     string
	Quality  []item.Quality
	Sockets  []int
	Ethereal *bool
	Stats    map[string]int
}

type ItemStat *int

// Load reads the config.ini file and returns a Config struct filled with data from the ini file
func Load() error {
	r, err := os.Open("config/config.yaml")
	if err != nil {
		return fmt.Errorf("error loading config.yaml: %w", err)
	}

	d := yaml.NewDecoder(r)
	if err = d.Decode(&Config); err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	r, err = os.Open("config/pickit.yaml")
	if err != nil {
		return fmt.Errorf("error loading pickit.yaml: %w", err)
	}

	d = yaml.NewDecoder(r)
	if err = d.Decode(&Pickit); err != nil {
		return fmt.Errorf("error reading pickit: %w", err)
	}

	b, err := os.ReadFile("config/pickit.yaml")
	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(b, &m)
	if err != nil {
		return fmt.Errorf("error decoding pickit items: %w", err)
	}
	items := parsePickitItems(m["items"].([]interface{}))
	Pickit.Items = items

	secs := float32(Config.Character.CastingFrames)*0.04 + 0.01
	Config.Runtime.CastDuration = time.Duration(secs*1000) * time.Millisecond

	return nil
}

func parsePickitItems(items []interface{}) []ItemPickit {
	var itemsToPickit []ItemPickit
	for _, it := range items {
		for name, props := range it.(map[string]interface{}) {
			ip := ItemPickit{
				Name:  name,
				Stats: map[string]int{},
			}

			if props != nil {
				for statName, statValue := range props.(map[string]interface{}) {
					statName = strings.ToLower(statName)
					switch statName {
					case "sockets":
						v, ok := statValue.(int)
						if ok {
							ip.Sockets = []int{v}
						} else {
							for _, val := range statValue.([]interface{}) {
								ip.Sockets = append(ip.Sockets, val.(int))
							}
						}
					case "quality":
						v, ok := statValue.(string)
						if ok {
							ip.Quality = []item.Quality{qualityToEnum(v)}
						} else {
							for _, val := range statValue.([]interface{}) {
								ip.Quality = append(ip.Quality, qualityToEnum(val.(string)))
							}
						}
					case "ethereal":
						ethp := statValue.(bool)
						ip.Ethereal = &ethp
					default:
						ip.Stats[statName] = statValue.(int)
					}
				}
			}
			itemsToPickit = append(itemsToPickit, ip)
		}
	}

	return itemsToPickit
}

func qualityToEnum(quality string) item.Quality {
	switch quality {
	case "normal":
		return item.ItemQualityNormal
	case "magic":
		return item.ItemQualityMagic
	case "rare":
		return item.ItemQualityRare
	case "unique":
		return item.ItemQualityUnique
	case "set":
		return item.ItemQualitySet
	case "superior":
		return item.ItemQualitySuperior
	}

	return item.ItemQualityNormal
}
