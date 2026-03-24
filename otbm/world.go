package otbm

import (
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/codefatherllc/wypas-lib/gamedata"
)

type WorldData struct {
	Tiles     []gamedata.MapTile
	Spawns    []gamedata.Spawn
	Houses    []gamedata.House
	Towns     []gamedata.Town
	Waypoints []gamedata.Waypoint
}

type WorldOption func(*worldConfig)

type worldConfig struct {
	spawnsPath string
	housesPath string
}

func WithSpawns(path string) WorldOption {
	return func(c *worldConfig) {
		c.spawnsPath = path
	}
}

func WithHouses(path string) WorldOption {
	return func(c *worldConfig) {
		c.housesPath = path
	}
}

func LoadWorld(otbmPath string, opts ...WorldOption) (*WorldData, error) {
	var cfg worldConfig
	for _, o := range opts {
		o(&cfg)
	}

	gm, err := ParseOTBM(otbmPath)
	if err != nil {
		return nil, fmt.Errorf("otbm: parse: %w", err)
	}

	wd := &WorldData{}

	wd.Tiles = convertTiles(gm)
	wd.Towns = convertTowns(gm.Towns)
	wd.Waypoints = convertWaypoints(gm.Waypoints)

	if cfg.spawnsPath != "" {
		spawns, err := loadSpawns(cfg.spawnsPath)
		if err != nil {
			return nil, err
		}
		wd.Spawns = spawns
	}

	if cfg.housesPath != "" {
		houses, err := loadHouses(cfg.housesPath)
		if err != nil {
			return nil, err
		}
		wd.Houses = houses
	}

	return wd, nil
}

func convertTiles(gm *GameMap) []gamedata.MapTile {
	tiles := make([]gamedata.MapTile, 0, len(gm.Tiles))
	for key, tile := range gm.Tiles {
		x := uint16(key >> 24)
		y := uint16((key >> 8) & 0xFFFF)
		z := uint8(key & 0xFF)

		var groundID uint16
		items := tile.Items
		if len(items) > 0 {
			groundID = items[0]
			items = items[1:]
		}

		var itemsCopy []uint16
		if len(items) > 0 {
			itemsCopy = make([]uint16, len(items))
			copy(itemsCopy, items)
		}

		tiles = append(tiles, gamedata.MapTile{
			X:        x,
			Y:        y,
			Z:        z,
			GroundID: groundID,
			Flags:    tile.Flags,
			HouseID:  tile.HouseID,
			Items:    itemsCopy,
		})
	}
	return tiles
}

func convertTowns(towns []Town) []gamedata.Town {
	out := make([]gamedata.Town, len(towns))
	for i, t := range towns {
		out[i] = gamedata.Town{
			ID:     t.ID,
			Name:   t.Name,
			EntryX: t.X,
			EntryY: t.Y,
			EntryZ: t.Z,
		}
	}
	return out
}

func convertWaypoints(wps []Waypoint) []gamedata.Waypoint {
	out := make([]gamedata.Waypoint, len(wps))
	for i, w := range wps {
		out[i] = gamedata.Waypoint{
			Name: w.Name,
			X:    w.X,
			Y:    w.Y,
			Z:    w.Z,
		}
	}
	return out
}

func loadSpawns(path string) ([]gamedata.Spawn, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("otbm: read spawns %s: %w", path, err)
	}
	var xs xmlSpawns
	if err := xml.Unmarshal(data, &xs); err != nil {
		return nil, fmt.Errorf("otbm: parse spawns: %w", err)
	}

	spawns := make([]gamedata.Spawn, 0, len(xs.Spawns))
	for _, s := range xs.Spawns {
		gs := gamedata.Spawn{
			CenterX: s.CenterX,
			CenterY: s.CenterY,
			CenterZ: s.CenterZ,
			Radius:  s.Radius,
		}
		for _, m := range s.Monsters {
			gs.Creatures = append(gs.Creatures, gamedata.SpawnCreature{
				Name: m.Name, Type: "monster",
				OffsetX: m.X, OffsetY: m.Y, OffsetZ: m.Z,
				SpawnTime: m.SpawnTime,
			})
		}
		for _, n := range s.NPCs {
			gs.Creatures = append(gs.Creatures, gamedata.SpawnCreature{
				Name: n.Name, Type: "npc",
				OffsetX: n.X, OffsetY: n.Y, OffsetZ: n.Z,
				SpawnTime: n.SpawnTime,
			})
		}
		spawns = append(spawns, gs)
	}
	return spawns, nil
}

func loadHouses(path string) ([]gamedata.House, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("otbm: read houses %s: %w", path, err)
	}
	var xh xmlHouses
	if err := xml.Unmarshal(data, &xh); err != nil {
		return nil, fmt.Errorf("otbm: parse houses: %w", err)
	}

	houses := make([]gamedata.House, len(xh.Houses))
	for i, h := range xh.Houses {
		houses[i] = gamedata.House{
			ID: h.ID, Name: h.Name,
			EntryX: h.EntryX, EntryY: h.EntryY, EntryZ: h.EntryZ,
			Rent: h.Rent, TownID: h.TownID, Size: h.Size,
		}
	}
	return houses, nil
}

func EncodeItemsBlob(items []uint16) []byte {
	if len(items) == 0 {
		return nil
	}
	buf := make([]byte, len(items)*2)
	for i, id := range items {
		binary.LittleEndian.PutUint16(buf[i*2:], id)
	}
	return buf
}
