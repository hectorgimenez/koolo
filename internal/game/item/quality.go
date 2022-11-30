package item

type Quality int

const (
	ItemQualityNormal   Quality = 0x02
	ItemQualitySuperior Quality = 0x03
	ItemQualityMagic    Quality = 0x04
	ItemQualitySet      Quality = 0x05
	ItemQualityRare     Quality = 0x06
	ItemQualityUnique   Quality = 0x07
)
