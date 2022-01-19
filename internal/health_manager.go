package koolo

import (
	"bytes"
	"context"
	"github.com/otiai10/gosseract/v2"
	"image/png"
	"time"
)

const monitorEvery = time.Millisecond * 500

// HealthManager responsibility is to keep our character and mercenary alive, monitoring life and giving potions when needed
type HealthManager struct {
	display      Display
	tf           TemplateFinder
	lastHeal     time.Time
	lastMana     time.Time
	lastMercHeal time.Time
}

type Status struct {
	Health     int
	MaxHealth  int
	Mana       int
	MaxMana    int
	MercHealth int
}

func NewHealthManager(display Display, tf TemplateFinder) HealthManager {
	return HealthManager{
		display: display,
		tf:      tf,
	}
}

// Start will keep looking at life/mana levels from our character and mercenary and do best effort to keep them up
func (hm HealthManager) Start(ctx context.Context) error {
	ticker := time.NewTicker(monitorEvery)

	for {
		select {
		case <-ticker.C:
			hm.handleHealthAndMana()
		case <-ctx.Done():
			return nil
		}
	}
}

func (hm HealthManager) handleHealthAndMana() {
	status := hm.getStatus()
}

func (hm HealthManager) getStatus() Status {
	img := hm.display.Capture()
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	if err != nil {
	}

	client := gosseract.NewClient()
	defer client.Close()
	err = client.SetImageFromBytes(buf.Bytes())

	return Status{}
}
