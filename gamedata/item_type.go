package gamedata

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type ItemType struct {
	ServerID           uint16          `db:"server_id"`
	ClientID           uint16          `db:"client_id"`
	ItemGroup          uint8           `db:"item_group"`
	ItemTypeVal        uint8           `db:"item_type"`
	Flags              uint32          `db:"flags"`
	Speed              uint16          `db:"speed"`
	TopOrder           int8            `db:"top_order"`
	Name               string          `db:"name"`
	Article            string          `db:"article"`
	Plural             string          `db:"plural"`
	Description        string          `db:"description"`
	RuneSpellName      string          `db:"rune_spell_name"`
	Weight             float32         `db:"weight"`
	Armor              int16           `db:"armor"`
	Defense            int16           `db:"defense"`
	ExtraDefense       int16           `db:"extra_defense"`
	Attack             int16           `db:"attack"`
	ExtraAttack        int16           `db:"extra_attack"`
	AttackSpeed        uint32          `db:"attack_speed"`
	RotateTo           uint16          `db:"rotate_to"`
	ContainerSize      uint8           `db:"container_size"`
	MaxTextLength      uint16          `db:"max_text_length"`
	WriteOnceItemID    uint16          `db:"write_once_item_id"`
	Charges            uint32          `db:"charges"`
	DecayTo            *uint16         `db:"decay_to"`
	DecayTime          uint32          `db:"decay_time"`
	TransformEquipTo   uint16          `db:"transform_equip_to"`
	TransformDeequipTo uint16          `db:"transform_deequip_to"`
	TransformUseTo     uint16          `db:"transform_use_to"`
	Duration           uint32          `db:"duration"`
	ShowDuration       bool            `db:"show_duration"`
	ShowCharges        bool            `db:"show_charges"`
	ShowCount          bool            `db:"show_count"`
	ShowAttributes     bool            `db:"show_attributes"`
	BreakChance        int8            `db:"break_chance"`
	HitChance          int8            `db:"hit_chance"`
	MaxHitChance       int8            `db:"max_hit_chance"`
	DualWield          bool            `db:"dual_wield"`
	ShootRange         uint8           `db:"shoot_range"`
	Worth              uint32          `db:"worth"`
	LevelDoor          uint32          `db:"level_door"`
	SpecialDoor        bool            `db:"special_door"`
	ClosingDoor        bool            `db:"closing_door"`
	WareID             uint16          `db:"ware_id"`
	ForceSerialize     bool            `db:"force_serialize"`
	WeaponType         uint8           `db:"weapon_type"`
	AmmoType           uint8           `db:"ammo_type"`
	AmmoAction         uint8           `db:"ammo_action"`
	ShootType          uint8           `db:"shoot_type"`
	MagicEffect        uint8           `db:"magic_effect"`
	SlotPosition       uint32          `db:"slot_position"`
	WieldPosition      uint32          `db:"wield_position"`
	FluidSource        uint8           `db:"fluid_source"`
	CorpseType         uint8           `db:"corpse_type"`
	LightLevel         int16           `db:"light_level"`
	LightColor         int16           `db:"light_color"`
	MinimapColor       uint16          `db:"minimap_color"`
	BlockSolid         bool            `db:"block_solid"`
	BlockProjectile    bool            `db:"block_projectile"`
	BlockPathFind      bool            `db:"block_path_find"`
	AllowDistRead      bool            `db:"allow_dist_read"`
	Movable            bool            `db:"movable"`
	Pickupable         bool            `db:"pickupable"`
	AllowPickupable    bool            `db:"allow_pickupable"`
	IsVertical         bool            `db:"is_vertical"`
	IsHorizontal       bool            `db:"is_horizontal"`
	WalkStack          bool            `db:"walk_stack"`
	Replaceable        bool            `db:"replaceable"`
	CanWriteText       bool            `db:"can_write_text"`
	CanReadText        bool            `db:"can_read_text"`
	StopTime           bool            `db:"stop_time"`
	Floorchange        uint16          `db:"floorchange"`
	BedPartnerDir      uint8           `db:"bed_partner_dir"`
	MaleTransformTo    uint16          `db:"male_transform_to"`
	MaleLooktype       uint16          `db:"male_looktype"`
	FemaleTransformTo  uint16          `db:"female_transform_to"`
	FemaleLooktype     uint16          `db:"female_looktype"`
	Abilities          NullableAbilities `db:"abilities"`
}

// Combat type indices (bitfield values matching server enums)
const (
	CombatPhysical  = 1 << 0
	CombatEnergy    = 1 << 1
	CombatEarth     = 1 << 2
	CombatFire      = 1 << 3
	CombatUndefined = 1 << 4
	CombatLifeDrain = 1 << 5
	CombatManaDrain = 1 << 6
	CombatHealing   = 1 << 7
	CombatDrown     = 1 << 8
	CombatIce       = 1 << 9
	CombatHoly      = 1 << 10
	CombatDeath     = 1 << 11
)

type Abilities struct {
	Speed     int32 `json:"speed,omitempty"`
	HealthGain int32 `json:"healthGain,omitempty"`
	HealthTicks int32 `json:"healthTicks,omitempty"`
	ManaGain   int32 `json:"manaGain,omitempty"`
	ManaTicks  int32 `json:"manaTicks,omitempty"`

	ManaShield   bool `json:"manaShield,omitempty"`
	Invisible    bool `json:"invisible,omitempty"`
	Regeneration bool `json:"regeneration,omitempty"`
	PreventLoss  bool `json:"preventLoss,omitempty"`
	PreventDrop  bool `json:"preventDrop,omitempty"`

	// skills[0..6]: fist, club, sword, axe, dist, shield, fish
	Skills        [7]int32 `json:"skills,omitempty"`
	SkillsPercent [7]int32 `json:"skillsPercent,omitempty"`

	// stats[0..3]: maxHealth, maxMana, soul, magicLevel
	Stats        [4]int32 `json:"stats,omitempty"`
	StatsPercent [4]int32 `json:"statsPercent,omitempty"`

	// increment[0..3]: healingValue, healingPercent, magicValue, magicPercent
	Increment [4]int16 `json:"increment,omitempty"`

	// Keyed by combat type bitfield value (1,2,4,8,...,2048)
	Absorb      map[int]int16 `json:"absorb,omitempty"`
	FieldAbsorb map[int]int16 `json:"fieldAbsorb,omitempty"`

	// reflect[type][combatType] — "percent" and "chance"
	ReflectPercent map[int]int16 `json:"reflectPercent,omitempty"`
	ReflectChance  map[int]int16 `json:"reflectChance,omitempty"`

	ElementType   int   `json:"elementType,omitempty"`
	ElementDamage int16 `json:"elementDamage,omitempty"`

	ConditionSuppressions int32 `json:"conditionSuppressions,omitempty"`
}

// NullableAbilities wraps Abilities for nullable JSON column scanning.
type NullableAbilities struct {
	Abilities
	Valid bool
}

func (a NullableAbilities) Value() (driver.Value, error) {
	if !a.Valid {
		return nil, nil
	}
	return json.Marshal(a.Abilities)
}

func (a *NullableAbilities) Scan(src interface{}) error {
	if src == nil {
		a.Valid = false
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("gamedata: cannot scan %T into NullableAbilities", src)
	}
	a.Valid = true
	return json.Unmarshal(data, &a.Abilities)
}
