package otbm

import (
	"fmt"
	"math"
)

const (
	NodeMapData   = 2
	NodeTileArea  = 4
	NodeTile      = 5
	NodeItem      = 6
	NodeTowns     = 12
	NodeTown      = 13
	NodeHouseTile = 14
	NodeWaypoints = 15
	NodeWaypoint  = 16
)

const (
	AttrDescription = 1
	AttrExtFile     = 2
	AttrTileFlags   = 3
	AttrActionID    = 4
	AttrUniqueID    = 5
	AttrText        = 6
	AttrDesc        = 7
	AttrTeleDest    = 8
	AttrItem        = 9
	AttrDepotID     = 10
	AttrSpawnFile   = 11
	AttrRuneCharges = 12
	AttrHouseFile   = 13
	AttrDoorID      = 14
	AttrCount       = 15
	AttrDuration    = 16
	AttrDecayState  = 17
	AttrWrittenDate = 18
	AttrWrittenBy   = 19
	AttrSleeperGUID = 20
	AttrSleepStart  = 21
	AttrCharges        = 22
	AttrContainerItems = 23
	AttrAttrMap        = 128
)

const (
	TileFlagProtectionZone = 0x0001
	TileFlagNoPVP          = 0x0004
	TileFlagNoLogout       = 0x0008
	TileFlagPVPZone        = 0x0010
	TileFlagRefresh        = 0x0020
)

type TeleportDest struct {
	X, Y uint16
	Z    uint8
}

type ItemAttrValue struct {
	Type    uint8 // 1=string, 2=int32, 3=float, 4=bool
	Str     string
	Int     int32
	Float   float32
	Boolean bool
}

type MapItem struct {
	ID           uint16
	ActionID     uint16
	UniqueID     uint16
	TeleDest     *TeleportDest
	DoorID       uint8
	DepotID      uint16
	Text         string
	Description  string
	Charges      uint16
	RuneCharges  uint8
	Count        uint8
	Duration     uint32
	WrittenDate  uint32
	WrittenBy    string
	SleeperGUID  uint32
	SleepStart   uint32
	CustomAttrs  map[string]ItemAttrValue
	SubItems     []MapItem
}

type MapTile struct {
	Items     []uint16
	RichItems []MapItem
	Flags     uint32
	HouseID   uint32
}

type Town struct {
	ID   uint32
	Name string
	X, Y uint16
	Z    uint8
}

type Waypoint struct {
	Name string
	X, Y uint16
	Z    uint8
}

type FloorBounds struct {
	MinX, MinY, MaxX, MaxY uint16
}

type GameMap struct {
	Version      uint32
	Width        uint16
	Height       uint16
	Descriptions []string
	SpawnFile    string
	HouseFile    string
	Tiles        map[uint64]*MapTile
	MinX, MinY   uint16
	MaxX, MaxY   uint16
	Floors       []uint8
	Towns        []Town
	Waypoints    []Waypoint
	FloorBounds  map[uint8]*FloorBounds
}

func PackPos(x, y uint16, z uint8) uint64 {
	return uint64(x)<<24 | uint64(y)<<8 | uint64(z)
}

func ParseOTBM(path string) (*GameMap, error) {
	root, err := ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("parse otbm tree: %w", err)
	}

	root.ResetPos()
	version, _ := root.GetU32()
	width, _ := root.GetU16()
	height, _ := root.GetU16()
	if err := root.Skip(4 + 4); err != nil {
		return nil, fmt.Errorf("skip root header items version: %w", err)
	}

	gm := &GameMap{
		Version:     version,
		Width:       width,
		Height:      height,
		Tiles:       make(map[uint64]*MapTile),
		MinX:        math.MaxUint16,
		MinY:        math.MaxUint16,
		FloorBounds: make(map[uint8]*FloorBounds),
	}
	floorSet := make(map[uint8]bool)

	for _, mapDataNode := range root.Children {
		if mapDataNode.Type != NodeMapData {
			continue
		}

		mapDataNode.ResetPos()
		for mapDataNode.Remaining() > 0 {
			attr, err := mapDataNode.GetU8()
			if err != nil {
				break
			}
			switch attr {
			case AttrDescription:
				desc, err := mapDataNode.GetString()
				if err != nil {
					break
				}
				gm.Descriptions = append(gm.Descriptions, desc)
			case AttrSpawnFile:
				sf, err := mapDataNode.GetString()
				if err != nil {
					break
				}
				gm.SpawnFile = sf
			case AttrHouseFile:
				hf, err := mapDataNode.GetString()
				if err != nil {
					break
				}
				gm.HouseFile = hf
			default:
				goto doneMapAttrs
			}
		}
	doneMapAttrs:

		for _, child := range mapDataNode.Children {
			switch child.Type {
			case NodeTileArea:
				if err := parseTileArea(child, gm, floorSet); err != nil {
					return nil, err
				}
			case NodeTowns:
				parseTowns(child, gm)
			case NodeWaypoints:
				parseWaypoints(child, gm)
			}
		}
	}

	for f := range floorSet {
		gm.Floors = append(gm.Floors, f)
	}

	return gm, nil
}

func parseTileArea(node *Node, gm *GameMap, floorSet map[uint8]bool) error {
	node.ResetPos()
	baseX, err := node.GetU16()
	if err != nil {
		return fmt.Errorf("tile area baseX: %w", err)
	}
	baseY, err := node.GetU16()
	if err != nil {
		return fmt.Errorf("tile area baseY: %w", err)
	}
	baseZ, err := node.GetU8()
	if err != nil {
		return fmt.Errorf("tile area baseZ: %w", err)
	}

	for _, tileNode := range node.Children {
		if tileNode.Type != NodeTile && tileNode.Type != NodeHouseTile {
			continue
		}
		tileNode.ResetPos()

		offsetX, err := tileNode.GetU8()
		if err != nil {
			continue
		}
		offsetY, err := tileNode.GetU8()
		if err != nil {
			continue
		}

		x := baseX + uint16(offsetX)
		y := baseY + uint16(offsetY)
		z := baseZ

		tile := &MapTile{}
		if tileNode.Type == NodeHouseTile {
			hid, err := tileNode.GetU32()
			if err != nil {
				continue
			}
			tile.HouseID = hid
		}

		for tileNode.Remaining() > 0 {
			attr, err := tileNode.GetU8()
			if err != nil {
				break
			}
			switch attr {
			case AttrTileFlags:
				flags, err := tileNode.GetU32()
				if err != nil {
					break
				}
				tile.Flags = flags
			case AttrItem:
				sid, err := tileNode.GetU16()
				if err != nil {
					break
				}
				tile.Items = append(tile.Items, sid)
				tile.RichItems = append(tile.RichItems, MapItem{ID: sid})
			default:
				if !skipTileAttr(tileNode, attr) {
					goto doneTileAttrs
				}
			}
		}
	doneTileAttrs:

		for _, itemNode := range tileNode.Children {
			if itemNode.Type != NodeItem {
				continue
			}
			item := parseItemNode(itemNode)
			tile.Items = append(tile.Items, item.ID)
			tile.RichItems = append(tile.RichItems, item)
		}

		if len(tile.Items) > 0 || tile.Flags != 0 || tile.HouseID != 0 {
			key := PackPos(x, y, z)
			gm.Tiles[key] = tile

			if x < gm.MinX {
				gm.MinX = x
			}
			if x > gm.MaxX {
				gm.MaxX = x
			}
			if y < gm.MinY {
				gm.MinY = y
			}
			if y > gm.MaxY {
				gm.MaxY = y
			}
			floorSet[z] = true

			fb := gm.FloorBounds[z]
			if fb == nil {
				fb = &FloorBounds{MinX: x, MinY: y, MaxX: x, MaxY: y}
				gm.FloorBounds[z] = fb
			} else {
				if x < fb.MinX {
					fb.MinX = x
				}
				if x > fb.MaxX {
					fb.MaxX = x
				}
				if y < fb.MinY {
					fb.MinY = y
				}
				if y > fb.MaxY {
					fb.MaxY = y
				}
			}
		}
	}
	return nil
}

func skipTileAttr(node *Node, attr uint8) bool {
	switch attr {
	case AttrDescription, AttrSpawnFile, AttrHouseFile, AttrExtFile:
		_, err := node.GetString()
		return err == nil
	default:
		return false
	}
}

func parseItemNode(node *Node) MapItem {
	node.ResetPos()
	if node.Remaining() < 2 {
		return MapItem{}
	}
	sid, _ := node.GetU16()
	item := MapItem{ID: sid}
	parseItemAttrs(node, &item)

	for _, child := range node.Children {
		if child.Type != NodeItem {
			continue
		}
		subItem := parseItemNode(child)
		item.SubItems = append(item.SubItems, subItem)
	}
	return item
}

const (
	attrMapTypeString  = 1
	attrMapTypeInteger = 2
	attrMapTypeFloat   = 3
	attrMapTypeBoolean = 4
)

func parseAttrMap(node *Node, item *MapItem) {
	count, err := node.GetU16()
	if err != nil {
		return
	}
	for i := 0; i < int(count); i++ {
		key, err := node.GetString()
		if err != nil {
			return
		}
		typeByte, err := node.GetU8()
		if err != nil {
			return
		}
		var val ItemAttrValue
		val.Type = typeByte
		switch typeByte {
		case attrMapTypeString:
			v, err := node.GetLongString()
			if err != nil {
				return
			}
			val.Str = v
		case attrMapTypeInteger:
			v, err := node.GetU32()
			if err != nil {
				return
			}
			val.Int = int32(v)
		case attrMapTypeFloat:
			v, err := node.GetFloat32()
			if err != nil {
				return
			}
			val.Float = v
		case attrMapTypeBoolean:
			v, err := node.GetU8()
			if err != nil {
				return
			}
			val.Boolean = v != 0
		default:
			return
		}
		if item.CustomAttrs == nil {
			item.CustomAttrs = make(map[string]ItemAttrValue)
		}
		item.CustomAttrs[key] = val
	}
}

func parseItemAttrs(node *Node, item *MapItem) {
	for node.Remaining() > 0 {
		attr, err := node.GetU8()
		if err != nil {
			return
		}
		switch attr {
		case AttrActionID:
			v, err := node.GetU16()
			if err != nil {
				return
			}
			item.ActionID = v
		case AttrUniqueID:
			v, err := node.GetU16()
			if err != nil {
				return
			}
			item.UniqueID = v
		case AttrText:
			v, err := node.GetString()
			if err != nil {
				return
			}
			item.Text = v
		case AttrDesc:
			v, err := node.GetString()
			if err != nil {
				return
			}
			item.Description = v
		case AttrTeleDest:
			x, err := node.GetU16()
			if err != nil {
				return
			}
			y, err := node.GetU16()
			if err != nil {
				return
			}
			z, err := node.GetU8()
			if err != nil {
				return
			}
			item.TeleDest = &TeleportDest{X: x, Y: y, Z: z}
		case AttrDepotID:
			v, err := node.GetU16()
			if err != nil {
				return
			}
			item.DepotID = v
		case AttrDoorID:
			v, err := node.GetU8()
			if err != nil {
				return
			}
			item.DoorID = v
		case AttrCharges:
			v, err := node.GetU16()
			if err != nil {
				return
			}
			item.Charges = v
		case AttrRuneCharges:
			v, err := node.GetU8()
			if err != nil {
				return
			}
			item.RuneCharges = v
		case AttrCount:
			v, err := node.GetU8()
			if err != nil {
				return
			}
			item.Count = v
		case AttrDuration:
			v, err := node.GetU32()
			if err != nil {
				return
			}
			item.Duration = v
		case AttrDecayState:
			if err := node.Skip(1); err != nil {
				return
			}
		case AttrWrittenDate:
			v, err := node.GetU32()
			if err != nil {
				return
			}
			item.WrittenDate = v
		case AttrWrittenBy:
			v, err := node.GetString()
			if err != nil {
				return
			}
			item.WrittenBy = v
		case AttrSleeperGUID:
			v, err := node.GetU32()
			if err != nil {
				return
			}
			item.SleeperGUID = v
		case AttrSleepStart:
			v, err := node.GetU32()
			if err != nil {
				return
			}
			item.SleepStart = v
		case AttrContainerItems:
			if err := node.Skip(4); err != nil {
				return
			}
		case AttrAttrMap:
			parseAttrMap(node, item)
			return
		default:
			return
		}
	}
}

func parseTowns(node *Node, gm *GameMap) {
	for _, townNode := range node.Children {
		if townNode.Type != NodeTown {
			continue
		}
		townNode.ResetPos()

		id, err := townNode.GetU32()
		if err != nil {
			continue
		}
		name, err := townNode.GetString()
		if err != nil {
			continue
		}
		x, err := townNode.GetU16()
		if err != nil {
			continue
		}
		y, err := townNode.GetU16()
		if err != nil {
			continue
		}
		z, err := townNode.GetU8()
		if err != nil {
			continue
		}

		gm.Towns = append(gm.Towns, Town{
			ID:   id,
			Name: name,
			X:    x,
			Y:    y,
			Z:    z,
		})
	}
}

func parseWaypoints(node *Node, gm *GameMap) {
	for _, wpNode := range node.Children {
		if wpNode.Type != NodeWaypoint {
			continue
		}
		wpNode.ResetPos()

		name, err := wpNode.GetString()
		if err != nil {
			continue
		}
		x, err := wpNode.GetU16()
		if err != nil {
			continue
		}
		y, err := wpNode.GetU16()
		if err != nil {
			continue
		}
		z, err := wpNode.GetU8()
		if err != nil {
			continue
		}

		gm.Waypoints = append(gm.Waypoints, Waypoint{
			Name: name,
			X:    x,
			Y:    y,
			Z:    z,
		})
	}
}
