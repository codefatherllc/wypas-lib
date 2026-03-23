package gamedata

const (
	FloorchangeDown    uint8 = 1 << 0
	FloorchangeNorth   uint8 = 1 << 1
	FloorchangeSouth   uint8 = 1 << 2
	FloorchangeEast    uint8 = 1 << 3
	FloorchangeWest    uint8 = 1 << 4
	FloorchangeNorthEx uint8 = 1 << 5
	FloorchangeSouthEx uint8 = 1 << 6
	FloorchangeWestEx  uint8 = 1 << 7
)

func HasFloorchange(flags, dir uint8) bool {
	return flags&dir != 0
}
