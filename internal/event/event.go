package event

import (
	"image"
	"time"
)

type FinishReason string
type InteractionType string

type Event interface {
	Message() string
	Image() image.Image
	OccurredAt() time.Time
	Supervisor() string
}

type BaseEvent struct {
	message    string
	image      image.Image
	occurredAt time.Time
	supervisor string
}

func (b BaseEvent) Message() string {
	return b.message
}

func (b BaseEvent) Image() image.Image {
	return b.image
}

func (b BaseEvent) OccurredAt() time.Time {
	return b.occurredAt
}

func (b BaseEvent) Supervisor() string {
	return b.supervisor
}

func WithScreenshot(supervisor string, message string, img image.Image) BaseEvent {
	return BaseEvent{
		message:    message,
		image:      img,
		occurredAt: time.Now(),
		supervisor: supervisor,
	}
}

func Text(supervisor string, message string) BaseEvent {
	return BaseEvent{
		message:    message,
		occurredAt: time.Now(),
		supervisor: supervisor,
	}
}
