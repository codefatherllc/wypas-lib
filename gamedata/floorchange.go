package gamedata

const (
	FloorchangeDown    uint16 = 1 << 0
	FloorchangeNorth   uint16 = 1 << 1
	FloorchangeSouth   uint16 = 1 << 2
	FloorchangeEast    uint16 = 1 << 3
	FloorchangeWest    uint16 = 1 << 4
	FloorchangeNorthEx uint16 = 1 << 5
	FloorchangeSouthEx uint16 = 1 << 6
	FloorchangeEastEx  uint16 = 1 << 7
	FloorchangeWestEx  uint16 = 1 << 8
)

func HasFloorchange(flags, dir uint16) bool {
	return flags&dir != 0
}
