package otbm

import (
	"fmt"
	"math"
)

const (
	NodeMapData  = 2
	NodeTileArea = 4
	NodeTile     = 5
	NodeItem     = 6
	NodeTowns    = 12
	NodeTown     = 13
	NodeHouseTile = 14
)

const (
	AttrDescription = 1
	AttrTileFlags   = 3
	AttrItem        = 9
	AttrSpawnFile   = 11
	AttrHouseFile   = 13
)

type MapTile struct {
	Items []uint16
	Flags uint32
}

type Town struct {
	ID   uint32
	Name string
	X, Y uint16
	Z    uint8
}

type FloorBounds struct {
	MinX, MinY, MaxX, MaxY uint16
}

type GameMap struct {
	Tiles              map[uint64]*MapTile
	MinX, MinY         uint16
	MaxX, MaxY         uint16
	Floors             []uint8
	Towns              []Town
	FloorBounds        map[uint8]*FloorBounds
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
	if err := root.Skip(4 + 2 + 2 + 1 + 3 + 4); err != nil {
		return nil, fmt.Errorf("skip root header: %w", err)
	}

	gm := &GameMap{
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

		for _, child := range mapDataNode.Children {
			switch child.Type {
			case NodeTileArea:
				if err := parseTileArea(child, gm, floorSet); err != nil {
					return nil, err
				}
			case NodeTowns:
				parseTowns(child, gm)
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

		if tileNode.Type == NodeHouseTile {
			tileNode.Skip(4)
		}

		tile := &MapTile{}

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
			default:
				if attr == AttrDescription {
					s, err := tileNode.GetString()
					_ = s
					if err != nil {
						goto doneTileAttrs
					}
				} else if attr == AttrSpawnFile || attr == AttrHouseFile {
					s, err := tileNode.GetString()
					_ = s
					if err != nil {
						goto doneTileAttrs
					}
				} else {
					goto doneTileAttrs
				}
			}
		}
	doneTileAttrs:

		for _, itemNode := range tileNode.Children {
			if itemNode.Type != NodeItem {
				continue
			}
			itemNode.ResetPos()
			if itemNode.Remaining() >= 2 {
				sid, _ := itemNode.GetU16()
				tile.Items = append(tile.Items, sid)
			}
		}

		if len(tile.Items) > 0 || tile.Flags != 0 {
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
