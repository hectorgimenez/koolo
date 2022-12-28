package object

type InteractType uint

const (
	InteractTypeNone   InteractType = 0x00
	InteractTypeTrap   InteractType = 0x04
	InteractTypeLocked InteractType = 0x80
)
