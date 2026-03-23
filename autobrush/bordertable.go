package autobrush

const (
	BorderNone         = 0
	NorthHorizontal    = 1
	EastHorizontal     = 2
	SouthHorizontal    = 3
	WestHorizontal     = 4
	NorthwestCorner    = 5
	NortheastCorner    = 6
	SouthwestCorner    = 7
	SoutheastCorner    = 8
	NorthwestDiagonal  = 9
	NortheastDiagonal  = 10
	SoutheastDiagonal  = 11
	SouthwestDiagonal  = 12
	CarpetCenter       = 13

	TileNorthwest = 1
	TileNorth     = 2
	TileNortheast = 4
	TileWest      = 8
	TileEast      = 16
	TileSouthwest = 32
	TileSouth     = 64
	TileSoutheast = 128

	WallPole              = 0
	WallSouthEnd          = 1
	WallEastEnd           = 2
	WallNorthwestDiagonal = 3
	WallWestEnd           = 4
	WallNortheastDiagonal = 5
	WallHorizontal        = 6
	WallSouthT            = 7
	WallNorthEnd          = 8
	WallVertical          = 9
	WallSouthwestDiagonal = 10
	WallEastT             = 11
	WallSoutheastDiagonal = 12
	WallWestT             = 13
	WallNorthT            = 14
	WallIntersection      = 15
	WallUntouchable       = 16

	WalltileNorth = 1
	WalltileWest  = 2
	WalltileEast  = 4
	WalltileSouth = 8

	TableNorthEnd  = 0
	TableSouthEnd  = 1
	TableEastEnd   = 2
	TableWestEnd   = 3
	TableHorizontal = 4
	TableVertical   = 5
	TableAlone      = 6
)

var groundBorderTypes [256]uint32
var wallFullBorderTypes [16]uint32
var wallHalfBorderTypes [16]uint32

func init() {
	initGroundBorderTypes()
	initWallBorderTypes()
}

func pack(a, b, c, d int) uint32 {
	return uint32(a) | uint32(b)<<8 | uint32(c)<<16 | uint32(d)<<24
}

func unpackDirections(packed uint32) [4]int {
	return [4]int{
		int(packed & 0xFF),
		int((packed >> 8) & 0xFF),
		int((packed >> 16) & 0xFF),
		int((packed >> 24) & 0xFF),
	}
}

func initGroundBorderTypes() {
	t := &groundBorderTypes
	NW := TileNorthwest
	N := TileNorth
	NE := TileNortheast
	W := TileWest
	E := TileEast
	SW := TileSouthwest
	S := TileSouth
	SE := TileSoutheast

	t[0] = uint32(BorderNone)
	t[NW] = uint32(NorthwestCorner)
	t[N] = uint32(NorthHorizontal)
	t[N|NW] = uint32(NorthHorizontal)
	t[NE] = uint32(NortheastCorner)
	t[NE|NW] = pack(NorthwestCorner, NortheastCorner, 0, 0)
	t[NE|N] = uint32(NorthHorizontal)
	t[NE|N|NW] = uint32(NorthHorizontal)
	t[W] = uint32(WestHorizontal)
	t[W|NW] = uint32(WestHorizontal)
	t[W|N] = uint32(NorthwestDiagonal)
	t[W|N|NW] = uint32(NorthwestDiagonal)
	t[W|NE] = pack(WestHorizontal, NortheastCorner, 0, 0)
	t[W|NE|NW] = pack(WestHorizontal, NortheastCorner, 0, 0)
	t[W|NE|N] = uint32(NorthwestDiagonal)
	t[W|NE|N|NW] = uint32(NorthwestDiagonal)
	t[E] = uint32(EastHorizontal)
	t[E|NW] = pack(NorthwestCorner, EastHorizontal, 0, 0)
	t[E|N] = uint32(NortheastDiagonal)
	t[E|N|NW] = uint32(NortheastDiagonal)
	t[E|NE] = uint32(EastHorizontal)
	t[E|NE|NW] = pack(NorthwestCorner, EastHorizontal, 0, 0)
	t[E|NE|N] = uint32(NortheastDiagonal)
	t[E|NE|N|NW] = uint32(NortheastDiagonal)
	t[E|W] = pack(WestHorizontal, EastHorizontal, 0, 0)
	t[E|W|NW] = pack(WestHorizontal, EastHorizontal, 0, 0)
	t[E|W|N] = pack(NorthHorizontal, WestHorizontal, EastHorizontal, 0)
	t[E|W|N|NW] = pack(NorthHorizontal, WestHorizontal, EastHorizontal, 0)
	t[E|W|NE] = pack(EastHorizontal, WestHorizontal, 0, 0)
	t[E|W|NE|NW] = pack(EastHorizontal, WestHorizontal, 0, 0)
	t[E|W|NE|N] = pack(NorthHorizontal, EastHorizontal, WestHorizontal, 0)
	t[E|W|NE|N|NW] = pack(NorthHorizontal, EastHorizontal, WestHorizontal, 0)

	t[SW] = uint32(SouthwestCorner)
	t[SW|NW] = pack(SouthwestCorner, NorthwestCorner, 0, 0)
	t[SW|N] = pack(SouthwestCorner, NorthHorizontal, 0, 0)
	t[SW|N|NW] = pack(SouthwestCorner, NorthHorizontal, 0, 0)
	t[SW|NE] = pack(SouthwestCorner, NortheastCorner, 0, 0)
	t[SW|NE|NW] = pack(SouthwestCorner, NortheastCorner, NorthwestCorner, 0)
	t[SW|NE|N] = pack(SouthwestCorner, NorthHorizontal, 0, 0)
	t[SW|NE|N|NW] = pack(SouthwestCorner, NorthHorizontal, 0, 0)
	t[SW|W] = uint32(WestHorizontal)
	t[SW|W|NW] = uint32(WestHorizontal)
	t[SW|W|N] = uint32(NorthwestDiagonal)
	t[SW|W|N|NW] = uint32(NorthwestDiagonal)
	t[SW|W|NE] = pack(WestHorizontal, NortheastCorner, 0, 0)
	t[SW|W|NE|NW] = pack(WestHorizontal, NortheastCorner, 0, 0)
	t[SW|W|NE|N] = uint32(NorthwestDiagonal)
	t[SW|W|NE|N|NW] = uint32(NorthwestDiagonal)
	t[SW|E] = pack(SouthwestCorner, EastHorizontal, 0, 0)
	t[SW|E|NW] = pack(SouthwestCorner, EastHorizontal, NorthwestCorner, 0)
	t[SW|E|N] = pack(SouthwestCorner, NortheastDiagonal, 0, 0)
	t[SW|E|N|NW] = pack(SouthwestCorner, NortheastDiagonal, 0, 0)
	t[SW|E|NE] = pack(SouthwestCorner, EastHorizontal, 0, 0)
	t[SW|E|NE|NW] = pack(SouthwestCorner, EastHorizontal, NorthwestCorner, 0)
	t[SW|E|NE|N] = pack(SouthwestCorner, NortheastDiagonal, 0, 0)
	t[SW|E|NE|N|NW] = pack(SouthwestCorner, NortheastDiagonal, 0, 0)
	t[SW|E|W] = pack(WestHorizontal, EastHorizontal, 0, 0)
	t[SW|E|W|NW] = pack(WestHorizontal, EastHorizontal, 0, 0)
	t[SW|E|W|N] = pack(WestHorizontal, EastHorizontal, NorthHorizontal, 0)
	t[SW|E|W|N|NW] = pack(WestHorizontal, EastHorizontal, NorthHorizontal, 0)
	t[SW|E|W|NE] = pack(WestHorizontal, EastHorizontal, 0, 0)
	t[SW|E|W|NE|NW] = pack(WestHorizontal, EastHorizontal, 0, 0)
	t[SW|E|W|NE|N] = pack(WestHorizontal, EastHorizontal, NorthHorizontal, 0)
	t[SW|E|W|NE|N|NW] = pack(WestHorizontal, EastHorizontal, NorthHorizontal, 0)

	t[S] = uint32(SouthHorizontal)
	t[S|NW] = pack(SouthHorizontal, NorthwestCorner, 0, 0)
	t[S|N] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[S|N|NW] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[S|NE] = pack(SouthHorizontal, NortheastCorner, 0, 0)
	t[S|NE|NW] = pack(SouthHorizontal, NortheastCorner, NorthwestCorner, 0)
	t[S|NE|N] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[S|NE|N|NW] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[S|W] = uint32(SouthwestDiagonal)
	t[S|W|NW] = uint32(SouthwestDiagonal)
	t[S|W|N] = pack(SouthHorizontal, NorthHorizontal, WestHorizontal, 0)
	t[S|W|N|NW] = pack(SouthHorizontal, NorthHorizontal, WestHorizontal, 0)
	t[S|W|NE] = pack(SouthwestDiagonal, NortheastCorner, 0, 0)
	t[S|W|NE|NW] = pack(SouthwestDiagonal, NortheastCorner, 0, 0)
	t[S|W|NE|N] = pack(SouthHorizontal, NorthHorizontal, WestHorizontal, 0)
	t[S|W|NE|N|NW] = pack(SouthHorizontal, NorthHorizontal, WestHorizontal, 0)
	t[S|E] = uint32(SoutheastDiagonal)
	t[S|E|NW] = pack(SoutheastDiagonal, NorthwestCorner, 0, 0)
	t[S|E|N] = pack(SouthHorizontal, NorthHorizontal, EastHorizontal, 0)
	t[S|E|N|NW] = pack(SouthHorizontal, NorthHorizontal, EastHorizontal, 0)
	t[S|E|NE] = uint32(SoutheastDiagonal)
	t[S|E|NE|NW] = pack(SoutheastDiagonal, NorthwestCorner, 0, 0)
	t[S|E|NE|N] = pack(SouthHorizontal, NorthHorizontal, EastHorizontal, 0)
	t[S|E|NE|N|NW] = pack(SouthHorizontal, NorthHorizontal, EastHorizontal, 0)
	t[S|E|W] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, 0)
	t[S|E|W|NW] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, 0)
	t[S|E|W|N] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, NorthHorizontal)
	t[S|E|W|N|NW] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, NorthHorizontal)
	t[S|E|W|NE] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, 0)
	t[S|E|W|NE|NW] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, 0)
	t[S|E|W|NE|N] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, NorthHorizontal)
	t[S|E|W|NE|N|NW] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, NorthHorizontal)

	t[S|SW] = uint32(SouthHorizontal)
	t[S|SW|NW] = pack(SouthHorizontal, NorthwestCorner, 0, 0)
	t[S|SW|N] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[S|SW|N|NW] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[S|SW|NE] = pack(SouthHorizontal, NortheastCorner, 0, 0)
	t[S|SW|NE|NW] = pack(SouthHorizontal, NorthwestCorner, NortheastCorner, 0)
	t[S|SW|NE|N] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[S|SW|NE|N|NW] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[S|SW|W] = uint32(SouthwestDiagonal)
	t[S|SW|W|NW] = uint32(SouthwestDiagonal)
	t[S|SW|W|N] = pack(SouthHorizontal, WestHorizontal, NorthHorizontal, 0)
	t[S|SW|W|N|NW] = pack(SouthHorizontal, WestHorizontal, NorthHorizontal, 0)
	t[S|SW|W|NE] = pack(SouthwestDiagonal, NortheastCorner, 0, 0)
	t[S|SW|W|NE|NW] = pack(SouthwestDiagonal, NortheastCorner, 0, 0)
	t[S|SW|W|NE|N] = pack(SouthHorizontal, WestHorizontal, NorthHorizontal, 0)
	t[S|SW|W|NE|N|NW] = pack(SouthHorizontal, WestHorizontal, NorthHorizontal, 0)
	t[S|SW|E] = uint32(SoutheastDiagonal)
	t[S|SW|E|NW] = pack(SoutheastDiagonal, NorthwestCorner, 0, 0)
	t[S|SW|E|N] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, 0)
	t[S|SW|E|N|NW] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, 0)
	t[S|SW|E|NE] = uint32(SoutheastDiagonal)
	t[S|SW|E|NE|NW] = pack(SoutheastDiagonal, NorthwestCorner, 0, 0)
	t[S|SW|E|NE|N] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, 0)
	t[S|SW|E|NE|N|NW] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, 0)
	t[S|SW|E|W] = pack(SouthHorizontal, EastHorizontal, WestHorizontal, 0)
	t[S|SW|E|W|NW] = pack(SouthHorizontal, EastHorizontal, WestHorizontal, 0)
	t[S|SW|E|W|N] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, WestHorizontal)
	t[S|SW|E|W|N|NW] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, WestHorizontal)
	t[S|SW|E|W|NE] = pack(SouthHorizontal, EastHorizontal, WestHorizontal, 0)
	t[S|SW|E|W|NE|NW] = pack(SouthHorizontal, EastHorizontal, WestHorizontal, 0)
	t[S|SW|E|W|NE|N] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, WestHorizontal)
	t[S|SW|E|W|NE|N|NW] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, WestHorizontal)

	// SE combinations
	t[SE] = uint32(SoutheastCorner)
	t[SE|NW] = pack(NorthwestCorner, SoutheastCorner, 0, 0)
	t[SE|N] = pack(NorthHorizontal, SoutheastCorner, 0, 0)
	t[SE|N|NW] = pack(NorthHorizontal, SoutheastCorner, 0, 0)
	t[SE|NE] = pack(NortheastCorner, SoutheastCorner, 0, 0)
	t[SE|NE|NW] = pack(NortheastCorner, NorthwestCorner, SoutheastCorner, 0)
	t[SE|NE|N] = pack(NorthHorizontal, SoutheastCorner, 0, 0)
	t[SE|NE|N|NW] = pack(NorthHorizontal, SoutheastCorner, 0, 0)
	t[SE|W] = pack(WestHorizontal, SoutheastCorner, 0, 0)
	t[SE|W|NW] = pack(WestHorizontal, SoutheastCorner, 0, 0)
	t[SE|W|N] = pack(NorthwestDiagonal, SoutheastCorner, 0, 0)
	t[SE|W|N|NW] = pack(NorthwestDiagonal, SoutheastCorner, 0, 0)
	t[SE|W|NE] = pack(WestHorizontal, NortheastCorner, SoutheastCorner, 0)
	t[SE|W|NE|NW] = pack(WestHorizontal, NortheastCorner, SoutheastCorner, 0)
	t[SE|W|NE|N] = pack(NorthwestDiagonal, SoutheastCorner, 0, 0)
	t[SE|W|NE|N|NW] = pack(NorthwestDiagonal, SoutheastCorner, 0, 0)
	t[SE|E] = uint32(EastHorizontal)
	t[SE|E|NW] = pack(EastHorizontal, NorthwestCorner, 0, 0)
	t[SE|E|N] = uint32(NortheastDiagonal)
	t[SE|E|N|NW] = uint32(NortheastDiagonal)
	t[SE|E|NE] = uint32(EastHorizontal)
	t[SE|E|NE|NW] = pack(EastHorizontal, NorthwestCorner, 0, 0)
	t[SE|E|NE|N] = uint32(NortheastDiagonal)
	t[SE|E|NE|N|NW] = uint32(NortheastDiagonal)
	t[SE|E|W] = pack(EastHorizontal, WestHorizontal, 0, 0)
	t[SE|E|W|NW] = pack(EastHorizontal, WestHorizontal, 0, 0)
	t[SE|E|W|N] = pack(NorthHorizontal, EastHorizontal, WestHorizontal, 0)
	t[SE|E|W|N|NW] = pack(EastHorizontal, WestHorizontal, NorthHorizontal, 0)
	t[SE|E|W|NE] = pack(EastHorizontal, WestHorizontal, 0, 0)
	t[SE|E|W|NE|NW] = pack(EastHorizontal, WestHorizontal, 0, 0)
	t[SE|E|W|NE|N] = pack(NorthHorizontal, EastHorizontal, WestHorizontal, 0)
	t[SE|E|W|NE|N|NW] = pack(NorthHorizontal, EastHorizontal, WestHorizontal, 0)

	t[SE|SW] = pack(SouthwestCorner, SoutheastCorner, 0, 0)
	t[SE|SW|NW] = pack(SouthwestCorner, NorthwestCorner, SoutheastCorner, 0)
	t[SE|SW|N] = pack(SouthwestCorner, NorthHorizontal, SoutheastCorner, 0)
	t[SE|SW|N|NW] = pack(SouthwestCorner, NorthHorizontal, SoutheastCorner, 0)
	t[SE|SW|NE] = pack(SouthwestCorner, NortheastCorner, SoutheastCorner, 0)
	t[SE|SW|NE|NW] = pack(SouthwestCorner, NortheastCorner, NorthwestCorner, SoutheastCorner)
	t[SE|SW|NE|N] = pack(SouthwestCorner, NorthHorizontal, SoutheastCorner, 0)
	t[SE|SW|NE|N|NW] = pack(SouthwestCorner, NorthHorizontal, SoutheastCorner, 0)
	t[SE|SW|W] = pack(WestHorizontal, SoutheastCorner, 0, 0)
	t[SE|SW|W|NW] = pack(WestHorizontal, SoutheastCorner, 0, 0)
	t[SE|SW|W|N] = pack(NorthwestDiagonal, SoutheastCorner, 0, 0)
	t[SE|SW|W|N|NW] = pack(NorthwestDiagonal, SoutheastCorner, 0, 0)
	t[SE|SW|W|NE] = pack(WestHorizontal, NortheastCorner, SoutheastCorner, 0)
	t[SE|SW|W|NE|NW] = pack(WestHorizontal, NortheastCorner, SoutheastCorner, 0)
	t[SE|SW|W|NE|N] = pack(NorthwestDiagonal, SoutheastCorner, 0, 0)
	t[SE|SW|W|NE|N|NW] = pack(NorthwestDiagonal, SoutheastCorner, 0, 0)
	t[SE|SW|E] = pack(SouthwestCorner, EastHorizontal, 0, 0)
	t[SE|SW|E|NW] = pack(SouthwestCorner, EastHorizontal, NorthwestCorner, 0)
	t[SE|SW|E|N] = pack(SouthwestCorner, NortheastDiagonal, 0, 0)
	t[SE|SW|E|N|NW] = pack(SouthwestCorner, NortheastDiagonal, 0, 0)
	t[SE|SW|E|NE] = pack(SouthwestCorner, EastHorizontal, 0, 0)
	t[SE|SW|E|NE|NW] = pack(SouthwestCorner, EastHorizontal, NorthwestCorner, 0)
	t[SE|SW|E|NE|N] = pack(SouthwestCorner, NortheastDiagonal, 0, 0)
	t[SE|SW|E|NE|N|NW] = pack(SouthwestCorner, NortheastDiagonal, 0, 0)
	t[SE|SW|E|W] = pack(WestHorizontal, EastHorizontal, 0, 0)
	t[SE|SW|E|W|NW] = pack(WestHorizontal, EastHorizontal, 0, 0)
	t[SE|SW|E|W|N] = pack(WestHorizontal, EastHorizontal, NorthHorizontal, 0)
	t[SE|SW|E|W|N|NW] = pack(WestHorizontal, EastHorizontal, NorthHorizontal, 0)
	t[SE|SW|E|W|NE] = pack(WestHorizontal, EastHorizontal, 0, 0)
	t[SE|SW|E|W|NE|NW] = pack(WestHorizontal, EastHorizontal, 0, 0)
	t[SE|SW|E|W|NE|N] = pack(WestHorizontal, EastHorizontal, NorthHorizontal, 0)
	t[SE|SW|E|W|NE|N|NW] = pack(WestHorizontal, EastHorizontal, NorthHorizontal, 0)

	t[SE|S] = uint32(SouthHorizontal)
	t[SE|S|NW] = pack(SouthHorizontal, NorthwestCorner, 0, 0)
	t[SE|S|N] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[SE|S|N|NW] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[SE|S|NE] = pack(SouthHorizontal, NortheastCorner, 0, 0)
	t[SE|S|NE|NW] = pack(SouthHorizontal, NortheastCorner, NorthwestCorner, 0)
	t[SE|S|NE|N] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[SE|S|NE|N|NW] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[SE|S|W] = uint32(SouthwestDiagonal)
	t[SE|S|W|NW] = uint32(SouthwestDiagonal)
	t[SE|S|W|N] = pack(SouthHorizontal, NorthHorizontal, WestHorizontal, 0)
	t[SE|S|W|N|NW] = pack(SouthHorizontal, NorthHorizontal, WestHorizontal, 0)
	t[SE|S|W|NE] = pack(SouthwestDiagonal, NortheastCorner, 0, 0)
	t[SE|S|W|NE|NW] = pack(SouthwestDiagonal, NortheastCorner, 0, 0)
	t[SE|S|W|NE|N] = pack(SouthHorizontal, NorthHorizontal, WestHorizontal, 0)
	t[SE|S|W|NE|N|NW] = pack(SouthHorizontal, NorthHorizontal, WestHorizontal, 0)
	t[SE|S|E] = uint32(SoutheastDiagonal)
	t[SE|S|E|NW] = pack(SoutheastDiagonal, NorthwestCorner, 0, 0)
	t[SE|S|E|N] = pack(SouthHorizontal, NorthHorizontal, EastHorizontal, 0)
	t[SE|S|E|N|NW] = pack(SouthHorizontal, NorthHorizontal, EastHorizontal, 0)
	t[SE|S|E|NE] = uint32(SoutheastDiagonal)
	t[SE|S|E|NE|NW] = pack(SoutheastDiagonal, NorthwestCorner, 0, 0)
	t[SE|S|E|NE|N] = pack(SouthHorizontal, NorthHorizontal, EastHorizontal, 0)
	t[SE|S|E|NE|N|NW] = pack(SouthHorizontal, NorthHorizontal, EastHorizontal, 0)
	t[SE|S|E|W] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, 0)
	t[SE|S|E|W|NW] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, 0)
	t[SE|S|E|W|N] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, NorthHorizontal)
	t[SE|S|E|W|N|NW] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, NorthHorizontal)
	t[SE|S|E|W|NE] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, 0)
	t[SE|S|E|W|NE|NW] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, 0)
	t[SE|S|E|W|NE|N] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, NorthHorizontal)
	t[SE|S|E|W|NE|N|NW] = pack(SouthHorizontal, WestHorizontal, EastHorizontal, NorthHorizontal)

	t[SE|S|SW] = uint32(SouthHorizontal)
	t[SE|S|SW|NW] = pack(SouthHorizontal, NorthwestCorner, 0, 0)
	t[SE|S|SW|N] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[SE|S|SW|N|NW] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[SE|S|SW|NE] = pack(SouthHorizontal, NortheastCorner, 0, 0)
	t[SE|S|SW|NE|NW] = pack(SouthHorizontal, NorthwestCorner, NortheastCorner, 0)
	t[SE|S|SW|NE|N] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[SE|S|SW|NE|N|NW] = pack(SouthHorizontal, NorthHorizontal, 0, 0)
	t[SE|S|SW|W] = uint32(SouthwestDiagonal)
	t[SE|S|SW|W|NW] = uint32(SouthwestDiagonal)
	t[SE|S|SW|W|N] = pack(SouthHorizontal, WestHorizontal, NorthHorizontal, 0)
	t[SE|S|SW|W|N|NW] = pack(SouthHorizontal, WestHorizontal, NorthHorizontal, 0)
	t[SE|S|SW|W|NE] = pack(SouthwestDiagonal, NortheastCorner, 0, 0)
	t[SE|S|SW|W|NE|NW] = pack(SouthwestDiagonal, NortheastCorner, 0, 0)
	t[SE|S|SW|W|NE|N] = pack(SouthHorizontal, WestHorizontal, NorthHorizontal, 0)
	t[SE|S|SW|W|NE|N|NW] = pack(SouthHorizontal, WestHorizontal, NorthHorizontal, 0)
	t[SE|S|SW|E] = uint32(SoutheastDiagonal)
	t[SE|S|SW|E|NW] = pack(SoutheastDiagonal, NorthwestCorner, 0, 0)
	t[SE|S|SW|E|N] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, 0)
	t[SE|S|SW|E|N|NW] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, 0)
	t[SE|S|SW|E|NE] = uint32(SoutheastDiagonal)
	t[SE|S|SW|E|NE|NW] = pack(SoutheastDiagonal, NorthwestCorner, 0, 0)
	t[SE|S|SW|E|NE|N] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, 0)
	t[SE|S|SW|E|NE|N|NW] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, 0)
	t[SE|S|SW|E|W] = pack(SouthHorizontal, EastHorizontal, WestHorizontal, 0)
	t[SE|S|SW|E|W|NW] = pack(SouthHorizontal, EastHorizontal, WestHorizontal, 0)
	t[SE|S|SW|E|W|N] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, WestHorizontal)
	t[SE|S|SW|E|W|N|NW] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, WestHorizontal)
	t[SE|S|SW|E|W|NE] = pack(SouthHorizontal, EastHorizontal, WestHorizontal, 0)
	t[SE|S|SW|E|W|NE|NW] = pack(SouthHorizontal, EastHorizontal, WestHorizontal, 0)
	t[SE|S|SW|E|W|NE|N] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, WestHorizontal)
	t[SE|S|SW|E|W|NE|N|NW] = pack(SouthHorizontal, EastHorizontal, NorthHorizontal, WestHorizontal)
}

func initWallBorderTypes() {
	f := &wallFullBorderTypes
	f[0] = WallPole
	f[WalltileNorth] = WallSouthEnd
	f[WalltileWest] = WallEastEnd
	f[WalltileWest|WalltileNorth] = WallNorthwestDiagonal
	f[WalltileEast] = WallWestEnd
	f[WalltileEast|WalltileNorth] = WallNortheastDiagonal
	f[WalltileEast|WalltileWest] = WallHorizontal
	f[WalltileEast|WalltileWest|WalltileNorth] = WallSouthT
	f[WalltileSouth] = WallNorthEnd
	f[WalltileSouth|WalltileNorth] = WallVertical
	f[WalltileSouth|WalltileWest] = WallSouthwestDiagonal
	f[WalltileSouth|WalltileWest|WalltileNorth] = WallEastT
	f[WalltileSouth|WalltileEast] = WallSoutheastDiagonal
	f[WalltileSouth|WalltileEast|WalltileNorth] = WallWestT
	f[WalltileSouth|WalltileEast|WalltileWest] = WallNorthT
	f[WalltileSouth|WalltileEast|WalltileWest|WalltileNorth] = WallIntersection

	h := &wallHalfBorderTypes
	h[0] = WallPole
	h[WalltileNorth] = WallVertical
	h[WalltileWest] = WallHorizontal
	h[WalltileWest|WalltileNorth] = WallNorthwestDiagonal
	h[WalltileEast] = WallPole
	h[WalltileEast|WalltileNorth] = WallVertical
	h[WalltileEast|WalltileWest] = WallHorizontal
	h[WalltileEast|WalltileWest|WalltileNorth] = WallNorthwestDiagonal
	h[WalltileSouth] = WallPole
	h[WalltileSouth|WalltileNorth] = WallVertical
	h[WalltileSouth|WalltileWest] = WallHorizontal
	h[WalltileSouth|WalltileWest|WalltileNorth] = WallNorthwestDiagonal
	h[WalltileSouth|WalltileEast] = WallPole
	h[WalltileSouth|WalltileEast|WalltileNorth] = WallVertical
	h[WalltileSouth|WalltileEast|WalltileWest] = WallHorizontal
	h[WalltileSouth|WalltileEast|WalltileWest|WalltileNorth] = WallNorthwestDiagonal
}
