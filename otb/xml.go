package otb

import "encoding/xml"

type xmlItems struct {
	Items []xmlItem `xml:"item"`
}

type xmlItem struct {
	ID      uint16    `xml:"id,attr"`
	FromID  uint16    `xml:"fromid,attr"`
	ToID    uint16    `xml:"toid,attr"`
	Name    string    `xml:"name,attr"`
	Article string    `xml:"article,attr"`
	Plural  string    `xml:"plural,attr"`
	Attrs   []xmlAttr `xml:"attribute"`
}

type xmlAttr struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

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
}

type xmlNPC struct {
	Name      string `xml:"name,attr"`
	X         int16  `xml:"x,attr"`
	Y         int16  `xml:"y,attr"`
	Z         int8   `xml:"z,attr"`
	SpawnTime uint32 `xml:"spawntime,attr"`
}

type xmlHouses struct {
	XMLName xml.Name   `xml:"houses"`
	Houses  []xmlHouse `xml:"house"`
}

type xmlHouse struct {
	ID     uint32 `xml:"houseid,attr"`
	Name   string `xml:"name,attr"`
	EntryX uint16 `xml:"entryx,attr"`
	EntryY uint16 `xml:"entryy,attr"`
	EntryZ uint8  `xml:"entryz,attr"`
	Rent   uint32 `xml:"rent,attr"`
	TownID uint32 `xml:"townid,attr"`
	Size   uint16 `xml:"size,attr"`
}
