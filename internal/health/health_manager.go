package health

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/game"
)

var ErrDied = errors.New("you died :(")
var ErrChicken = errors.New("chicken")
var ErrMercChicken = errors.New("mercenary chicken")

const (
	healingInterval     = time.Second * 4
	healingMercInterval = time.Second * 6
	manaInterval        = time.Second * 4
	rejuvInterval       = time.Second * 1
)

// Manager responsibility is to keep our character and mercenary alive, monitoring life and giving potions when needed
type Manager struct {
	lastRejuv     time.Time
	lastRejuvMerc time.Time
	lastHeal      time.Time
	lastMana      time.Time
	lastMercHeal  time.Time
	beltManager   *BeltManager
	data          *game.Data
}

func NewHealthManager(bm *BeltManager, data *game.Data) *Manager {
	return &Manager{
		beltManager: bm,
		data:        data,
	}
}

func (hm *Manager) HandleHealthAndMana() error {
	hpConfig := hm.data.CharacterCfg.Health
	// Safe area, skipping
	if hm.data.PlayerUnit.Area.IsTown() {
		return nil
	}

	if hm.data.PlayerUnit.HPPercent() <= 0 {
		return ErrDied
	}

	// Player chicken check
	if hm.data.PlayerUnit.HPPercent() <= hpConfig.ChickenAt {
		return fmt.Errorf("%w: Current Health: %d percent", ErrChicken, hm.data.PlayerUnit.HPPercent())
	}

	// Mercenary chicken check
	if hm.data.MercHPPercent() > 0 && hm.data.MercHPPercent() <= hpConfig.MercChickenAt {
		return fmt.Errorf("%w: Current Merc Health: %d percent", ErrMercChicken, hm.data.MercHPPercent())
	}

	// Player rejuvenation potion check
	if time.Since(hm.lastRejuv) > rejuvInterval &&
		(hm.data.PlayerUnit.HPPercent() <= hpConfig.RejuvPotionAtLife ||
			hm.data.PlayerUnit.MPPercent() < hpConfig.RejuvPotionAtMana) {
		if hm.beltManager.DrinkPotion(data.RejuvenationPotion, false) {
			hm.lastRejuv = time.Now()
			return nil
		}
	}

	// Player healing potion check
	if hm.data.PlayerUnit.HPPercent() <= hpConfig.HealingPotionAt &&
		time.Since(hm.lastHeal) > healingInterval {
		if hm.beltManager.DrinkPotion(data.HealingPotion, false) {
			hm.lastHeal = time.Now()
		}
	}

	// Player mana potion check
	if hm.data.PlayerUnit.MPPercent() <= hpConfig.ManaPotionAt &&
		time.Since(hm.lastMana) > manaInterval {
		if hm.beltManager.DrinkPotion(data.ManaPotion, false) {
			hm.lastMana = time.Now()
		}
	}

	// Mercenary healing logic
	if hm.data.MercHPPercent() > 0 {
		// Mercenary rejuvenation potion check
		if time.Since(hm.lastRejuvMerc) > rejuvInterval &&
			hm.data.MercHPPercent() <= hpConfig.MercRejuvPotionAt {
			if hm.beltManager.DrinkPotion(data.RejuvenationPotion, true) {
				hm.lastRejuvMerc = time.Now()
				return nil
			}
		}

		// Mercenary healing potion check
		if hm.data.MercHPPercent() <= hpConfig.MercHealingPotionAt &&
			time.Since(hm.lastMercHeal) > healingMercInterval {
			if hm.beltManager.DrinkPotion(data.HealingPotion, true) {
				hm.lastMercHeal = time.Now()
			}
		}
	}

	// Player scary aura check - great for use on hardcore
	chickenAurasConfig := hm.data.CharacterCfg.ChickenScaryAuras

	if chickenAurasConfig.AmplifyDamage && hm.data.PlayerUnit.States.HasState(state.Amplifydamage) {
		return fmt.Errorf("%w: Player has amplify damage aura", ErrChicken)
	}

	if chickenAurasConfig.BloodMana && hm.data.PlayerUnit.States.HasState(state.BloodMana) {
		life, _ := hm.data.PlayerUnit.FindStat(stat.Life, 0)
		mana, _ := hm.data.PlayerUnit.FindStat(stat.Mana, 0)

		// Blood mana causes damage to you every time you use mana to cast a spell.
		// The damage is equal to the amount of mana that you used up when you cast the spell.
		// This happens even for non-damaging spells, such as Teleport, or Enchant.

		// This curse only damages yourself when a character's mana exceeds their health (otherwise your def is reduced).
		if mana.Value >= life.Value {
			return fmt.Errorf("%w: Player has blood mana aura", ErrChicken)
		}
	}

	if chickenAurasConfig.Decrepify && hm.data.PlayerUnit.States.HasState(state.Decrepify) {
		return fmt.Errorf("%w: Player has decrepify aura", ErrChicken)
	}

	if chickenAurasConfig.LowerResist && hm.data.PlayerUnit.States.HasState(state.Lowerresist) {
		return fmt.Errorf("%w: Player has lower resist aura", ErrChicken)
	}

	if chickenAurasConfig.Fanaticism || chickenAurasConfig.Might || chickenAurasConfig.Conviction {
		for _, m := range hm.data.Monsters.Enemies() {
			var scaryAura string

			if chickenAurasConfig.Fanaticism && m.States.HasState(state.Fanaticism) {
				scaryAura = "fanaticism"
			}

			if chickenAurasConfig.Might && m.States.HasState(state.Might) {
				scaryAura = "might"
			}

			if chickenAurasConfig.Conviction && m.States.HasState(state.Conviction) {
				scaryAura = "conviction"
			}

			// TODO: Distance check?
			if scaryAura != "" {
				return fmt.Errorf("%w: Mob has %s aura", ErrChicken, scaryAura)
			}
		}
	}

	return nil
}
