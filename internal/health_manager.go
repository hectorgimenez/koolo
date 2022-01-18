package koolo

import "context"

// HealthManager responsibility is to keep our character and mercenary alive, monitoring life and giving potions when needed
type HealthManager struct {
}

func NewHealthManager(Display, TemplateFinder) HealthManager {
	return HealthManager{}
}

// Start will keep looking at life/mana levels from our character and mercenary and do best effort to keep them up
func (hm HealthManager) Start(ctx context.Context) error {
	return nil
}
