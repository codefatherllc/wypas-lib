package gamedata

import (
	"database/sql"
	"strings"
)

func LoadTiles(db *sql.DB) ([]MapTile, error) {
	return loadTilesQuery(db, "SELECT x, y, z, ground_id, flags, house_id, items FROM map_tiles")
}

func LoadTilesInArea(db *sql.DB, minX, minY, maxX, maxY uint16, z uint8) ([]MapTile, error) {
	return loadTilesQuery(db,
		"SELECT x, y, z, ground_id, flags, house_id, items FROM map_tiles "+
			"WHERE x >= ? AND x <= ? AND y >= ? AND y <= ? AND z = ?",
		minX, maxX, minY, maxY, z)
}

func loadTilesQuery(db *sql.DB, query string, args ...interface{}) ([]MapTile, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiles []MapTile
	for rows.Next() {
		var t MapTile
		var blob []byte
		if err := rows.Scan(&t.X, &t.Y, &t.Z, &t.GroundID, &t.Flags, &t.HouseID, &blob); err != nil {
			return nil, err
		}
		t.Items, err = DecodeItems(blob)
		if err != nil {
			return nil, err
		}
		tiles = append(tiles, t)
	}
	return tiles, rows.Err()
}

func LoadSpawns(db *sql.DB) ([]Spawn, error) {
	rows, err := db.Query("SELECT id, center_x, center_y, center_z, radius FROM map_spawns")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	spawnMap := make(map[uint32]int)
	var spawns []Spawn
	for rows.Next() {
		var s Spawn
		if err := rows.Scan(&s.ID, &s.CenterX, &s.CenterY, &s.CenterZ, &s.Radius); err != nil {
			return nil, err
		}
		spawnMap[s.ID] = len(spawns)
		spawns = append(spawns, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	erows, err := db.Query("SELECT id, spawn_id, name, type, offset_x, offset_y, offset_z, spawntime FROM map_spawn_entries")
	if err != nil {
		return nil, err
	}
	defer erows.Close()

	for erows.Next() {
		var c SpawnCreature
		if err := erows.Scan(&c.ID, &c.SpawnID, &c.Name, &c.Type, &c.OffsetX, &c.OffsetY, &c.OffsetZ, &c.SpawnTime); err != nil {
			return nil, err
		}
		if idx, ok := spawnMap[c.SpawnID]; ok {
			spawns[idx].Creatures = append(spawns[idx].Creatures, c)
		}
	}
	return spawns, erows.Err()
}

func LoadTowns(db *sql.DB) ([]Town, error) {
	rows, err := db.Query("SELECT id, name, entry_x, entry_y, entry_z FROM map_towns")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var towns []Town
	for rows.Next() {
		var t Town
		if err := rows.Scan(&t.ID, &t.Name, &t.EntryX, &t.EntryY, &t.EntryZ); err != nil {
			return nil, err
		}
		towns = append(towns, t)
	}
	return towns, rows.Err()
}

func LoadWaypoints(db *sql.DB) ([]Waypoint, error) {
	rows, err := db.Query("SELECT name, x, y, z FROM map_waypoints")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wps []Waypoint
	for rows.Next() {
		var w Waypoint
		if err := rows.Scan(&w.Name, &w.X, &w.Y, &w.Z); err != nil {
			return nil, err
		}
		wps = append(wps, w)
	}
	return wps, rows.Err()
}

func LoadHouses(db *sql.DB) ([]House, error) {
	rows, err := db.Query("SELECT id, name, entry_x, entry_y, entry_z, rent, town_id, size FROM map_houses")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var houses []House
	for rows.Next() {
		var h House
		if err := rows.Scan(&h.ID, &h.Name, &h.EntryX, &h.EntryY, &h.EntryZ, &h.Rent, &h.TownID, &h.Size); err != nil {
			return nil, err
		}
		houses = append(houses, h)
	}
	return houses, rows.Err()
}

func SaveTile(db *sql.DB, tile *MapTile) error {
	blob := EncodeItems(tile.Items)
	_, err := db.Exec(
		"INSERT INTO map_tiles (x, y, z, ground_id, flags, house_id, items) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?) "+
			"ON DUPLICATE KEY UPDATE ground_id=VALUES(ground_id), flags=VALUES(flags), "+
			"house_id=VALUES(house_id), items=VALUES(items)",
		tile.X, tile.Y, tile.Z, tile.GroundID, tile.Flags, tile.HouseID, blob,
	)
	return err
}

func BulkInsertTiles(tx *sql.Tx, tiles []MapTile) error {
	if len(tiles) == 0 {
		return nil
	}

	const batchSize = 500
	for i := 0; i < len(tiles); i += batchSize {
		end := i + batchSize
		if end > len(tiles) {
			end = len(tiles)
		}
		batch := tiles[i:end]

		rows := make([]string, len(batch))
		args := make([]interface{}, 0, len(batch)*7)
		for j := range batch {
			rows[j] = "(?,?,?,?,?,?,?)"
			args = append(args,
				batch[j].X, batch[j].Y, batch[j].Z,
				batch[j].GroundID, batch[j].Flags, batch[j].HouseID,
				EncodeItems(batch[j].Items),
			)
		}

		q := "INSERT INTO map_tiles (x, y, z, ground_id, flags, house_id, items) VALUES " +
			strings.Join(rows, ",") +
			" ON DUPLICATE KEY UPDATE ground_id=VALUES(ground_id), flags=VALUES(flags), " +
			"house_id=VALUES(house_id), items=VALUES(items)"

		if _, err := tx.Exec(q, args...); err != nil {
			return err
		}
	}
	return nil
}
