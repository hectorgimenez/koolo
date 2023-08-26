package event

import (
	"image"

	"github.com/hectorgimenez/koolo/internal/helper"
)

const (
	Kill        Event = "kill"
	Death       Event = "death"
	Chicken     Event = "chicken"
	MercChicken Event = "merc chicken"
	Error       Event = "error"
)

type Event string

var Events = make(chan Message, 10)

type Message struct {
	Message string
	Image   image.Image
}

func WithScreenshot(message string) Message {
	return Message{
		Message: message,
		Image:   helper.Screenshot(),
	}
}

func Text(message string) Message {
	return Message{
		Message: message,
	}
}
