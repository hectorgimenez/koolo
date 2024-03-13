package game

type HID struct {
	gr *MemoryReader
	gi *MemoryInjector
}

func NewHID(gr *MemoryReader, gi *MemoryInjector) *HID {
	return &HID{
		gr: gr,
		gi: gi,
	}
}
