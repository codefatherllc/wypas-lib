package gamedata

import (
	"database/sql"
	"strings"
)

var itemColumns = []string{
	"server_id", "client_id", "item_group", "item_type", "flags",
	"speed", "top_order", "floor_change", "light_level", "light_color",
	"container_size", "fluid_source", "decay_to", "decay_time",
	"charges", "weight", "cacheable", "must_serialize", "attributes",
}

func scanItemType(row interface{ Scan(...interface{}) error }) (*ItemType, error) {
	var it ItemType
	return &it, row.Scan(
		&it.ServerID, &it.ClientID, &it.ItemGroup, &it.ItemTypeVal, &it.Flags,
		&it.Speed, &it.TopOrder, &it.FloorChange, &it.LightLevel, &it.LightColor,
		&it.ContainerSize, &it.FluidSource, &it.DecayTo, &it.DecayTime,
		&it.Charges, &it.Weight, &it.Cacheable, &it.MustSerialize, &it.Attributes,
	)
}

func itemValues(it *ItemType) []interface{} {
	return []interface{}{
		it.ServerID, it.ClientID, it.ItemGroup, it.ItemTypeVal, it.Flags,
		it.Speed, it.TopOrder, it.FloorChange, it.LightLevel, it.LightColor,
		it.ContainerSize, it.FluidSource, it.DecayTo, it.DecayTime,
		it.Charges, it.Weight, it.Cacheable, it.MustSerialize, it.Attributes,
	}
}

func LoadItemTypes(db *sql.DB) ([]ItemType, error) {
	rows, err := db.Query("SELECT " + strings.Join(itemColumns, ",") + " FROM item_types")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ItemType
	for rows.Next() {
		it, err := scanItemType(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *it)
	}
	return items, rows.Err()
}

func LoadItemType(db *sql.DB, serverID uint16) (*ItemType, error) {
	row := db.QueryRow("SELECT "+strings.Join(itemColumns, ",")+
		" FROM item_types WHERE server_id = ?", serverID)
	return scanItemType(row)
}

func SaveItemType(db *sql.DB, it *ItemType) error {
	cols := strings.Join(itemColumns, ",")
	placeholders := strings.Repeat("?,", len(itemColumns))
	placeholders = placeholders[:len(placeholders)-1]

	var updates []string
	for _, c := range itemColumns {
		if c == "server_id" {
			continue
		}
		updates = append(updates, c+" = VALUES("+c+")")
	}

	q := "INSERT INTO item_types (" + cols + ") VALUES (" + placeholders +
		") ON DUPLICATE KEY UPDATE " + strings.Join(updates, ",")

	_, err := db.Exec(q, itemValues(it)...)
	return err
}

func BulkInsertItems(tx *sql.Tx, items []ItemType) error {
	if len(items) == 0 {
		return nil
	}

	cols := strings.Join(itemColumns, ",")
	singleRow := "(" + strings.Repeat("?,", len(itemColumns))
	singleRow = singleRow[:len(singleRow)-1] + ")"

	var updates []string
	for _, c := range itemColumns {
		if c == "server_id" {
			continue
		}
		updates = append(updates, c+" = VALUES("+c+")")
	}
	updateClause := strings.Join(updates, ",")

	const batchSize = 500
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]

		rows := make([]string, len(batch))
		var args []interface{}
		for j := range batch {
			rows[j] = singleRow
			args = append(args, itemValues(&batch[j])...)
		}

		q := "INSERT INTO item_types (" + cols + ") VALUES " +
			strings.Join(rows, ",") +
			" ON DUPLICATE KEY UPDATE " + updateClause

		if _, err := tx.Exec(q, args...); err != nil {
			return err
		}
	}
	return nil
}
