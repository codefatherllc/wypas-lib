package gamedata

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type RichItem struct {
	ServerID    uint16
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
	Count       uint8
	Duration    uint32
	WrittenDate uint32
	WrittenBy   string
	SleeperGUID uint32
	SleepStart  uint32
	SubItems    []RichItem
}

type MapTile struct {
	X         uint16     `db:"x"`
	Y         uint16     `db:"y"`
	Z         uint8      `db:"z"`
	GroundID  uint16     `db:"ground_id"`
	Flags     uint32     `db:"flags"`
	HouseID   uint32     `db:"house_id"`
	Items     []uint16   // decoded from MEDIUMBLOB
	RichItems []RichItem // items with attributes
}

func DecodeItems(blob []byte) ([]uint16, error) {
	if len(blob) == 0 {
		return nil, nil
	}
	if len(blob)%2 != 0 {
		return nil, fmt.Errorf("gamedata: item blob length %d is not even", len(blob))
	}
	items := make([]uint16, len(blob)/2)
	if err := binary.Read(bytes.NewReader(blob), binary.LittleEndian, items); err != nil {
		return nil, err
	}
	return items, nil
}

func EncodeItems(items []uint16) []byte {
	if len(items) == 0 {
		return nil
	}
	// Each item: 2 bytes (uint16 ID) + 1 byte (ATTR_END = 0x00)
	buf := make([]byte, len(items)*3)
	for i, id := range items {
		binary.LittleEndian.PutUint16(buf[i*3:], id)
		buf[i*3+2] = 0x00 // ATTR_END
	}
	return buf
}

type Spawn struct {
	ID        uint32           `db:"id"`
	CenterX   uint16           `db:"center_x"`
	CenterY   uint16           `db:"center_y"`
	CenterZ   uint8            `db:"center_z"`
	Radius    uint8            `db:"radius"`
	Creatures []SpawnCreature
}

type SpawnCreature struct {
	ID        uint32 `db:"id"`
	SpawnID   uint32 `db:"spawn_id"`
	Name      string `db:"name"`
	Type      string `db:"type"`
	OffsetX   int16  `db:"offset_x"`
	OffsetY   int16  `db:"offset_y"`
	OffsetZ   int8   `db:"offset_z"`
	SpawnTime uint32 `db:"spawntime"`
	Direction uint8  `db:"direction"`
}

type Town struct {
	ID     uint32 `db:"id"`
	Name   string `db:"name"`
	X uint16 `db:"x"`
	Y uint16 `db:"y"`
	Z uint8  `db:"z"`
}

type Waypoint struct {
	Name string `db:"name"`
	X    uint16 `db:"x"`
	Y    uint16 `db:"y"`
	Z    uint8  `db:"z"`
}

type House struct {
	ID        uint32 `db:"id"`
	Name      string `db:"name"`
	EntryX    uint16 `db:"entry_x"`
	EntryY    uint16 `db:"entry_y"`
	EntryZ    uint8  `db:"entry_z"`
	Rent      uint32 `db:"rent"`
	TownID    uint32 `db:"town_id"`
	Size      uint16 `db:"size"`
	Guildhall bool   `db:"guildhall"`
}
