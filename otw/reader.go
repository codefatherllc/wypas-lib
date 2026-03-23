package otw

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	fbOtw "github.com/codefatherllc/wypas-proto/otw"
)

var magicOTW = [4]byte{'O', 'T', 'W', 0}

const flagGzip = 1

func ReadFile(path string) (*WorldMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("otw: read file: %w", err)
	}
	if len(data) < 8 {
		return nil, fmt.Errorf("otw: file too short")
	}

	var magic [4]byte
	copy(magic[:], data[:4])
	if magic != magicOTW {
		return nil, fmt.Errorf("otw: invalid magic %q", magic)
	}

	flags := binary.LittleEndian.Uint16(data[6:8])
	payload := data[8:]

	if flags&flagGzip != 0 {
		payload, err = decompressGzip(payload)
		if err != nil {
			return nil, fmt.Errorf("otw: decompress: %w", err)
		}
	}

	fb := fbOtw.GetRootAsWorldMap(payload, 0)
	return convertWorldMap(fb), nil
}

func decompressGzip(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func convertWorldMap(fb *fbOtw.WorldMap) *WorldMap {
	wm := &WorldMap{
		Version:     fb.Version(),
		Width:       fb.Width(),
		Height:      fb.Height(),
		Description: string(fb.Description()),
	}

	if n := fb.TileAreasLength(); n > 0 {
		wm.TileAreas = make([]TileArea, n)
		var fbTA fbOtw.TileArea
		for i := 0; i < n; i++ {
			if fb.TileAreas(&fbTA, i) {
				wm.TileAreas[i] = convertTileArea(&fbTA)
			}
		}
	}

	if n := fb.TownsLength(); n > 0 {
		wm.Towns = make([]Town, n)
		var fbT fbOtw.Town
		for i := 0; i < n; i++ {
			if fb.Towns(&fbT, i) {
				wm.Towns[i] = Town{
					ID:      fbT.Id(),
					Name:    string(fbT.Name()),
					TempleX: fbT.TempleX(),
					TempleY: fbT.TempleY(),
					TempleZ: fbT.TempleZ(),
				}
			}
		}
	}

	if n := fb.WaypointsLength(); n > 0 {
		wm.Waypoints = make([]Waypoint, n)
		var fbW fbOtw.Waypoint
		for i := 0; i < n; i++ {
			if fb.Waypoints(&fbW, i) {
				wm.Waypoints[i] = Waypoint{
					Name: string(fbW.Name()),
					X:    fbW.X(),
					Y:    fbW.Y(),
					Z:    fbW.Z(),
				}
			}
		}
	}

	if n := fb.SpawnsLength(); n > 0 {
		wm.Spawns = make([]Spawn, n)
		var fbS fbOtw.Spawn
		for i := 0; i < n; i++ {
			if fb.Spawns(&fbS, i) {
				wm.Spawns[i] = convertSpawn(&fbS)
			}
		}
	}

	if n := fb.HousesLength(); n > 0 {
		wm.Houses = make([]House, n)
		var fbH fbOtw.House
		for i := 0; i < n; i++ {
			if fb.Houses(&fbH, i) {
				wm.Houses[i] = House{
					ID:     fbH.Id(),
					Name:   string(fbH.Name()),
					EntryX: fbH.EntryX(),
					EntryY: fbH.EntryY(),
					EntryZ: fbH.EntryZ(),
					Size:   fbH.Size(),
					Rent:   fbH.Rent(),
					TownID: fbH.TownId(),
				}
			}
		}
	}

	return wm
}

func convertTileArea(fb *fbOtw.TileArea) TileArea {
	ta := TileArea{
		BaseX: fb.BaseX(),
		BaseY: fb.BaseY(),
		BaseZ: fb.BaseZ(),
	}
	if n := fb.TilesLength(); n > 0 {
		ta.Tiles = make([]Tile, n)
		var fbT fbOtw.Tile
		for i := 0; i < n; i++ {
			if fb.Tiles(&fbT, i) {
				ta.Tiles[i] = convertTile(&fbT)
			}
		}
	}
	return ta
}

func convertTile(fb *fbOtw.Tile) Tile {
	t := Tile{
		OffsetX: fb.OffsetX(),
		OffsetY: fb.OffsetY(),
		Flags:   fb.Flags(),
		HouseID: fb.HouseId(),
	}
	if n := fb.ItemsLength(); n > 0 {
		t.Items = make([]MapItem, n)
		var fbI fbOtw.MapItem
		for i := 0; i < n; i++ {
			if fb.Items(&fbI, i) {
				t.Items[i] = convertMapItem(&fbI)
			}
		}
	}
	return t
}

func convertMapItem(fb *fbOtw.MapItem) MapItem {
	mi := MapItem{
		ServerID:    fb.ServerId(),
		Count:       fb.Count(),
		ActionID:    fb.ActionId(),
		UniqueID:    fb.UniqueId(),
		TeleDestX:   fb.TeleDestX(),
		TeleDestY:   fb.TeleDestY(),
		TeleDestZ:   fb.TeleDestZ(),
		DoorID:      fb.DoorId(),
		DepotID:     fb.DepotId(),
		Text:        string(fb.Text()),
		Description: string(fb.Description()),
		Charges:     fb.Charges(),
		RuneCharges: fb.RuneCharges(),
		Duration:    fb.Duration(),
		DecayState:  fb.DecayingState(),
		WrittenDate: fb.WrittenDate(),
		WrittenBy:   string(fb.WrittenBy()),
		SleeperGUID: fb.SleeperGuid(),
		SleepStart:  fb.SleepStart(),
	}
	if n := fb.SubItemsLength(); n > 0 {
		mi.SubItems = make([]MapItem, n)
		var fbSub fbOtw.MapItem
		for i := 0; i < n; i++ {
			if fb.SubItems(&fbSub, i) {
				mi.SubItems[i] = convertMapItem(&fbSub)
			}
		}
	}
	return mi
}

func convertSpawn(fb *fbOtw.Spawn) Spawn {
	s := Spawn{
		CenterX: fb.CenterX(),
		CenterY: fb.CenterY(),
		CenterZ: fb.CenterZ(),
		Radius:  fb.Radius(),
	}
	if n := fb.CreaturesLength(); n > 0 {
		s.Creatures = make([]SpawnCreature, n)
		var fbC fbOtw.SpawnCreature
		for i := 0; i < n; i++ {
			if fb.Creatures(&fbC, i) {
				s.Creatures[i] = SpawnCreature{
					Name:      string(fbC.Name()),
					OffsetX:   fbC.OffsetX(),
					OffsetY:   fbC.OffsetY(),
					SpawnTime: fbC.SpawnTime(),
					Direction: fbC.Direction(),
				}
			}
		}
	}
	return s
}
