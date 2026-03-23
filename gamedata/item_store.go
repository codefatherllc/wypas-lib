package gamedata

import (
	"database/sql"
	"strings"
)

var itemColumns = []string{
	"server_id", "client_id", "item_group", "item_type", "flags", "speed", "top_order",
	"name", "article", "plural", "description", "rune_spell_name",
	"weight", "armor", "defense", "extra_defense", "attack", "extra_attack", "attack_speed",
	"rotate_to", "container_size", "max_text_length", "write_once_item_id",
	"charges", "decay_to", "decay_time",
	"transform_equip_to", "transform_deequip_to", "transform_use_to",
	"duration", "show_duration", "show_charges", "show_count", "show_attributes",
	"break_chance", "hit_chance", "max_hit_chance", "dual_wield",
	"shoot_range", "worth", "level_door", "special_door", "closing_door",
	"ware_id", "force_serialize",
	"weapon_type", "ammo_type", "ammo_action", "shoot_type", "magic_effect",
	"slot_position", "wield_position", "fluid_source", "corpse_type",
	"light_level", "light_color", "minimap_color",
	"block_solid", "block_projectile", "block_path_find", "allow_dist_read",
	"movable", "pickupable", "allow_pickupable",
	"is_vertical", "is_horizontal", "walk_stack", "replaceable",
	"can_write_text", "can_read_text", "stop_time",
	"floorchange", "bed_partner_dir",
	"male_transform_to", "male_looktype", "female_transform_to", "female_looktype",
	"abilities",
}

func scanItemType(row interface{ Scan(...interface{}) error }) (*ItemType, error) {
	var it ItemType
	return &it, row.Scan(
		&it.ServerID, &it.ClientID, &it.ItemGroup, &it.ItemTypeVal, &it.Flags, &it.Speed, &it.TopOrder,
		&it.Name, &it.Article, &it.Plural, &it.Description, &it.RuneSpellName,
		&it.Weight, &it.Armor, &it.Defense, &it.ExtraDefense, &it.Attack, &it.ExtraAttack, &it.AttackSpeed,
		&it.RotateTo, &it.ContainerSize, &it.MaxTextLength, &it.WriteOnceItemID,
		&it.Charges, &it.DecayTo, &it.DecayTime,
		&it.TransformEquipTo, &it.TransformDeequipTo, &it.TransformUseTo,
		&it.Duration, &it.ShowDuration, &it.ShowCharges, &it.ShowCount, &it.ShowAttributes,
		&it.BreakChance, &it.HitChance, &it.MaxHitChance, &it.DualWield,
		&it.ShootRange, &it.Worth, &it.LevelDoor, &it.SpecialDoor, &it.ClosingDoor,
		&it.WareID, &it.ForceSerialize,
		&it.WeaponType, &it.AmmoType, &it.AmmoAction, &it.ShootType, &it.MagicEffect,
		&it.SlotPosition, &it.WieldPosition, &it.FluidSource, &it.CorpseType,
		&it.LightLevel, &it.LightColor, &it.MinimapColor,
		&it.BlockSolid, &it.BlockProjectile, &it.BlockPathFind, &it.AllowDistRead,
		&it.Movable, &it.Pickupable, &it.AllowPickupable,
		&it.IsVertical, &it.IsHorizontal, &it.WalkStack, &it.Replaceable,
		&it.CanWriteText, &it.CanReadText, &it.StopTime,
		&it.Floorchange, &it.BedPartnerDir,
		&it.MaleTransformTo, &it.MaleLooktype, &it.FemaleTransformTo, &it.FemaleLooktype,
		&it.Abilities,
	)
}

func itemValues(it *ItemType) []interface{} {
	return []interface{}{
		it.ServerID, it.ClientID, it.ItemGroup, it.ItemTypeVal, it.Flags, it.Speed, it.TopOrder,
		it.Name, it.Article, it.Plural, it.Description, it.RuneSpellName,
		it.Weight, it.Armor, it.Defense, it.ExtraDefense, it.Attack, it.ExtraAttack, it.AttackSpeed,
		it.RotateTo, it.ContainerSize, it.MaxTextLength, it.WriteOnceItemID,
		it.Charges, it.DecayTo, it.DecayTime,
		it.TransformEquipTo, it.TransformDeequipTo, it.TransformUseTo,
		it.Duration, it.ShowDuration, it.ShowCharges, it.ShowCount, it.ShowAttributes,
		it.BreakChance, it.HitChance, it.MaxHitChance, it.DualWield,
		it.ShootRange, it.Worth, it.LevelDoor, it.SpecialDoor, it.ClosingDoor,
		it.WareID, it.ForceSerialize,
		it.WeaponType, it.AmmoType, it.AmmoAction, it.ShootType, it.MagicEffect,
		it.SlotPosition, it.WieldPosition, it.FluidSource, it.CorpseType,
		it.LightLevel, it.LightColor, it.MinimapColor,
		it.BlockSolid, it.BlockProjectile, it.BlockPathFind, it.AllowDistRead,
		it.Movable, it.Pickupable, it.AllowPickupable,
		it.IsVertical, it.IsHorizontal, it.WalkStack, it.Replaceable,
		it.CanWriteText, it.CanReadText, it.StopTime,
		it.Floorchange, it.BedPartnerDir,
		it.MaleTransformTo, it.MaleLooktype, it.FemaleTransformTo, it.FemaleLooktype,
		it.Abilities,
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
