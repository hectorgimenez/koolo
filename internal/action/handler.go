package action

import "context"

type Handler struct {
	actionsChannel <-chan Action
}

func NewHandler(ch <-chan Action) Handler {
	return Handler{
		actionsChannel: ch,
	}
}

func (h *Handler) Listen(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case action := <-h.actionsChannel:
			action.run()
		}
	}
}
