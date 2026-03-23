package otw

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"os"

	flatbuffers "github.com/google/flatbuffers/go"

	fbOtw "github.com/codefatherllc/wypas-proto/otw"
)

func WriteFile(path string, wm *WorldMap) error {
	builder := flatbuffers.NewBuilder(4 * 1024 * 1024)

	descOff := builder.CreateString(wm.Description)

	tileAreaOffsets := buildTileAreas(builder, wm.TileAreas)
	townOffsets := buildTowns(builder, wm.Towns)
	waypointOffsets := buildWaypoints(builder, wm.Waypoints)
	spawnOffsets := buildSpawns(builder, wm.Spawns)
	houseOffsets := buildHouses(builder, wm.Houses)

	tileAreasVec := prependVector(builder, tileAreaOffsets, fbOtw.WorldMapStartTileAreasVector)
	townsVec := prependVector(builder, townOffsets, fbOtw.WorldMapStartTownsVector)
	waypointsVec := prependVector(builder, waypointOffsets, fbOtw.WorldMapStartWaypointsVector)
	spawnsVec := prependVector(builder, spawnOffsets, fbOtw.WorldMapStartSpawnsVector)
	housesVec := prependVector(builder, houseOffsets, fbOtw.WorldMapStartHousesVector)

	fbOtw.WorldMapStart(builder)
	fbOtw.WorldMapAddVersion(builder, wm.Version)
	fbOtw.WorldMapAddWidth(builder, wm.Width)
	fbOtw.WorldMapAddHeight(builder, wm.Height)
	fbOtw.WorldMapAddDescription(builder, descOff)
	fbOtw.WorldMapAddTileAreas(builder, tileAreasVec)
	fbOtw.WorldMapAddTowns(builder, townsVec)
	fbOtw.WorldMapAddWaypoints(builder, waypointsVec)
	fbOtw.WorldMapAddSpawns(builder, spawnsVec)
	fbOtw.WorldMapAddHouses(builder, housesVec)
	root := fbOtw.WorldMapEnd(builder)
	builder.Finish(root)

	payload := builder.FinishedBytes()

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("otw: create file: %w", err)
	}
	defer f.Close()

	var header [8]byte
	copy(header[:4], magicOTW[:])
	binary.LittleEndian.PutUint16(header[4:6], 1)
	binary.LittleEndian.PutUint16(header[6:8], flagGzip)
	if _, err := f.Write(header[:]); err != nil {
		return fmt.Errorf("otw: write header: %w", err)
	}

	gz, _ := gzip.NewWriterLevel(f, gzip.BestCompression)
	if _, err := gz.Write(payload); err != nil {
		return fmt.Errorf("otw: write compressed: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("otw: close gzip: %w", err)
	}

	return nil
}

func prependVector(b *flatbuffers.Builder, offsets []flatbuffers.UOffsetT, startVec func(*flatbuffers.Builder, int) flatbuffers.UOffsetT) flatbuffers.UOffsetT {
	if len(offsets) == 0 {
		return 0
	}
	startVec(b, len(offsets))
	for i := len(offsets) - 1; i >= 0; i-- {
		b.PrependUOffsetT(offsets[i])
	}
	return b.EndVector(len(offsets))
}

func buildTileAreas(b *flatbuffers.Builder, areas []TileArea) []flatbuffers.UOffsetT {
	offsets := make([]flatbuffers.UOffsetT, len(areas))
	for i := range areas {
		offsets[i] = buildTileArea(b, &areas[i])
	}
	return offsets
}

func buildTileArea(b *flatbuffers.Builder, ta *TileArea) flatbuffers.UOffsetT {
	tileOffsets := make([]flatbuffers.UOffsetT, len(ta.Tiles))
	for i := range ta.Tiles {
		tileOffsets[i] = buildTile(b, &ta.Tiles[i])
	}

	tilesVec := prependVector(b, tileOffsets, fbOtw.TileAreaStartTilesVector)

	fbOtw.TileAreaStart(b)
	fbOtw.TileAreaAddBaseX(b, ta.BaseX)
	fbOtw.TileAreaAddBaseY(b, ta.BaseY)
	fbOtw.TileAreaAddBaseZ(b, ta.BaseZ)
	fbOtw.TileAreaAddTiles(b, tilesVec)
	return fbOtw.TileAreaEnd(b)
}

func buildTile(b *flatbuffers.Builder, t *Tile) flatbuffers.UOffsetT {
	var itemsVec, richVec flatbuffers.UOffsetT

	if len(t.Items) > 0 {
		fbOtw.TileStartItemsVector(b, len(t.Items))
		for i := len(t.Items) - 1; i >= 0; i-- {
			b.PrependUint16(t.Items[i])
		}
		itemsVec = b.EndVector(len(t.Items))
	}

	if len(t.RichItems) > 0 {
		richOffsets := make([]flatbuffers.UOffsetT, len(t.RichItems))
		for i := range t.RichItems {
			richOffsets[i] = buildMapItem(b, &t.RichItems[i])
		}
		richVec = prependVector(b, richOffsets, fbOtw.TileStartRichItemsVector)
	}

	fbOtw.TileStart(b)
	fbOtw.TileAddOffsetX(b, t.OffsetX)
	fbOtw.TileAddOffsetY(b, t.OffsetY)
	if t.Flags != 0 {
		fbOtw.TileAddFlags(b, t.Flags)
	}
	if t.HouseID != 0 {
		fbOtw.TileAddHouseId(b, t.HouseID)
	}
	if itemsVec != 0 {
		fbOtw.TileAddItems(b, itemsVec)
	}
	if richVec != 0 {
		fbOtw.TileAddRichItems(b, richVec)
	}
	return fbOtw.TileEnd(b)
}

func buildMapItem(b *flatbuffers.Builder, mi *MapItem) flatbuffers.UOffsetT {
	var textOff, descOff, writtenByOff flatbuffers.UOffsetT
	if mi.Text != "" {
		textOff = b.CreateString(mi.Text)
	}
	if mi.Description != "" {
		descOff = b.CreateString(mi.Description)
	}
	if mi.WrittenBy != "" {
		writtenByOff = b.CreateString(mi.WrittenBy)
	}

	var subVec flatbuffers.UOffsetT
	if len(mi.SubItems) > 0 {
		subOffsets := make([]flatbuffers.UOffsetT, len(mi.SubItems))
		for i := range mi.SubItems {
			subOffsets[i] = buildMapItem(b, &mi.SubItems[i])
		}
		subVec = prependVector(b, subOffsets, fbOtw.MapItemStartSubItemsVector)
	}

	fbOtw.MapItemStart(b)
	fbOtw.MapItemAddServerId(b, mi.ServerID)
	if mi.Count != 0 {
		fbOtw.MapItemAddCount(b, mi.Count)
	}
	if mi.ActionID != 0 {
		fbOtw.MapItemAddActionId(b, mi.ActionID)
	}
	if mi.UniqueID != 0 {
		fbOtw.MapItemAddUniqueId(b, mi.UniqueID)
	}
	if mi.TeleDestX != 0 || mi.TeleDestY != 0 || mi.TeleDestZ != 0 {
		fbOtw.MapItemAddTeleDestX(b, mi.TeleDestX)
		fbOtw.MapItemAddTeleDestY(b, mi.TeleDestY)
		fbOtw.MapItemAddTeleDestZ(b, mi.TeleDestZ)
	}
	if mi.DoorID != 0 {
		fbOtw.MapItemAddDoorId(b, mi.DoorID)
	}
	if mi.DepotID != 0 {
		fbOtw.MapItemAddDepotId(b, mi.DepotID)
	}
	if textOff != 0 {
		fbOtw.MapItemAddText(b, textOff)
	}
	if descOff != 0 {
		fbOtw.MapItemAddDescription(b, descOff)
	}
	if mi.Charges != 0 {
		fbOtw.MapItemAddCharges(b, mi.Charges)
	}
	if mi.RuneCharges != 0 {
		fbOtw.MapItemAddRuneCharges(b, mi.RuneCharges)
	}
	if mi.Duration != 0 {
		fbOtw.MapItemAddDuration(b, mi.Duration)
	}
	if mi.DecayState != 0 {
		fbOtw.MapItemAddDecayingState(b, mi.DecayState)
	}
	if mi.WrittenDate != 0 {
		fbOtw.MapItemAddWrittenDate(b, mi.WrittenDate)
	}
	if writtenByOff != 0 {
		fbOtw.MapItemAddWrittenBy(b, writtenByOff)
	}
	if subVec != 0 {
		fbOtw.MapItemAddSubItems(b, subVec)
	}
	return fbOtw.MapItemEnd(b)
}

func buildTowns(b *flatbuffers.Builder, towns []Town) []flatbuffers.UOffsetT {
	offsets := make([]flatbuffers.UOffsetT, len(towns))
	for i := range towns {
		nameOff := b.CreateString(towns[i].Name)
		fbOtw.TownStart(b)
		fbOtw.TownAddId(b, towns[i].ID)
		fbOtw.TownAddName(b, nameOff)
		fbOtw.TownAddTempleX(b, towns[i].TempleX)
		fbOtw.TownAddTempleY(b, towns[i].TempleY)
		fbOtw.TownAddTempleZ(b, towns[i].TempleZ)
		offsets[i] = fbOtw.TownEnd(b)
	}
	return offsets
}

func buildWaypoints(b *flatbuffers.Builder, waypoints []Waypoint) []flatbuffers.UOffsetT {
	offsets := make([]flatbuffers.UOffsetT, len(waypoints))
	for i := range waypoints {
		nameOff := b.CreateString(waypoints[i].Name)
		fbOtw.WaypointStart(b)
		fbOtw.WaypointAddName(b, nameOff)
		fbOtw.WaypointAddX(b, waypoints[i].X)
		fbOtw.WaypointAddY(b, waypoints[i].Y)
		fbOtw.WaypointAddZ(b, waypoints[i].Z)
		offsets[i] = fbOtw.WaypointEnd(b)
	}
	return offsets
}

func buildSpawns(b *flatbuffers.Builder, spawns []Spawn) []flatbuffers.UOffsetT {
	offsets := make([]flatbuffers.UOffsetT, len(spawns))
	for i := range spawns {
		creatureOffsets := buildSpawnCreatures(b, spawns[i].Creatures)
		creaturesVec := prependVector(b, creatureOffsets, fbOtw.SpawnStartCreaturesVector)

		fbOtw.SpawnStart(b)
		fbOtw.SpawnAddCenterX(b, spawns[i].CenterX)
		fbOtw.SpawnAddCenterY(b, spawns[i].CenterY)
		fbOtw.SpawnAddCenterZ(b, spawns[i].CenterZ)
		fbOtw.SpawnAddRadius(b, spawns[i].Radius)
		fbOtw.SpawnAddCreatures(b, creaturesVec)
		offsets[i] = fbOtw.SpawnEnd(b)
	}
	return offsets
}

func buildSpawnCreatures(b *flatbuffers.Builder, creatures []SpawnCreature) []flatbuffers.UOffsetT {
	offsets := make([]flatbuffers.UOffsetT, len(creatures))
	for i := range creatures {
		nameOff := b.CreateString(creatures[i].Name)
		fbOtw.SpawnCreatureStart(b)
		fbOtw.SpawnCreatureAddName(b, nameOff)
		fbOtw.SpawnCreatureAddOffsetX(b, creatures[i].OffsetX)
		fbOtw.SpawnCreatureAddOffsetY(b, creatures[i].OffsetY)
		fbOtw.SpawnCreatureAddSpawnTime(b, creatures[i].SpawnTime)
		fbOtw.SpawnCreatureAddDirection(b, creatures[i].Direction)
		offsets[i] = fbOtw.SpawnCreatureEnd(b)
	}
	return offsets
}

func buildHouses(b *flatbuffers.Builder, houses []House) []flatbuffers.UOffsetT {
	offsets := make([]flatbuffers.UOffsetT, len(houses))
	for i := range houses {
		nameOff := b.CreateString(houses[i].Name)
		fbOtw.HouseStart(b)
		fbOtw.HouseAddId(b, houses[i].ID)
		fbOtw.HouseAddName(b, nameOff)
		fbOtw.HouseAddEntryX(b, houses[i].EntryX)
		fbOtw.HouseAddEntryY(b, houses[i].EntryY)
		fbOtw.HouseAddEntryZ(b, houses[i].EntryZ)
		fbOtw.HouseAddSize(b, houses[i].Size)
		fbOtw.HouseAddRent(b, houses[i].Rent)
		fbOtw.HouseAddTownId(b, houses[i].TownID)
		offsets[i] = fbOtw.HouseEnd(b)
	}
	return offsets
}
