package otbm

import "encoding/xml"

type xmlSpawns struct {
	XMLName xml.Name   `xml:"spawns"`
	Spawns  []xmlSpawn `xml:"spawn"`
}

type xmlSpawn struct {
	CenterX  uint16       `xml:"centerx,attr"`
	CenterY  uint16       `xml:"centery,attr"`
	CenterZ  uint8        `xml:"centerz,attr"`
	Radius   uint8        `xml:"radius,attr"`
	Monsters []xmlMonster `xml:"monster"`
	NPCs     []xmlNPC     `xml:"npc"`
}

type xmlMonster struct {
	Name      string `xml:"name,attr"`
	X         int16  `xml:"x,attr"`
	Y         int16  `xml:"y,attr"`
	Z         int8   `xml:"z,attr"`
	SpawnTime uint32 `xml:"spawntime,attr"`
	Direction uint8  `xml:"direction,attr,omitempty"`
}

type xmlNPC struct {
	Name      string `xml:"name,attr"`
	X         int16  `xml:"x,attr"`
	Y         int16  `xml:"y,attr"`
	Z         int8   `xml:"z,attr"`
	SpawnTime uint32 `xml:"spawntime,attr"`
	Direction uint8  `xml:"direction,attr,omitempty"`
}

type xmlHouses struct {
	XMLName xml.Name   `xml:"houses"`
	Houses  []xmlHouse `xml:"house"`
}

type xmlHouse struct {
	ID        uint32 `xml:"houseid,attr"`
	Name      string `xml:"name,attr"`
	EntryX    uint16 `xml:"entryx,attr"`
	EntryY    uint16 `xml:"entryy,attr"`
	EntryZ    uint8  `xml:"entryz,attr"`
	Rent      uint32 `xml:"rent,attr"`
	TownID    uint32 `xml:"townid,attr"`
	Size      uint16 `xml:"size,attr"`
	Guildhall bool   `xml:"guildhall,attr,omitempty"`
}
