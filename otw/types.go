package otw

type WorldMap struct {
	Version     uint32
	Width       uint16
	Height      uint16
	Description string
	TileAreas   []TileArea
	Towns       []Town
	Waypoints   []Waypoint
	Spawns      []Spawn
	Houses      []House
}

type TileArea struct {
	BaseX uint16
	BaseY uint16
	BaseZ uint8
	Tiles []Tile
}

type Tile struct {
	OffsetX   uint8
	OffsetY   uint8
	Flags     uint32
	HouseID   uint32
	Items     []uint16
	RichItems []MapItem
}

type MapItem struct {
	ServerID    uint16
	Count       uint8
	ActionID    uint16
	UniqueID    uint16
	TeleDestX   uint16
	TeleDestY   uint16
	TeleDestZ   uint8
	DoorID      uint8
	DepotID     uint16
	Text        string
	Description string
	Charges     uint16
	RuneCharges uint8
	Duration     uint32
	DecayState   uint8
	WrittenDate  uint32
	WrittenBy   string
	SubItems    []MapItem
}

type Town struct {
	ID      uint32
	Name    string
	TempleX uint16
	TempleY uint16
	TempleZ uint8
}

type Waypoint struct {
	Name string
	X    uint16
	Y    uint16
	Z    uint8
}

type Spawn struct {
	CenterX   uint16
	CenterY   uint16
	CenterZ   uint8
	Radius    int32
	Creatures []SpawnCreature
}

type SpawnCreature struct {
	Name      string
	OffsetX   int16
	OffsetY   int16
	SpawnTime int32
	Direction uint8
}

type House struct {
	ID     uint32
	Name   string
	EntryX uint16
	EntryY uint16
	EntryZ uint8
	Size   int32
	Rent   int32
	TownID int32
}
