package otw

import (
	"bytes"
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

	compressed, err := compressGzip(payload)
	if err != nil {
		return fmt.Errorf("otw: compress: %w", err)
	}

	var header [8]byte
	copy(header[:4], magicOTW[:])
	binary.LittleEndian.PutUint16(header[4:6], 1)
	binary.LittleEndian.PutUint16(header[6:8], flagGzip)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("otw: create file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(header[:]); err != nil {
		return fmt.Errorf("otw: write header: %w", err)
	}
	if _, err := f.Write(compressed); err != nil {
		return fmt.Errorf("otw: write payload: %w", err)
	}
	return nil
}

func compressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
	itemOffsets := make([]flatbuffers.UOffsetT, len(t.Items))
	for i := range t.Items {
		itemOffsets[i] = buildMapItem(b, &t.Items[i])
	}

	itemsVec := prependVector(b, itemOffsets, fbOtw.TileStartItemsVector)

	fbOtw.TileStart(b)
	fbOtw.TileAddOffsetX(b, t.OffsetX)
	fbOtw.TileAddOffsetY(b, t.OffsetY)
	fbOtw.TileAddFlags(b, t.Flags)
	fbOtw.TileAddHouseId(b, t.HouseID)
	fbOtw.TileAddItems(b, itemsVec)
	return fbOtw.TileEnd(b)
}

func buildMapItem(b *flatbuffers.Builder, mi *MapItem) flatbuffers.UOffsetT {
	textOff := b.CreateString(mi.Text)
	descOff := b.CreateString(mi.Description)
	writtenByOff := b.CreateString(mi.WrittenBy)

	subOffsets := make([]flatbuffers.UOffsetT, len(mi.SubItems))
	for i := range mi.SubItems {
		subOffsets[i] = buildMapItem(b, &mi.SubItems[i])
	}
	subVec := prependVector(b, subOffsets, fbOtw.MapItemStartSubItemsVector)

	fbOtw.MapItemStart(b)
	fbOtw.MapItemAddServerId(b, mi.ServerID)
	fbOtw.MapItemAddCount(b, mi.Count)
	fbOtw.MapItemAddActionId(b, mi.ActionID)
	fbOtw.MapItemAddUniqueId(b, mi.UniqueID)
	fbOtw.MapItemAddTeleDestX(b, mi.TeleDestX)
	fbOtw.MapItemAddTeleDestY(b, mi.TeleDestY)
	fbOtw.MapItemAddTeleDestZ(b, mi.TeleDestZ)
	fbOtw.MapItemAddDoorId(b, mi.DoorID)
	fbOtw.MapItemAddDepotId(b, mi.DepotID)
	fbOtw.MapItemAddText(b, textOff)
	fbOtw.MapItemAddDescription(b, descOff)
	fbOtw.MapItemAddCharges(b, mi.Charges)
	fbOtw.MapItemAddRuneCharges(b, mi.RuneCharges)
	fbOtw.MapItemAddDuration(b, mi.Duration)
	fbOtw.MapItemAddDecayingState(b, mi.DecayState)
	fbOtw.MapItemAddWrittenDate(b, mi.WrittenDate)
	fbOtw.MapItemAddWrittenBy(b, writtenByOff)
	fbOtw.MapItemAddSleeperGuid(b, mi.SleeperGUID)
	fbOtw.MapItemAddSleepStart(b, mi.SleepStart)
	fbOtw.MapItemAddSubItems(b, subVec)
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
