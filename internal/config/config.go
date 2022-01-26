package config

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/inventory"
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Display     int    `yaml:"display"`
	Debug       bool   `yaml:"debug"`
	LogFilePath string `yaml:"logFilePath"`
	Health      struct {
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
		Potion1 string `yaml:"potion1"`
		Potion2 string `yaml:"potion2"`
		Potion3 string `yaml:"potion3"`
		Potion4 string `yaml:"potion4"`
	} `yaml:"bindings"`
	Inventory struct {
		BeltColumn1 inventory.PotionType `yaml:"beltColumn1"`
		BeltColumn2 inventory.PotionType `yaml:"beltColumn2"`
		BeltColumn3 inventory.PotionType `yaml:"beltColumn3"`
		BeltColumn4 inventory.PotionType `yaml:"beltColumn4"`
	} `yaml:"inventory"`
	Character struct {
		UseMerc bool `yaml:"useMerc"`
	} `yaml:"character"`
	MapAssist struct {
		HostName string `yaml:"hostName"`
	} `yaml:"mapAssist"`
}

// Load reads the config.ini file and returns a Config struct filled with data from the ini file
func Load() (Config, error) {
	r, err := os.Open("config/config.yaml")
	if err != nil {
		return Config{}, fmt.Errorf("error loading config.yaml: %w", err)
	}

	d := yaml.NewDecoder(r)
	cfg := Config{}
	if err = d.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("error reading config: %w", err)
	}

	return cfg, nil
}
