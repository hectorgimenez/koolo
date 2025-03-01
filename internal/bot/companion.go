package bot

import (
	"context"
	"log/slog"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
)

// CompanionEventHandler handles events related to companion functionality
type CompanionEventHandler struct {
	supervisor string
	log        *slog.Logger
	cfg        *config.CharacterCfg
}

// NewCompanionEventHandler creates a new instance of CompanionEventHandler
func NewCompanionEventHandler(supervisor string, log *slog.Logger, cfg *config.CharacterCfg) *CompanionEventHandler {
	return &CompanionEventHandler{
		supervisor: supervisor,
		log:        log,
		cfg:        cfg,
	}
}

// Handle processes companion-related events
func (h *CompanionEventHandler) Handle(ctx context.Context, e event.Event) error {

	switch evt := e.(type) {

	case event.RequestCompanionJoinGameEvent:

		if h.cfg.Companion.Enabled && !h.cfg.Companion.Leader {

			// Check if the leader matches the one in our config or no leader set
			if h.cfg.Companion.LeaderName == "" || evt.Leader == h.cfg.Companion.LeaderName {
				h.log.Info("Companion join game event received", slog.String("supervisor", h.supervisor), slog.String("leader", evt.Leader), slog.String("name", evt.Name), slog.String("password", evt.Password))
				h.cfg.Companion.CompanionGameName = evt.Name
				h.cfg.Companion.CompanionGamePassword = evt.Password
			}
		}

	case event.ResetCompanionGameInfoEvent:

		// If this character is a companion (not a leader), clear game info
		if h.cfg.Companion.Enabled && !h.cfg.Companion.Leader {

			// Check if the leader matches the one in our config or no leader set.
			// Additional check for if LeaderName is the same as the character name for Manual join triggers
			if h.cfg.Companion.LeaderName == "" || evt.Leader == h.cfg.Companion.LeaderName || h.cfg.CharacterName == evt.Leader {
				h.log.Info("Companion reset game info event received", slog.String("supervisor", h.supervisor), slog.String("leader", evt.Leader))
				h.cfg.Companion.CompanionGameName = ""
				h.cfg.Companion.CompanionGamePassword = ""
			}
		}
	}

	return nil
}
