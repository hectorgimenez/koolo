package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"

	"os"
	"strings"

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
		Log         bool `yaml:"log"`
		Screenshots bool `yaml:"screenshots"`
		RenderMap   bool `yaml:"renderMap"`
	} `yaml:"debug"`
	FirstRun              bool   `yaml:"firstRun"`
	UseCustomSettings     bool   `yaml:"useCustomSettings"`
	GameWindowArrangement bool   `yaml:"gameWindowArrangement"`
	LogSaveDirectory      string `yaml:"logSaveDirectory"`
	D2LoDPath             string `yaml:"D2LoDPath"`
	D2RPath               string `yaml:"D2RPath"`
	Discord               struct {
		Enabled                      bool     `yaml:"enabled"`
		EnableGameCreatedMessages    bool     `yaml:"enableGameCreatedMessages"`
		EnableNewRunMessages         bool     `yaml:"enableNewRunMessages"`
		EnableRunFinishMessages      bool     `yaml:"enableRunFinishMessages"`
		EnableDiscordChickenMessages bool     `yaml:"enableDiscordChickenMessages"`
		BotAdmins                    []string `yaml:"botAdmins"`
		ChannelID                    string   `yaml:"channelId"`
		Token                        string   `yaml:"token"`
	} `yaml:"discord"`
	Telegram struct {
		Enabled bool   `yaml:"enabled"`
		ChatID  int64  `yaml:"chatId"`
		Token   string `yaml:"token"`
	}
}

type Day struct {
	DayOfWeek  int         `yaml:"dayOfWeek"`
	TimeRanges []TimeRange `yaml:"timeRange"`
}

type Scheduler struct {
	Enabled bool  `yaml:"enabled"`
	Days    []Day `yaml:"days"`
}

type TimeRange struct {
	Start time.Time `yaml:"start"`
	End   time.Time `yaml:"end"`
}

type CharacterCfg struct {
	MaxGameLength   int    `yaml:"maxGameLength"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	AuthMethod      string `yaml:"authMethod"`
	AuthToken       string `yaml:"authToken"`
	Realm           string `yaml:"realm"`
	CharacterName   string `yaml:"characterName"`
	CommandLineArgs string `yaml:"commandLineArgs"`
	KillD2OnStop    bool   `yaml:"killD2OnStop"`
	ClassicMode     bool   `yaml:"classicMode"`
	CloseMiniPanel  bool   `yaml:"closeMiniPanel"`

	Scheduler Scheduler `yaml:"scheduler"`
	Health    struct {
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
		CreateLobbyGames       bool                  `yaml:"createLobbyGames"`
		PublicGameCounter      int                   `yaml:"-"`
		Pindleskin             struct {
			SkipOnImmunities []stat.Resist `yaml:"skipOnImmunities"`
		} `yaml:"pindleskin"`
		Cows struct {
			OpenChests bool `yaml:"openChests"`
		}
		Pit struct {
			MoveThroughBlackMarsh bool `yaml:"moveThroughBlackMarsh"`
			OpenChests            bool `yaml:"openChests"`
			FocusOnElitePacks     bool `yaml:"focusOnElitePacks"`
			OnlyClearLevel2       bool `yaml:"onlyClearLevel2"`
		} `yaml:"pit"`
		Andariel struct {
			ClearRoom bool `yaml:"clearRoom"`
		}
		StonyTomb struct {
			OpenChests        bool `yaml:"openChests"`
			FocusOnElitePacks bool `yaml:"focusOnElitePacks"`
		} `yaml:"stony_tomb"`
		Mausoleum struct {
			OpenChests        bool `yaml:"openChests"`
			FocusOnElitePacks bool `yaml:"focusOnElitePacks"`
		} `yaml:"mausoleum"`
		AncientTunnels struct {
			OpenChests        bool `yaml:"openChests"`
			FocusOnElitePacks bool `yaml:"focusOnElitePacks"`
		} `yaml:"ancient_tunnels"`
		DrifterCavern struct {
			OpenChests        bool `yaml:"openChests"`
			FocusOnElitePacks bool `yaml:"focusOnElitePacks"`
		} `yaml:"drifter_cavern"`
		SpiderCavern struct {
			OpenChests        bool `yaml:"openChests"`
			FocusOnElitePacks bool `yaml:"focusOnElitePacks"`
		} `yaml:"spider_cavern"`
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
			KillDiablo        bool `yaml:"killDiablo"`
			FullClear         bool `yaml:"fullClear"`
			FocusOnElitePacks bool `yaml:"focusOnElitePacks"`
			SkipStormcasters  bool `yaml:"skipStormcasters"`
		} `yaml:"diablo"`
		Baal struct {
			KillBaal    bool `yaml:"killBaal"`
			DollQuit    bool `yaml:"dollQuit"`
			SoulQuit    bool `yaml:"soulQuit"`
			ClearFloors bool `yaml:"clearFloors"`
			OnlyElites  bool `yaml:"onlyElites"`
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
		Quests struct {
			ClearDen       bool `yaml:"clearDen"`
			RescueCain     bool `yaml:"rescueCain"`
			RetrieveHammer bool `yaml:"retrieveHammer"`
			GetCube        bool `yaml:"getCube"`
			KillRadament   bool `yaml:"killRadament"`
			RetrieveBook   bool `yaml:"retrieveBook"`
			KillIzual      bool `yaml:"killIzual"`
			KillShenk      bool `yaml:"killShenk"`
			RescueAnya     bool `yaml:"rescueAnya"`
			KillAncients   bool `yaml:"killAncients"`
		} `yaml:"quests"`
	} `yaml:"game"`
	Companion struct {
		Leader           bool   `yaml:"leader"`
		LeaderName       string `yaml:"leaderName"`
		GameNameTemplate string `yaml:"gameNameTemplate"`
		GamePassword     string `yaml:"gamePassword"`
	} `yaml:"companion"`
	Gambling struct {
		Enabled bool        `yaml:"enabled"`
		Items   []item.Name `yaml:"items"`
	} `yaml:"gambling"`
	CubeRecipes struct {
		Enabled        bool     `yaml:"enabled"`
		EnabledRecipes []string `yaml:"enabledRecipes"`
	} `yaml:"cubing"`
	BackToTown struct {
		NoHpPotions     bool `yaml:"noHpPotions"`
		NoMpPotions     bool `yaml:"noMpPotions"`
		MercDied        bool `yaml:"mercDied"`
		EquipmentBroken bool `yaml:"equipmentBroken"`
	} `yaml:"backtotown"`
	Runtime struct {
		Rules nip.Rules   `yaml:"-"`
		Drops []data.Item `yaml:"-"`
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

	// Get the absolute path of the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %w", err)
	}

	// Function to get absolute path
	getAbsPath := func(relPath string) string {
		return filepath.Join(cwd, relPath)
	}

	kooloPath := getAbsPath("config/koolo.yaml")
	r, err := os.Open(kooloPath)
	if err != nil {
		return fmt.Errorf("error loading koolo.yaml: %w", err)
	}
	defer r.Close()

	d := yaml.NewDecoder(r)
	if err = d.Decode(&Koolo); err != nil {
		return fmt.Errorf("error reading config %s: %w", kooloPath, err)
	}

	configDir := getAbsPath("config")
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return fmt.Errorf("error reading config directory %s: %w", configDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		charCfg := CharacterCfg{}
		charConfigPath := getAbsPath(filepath.Join("config", entry.Name(), "config.yaml"))
		r, err = os.Open(charConfigPath)
		if err != nil {
			return fmt.Errorf("error loading config.yaml: %w", err)
		}
		defer r.Close()

		d := yaml.NewDecoder(r)
		if err = d.Decode(&charCfg); err != nil {
			return fmt.Errorf("error reading %s character config: %w", charConfigPath, err)
		}

		pickitPath := getAbsPath(filepath.Join("config", entry.Name(), "pickit")) + "\\"
		rules, err := nip.ReadDir(pickitPath)
		if err != nil {
			return fmt.Errorf("error reading pickit directory %s: %w", pickitPath, err)
		}

		if len(charCfg.Game.Runs) > 0 && charCfg.Game.Runs[0] == "leveling" {
			levelingPickitPath := getAbsPath(filepath.Join("config", entry.Name(), "pickit_leveling")) + "\\"
			levelingRules, err := nip.ReadDir(levelingPickitPath)
			if err != nil {
				return fmt.Errorf("error reading pickit_leveling directory %s: %w", levelingPickitPath, err)
			}
			rules = append(rules, levelingRules...)
		}

		charCfg.Runtime.Rules = rules

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
