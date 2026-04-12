package gamedata

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type ItemType struct {
	ServerID      uint16       `db:"id"`
	ClientID      uint16       `db:"client_id"`
	ItemGroup     uint8        `db:"item_group"`
	ItemTypeVal   uint8        `db:"item_type"`
	Flags         uint32       `db:"flags"`
	Speed         uint16       `db:"speed"`
	TopOrder      int8         `db:"top_order"`
	FloorChange   uint16       `db:"floor_change"`
	LightLevel    int16        `db:"light_level"`
	LightColor    int16        `db:"light_color"`
	ContainerSize uint8        `db:"container_size"`
	FluidSource   uint8        `db:"fluid_source"`
	DecayTo       *uint16      `db:"decay_to"`
	DecayTime     uint32       `db:"decay_time"`
	Charges       uint32       `db:"charges"`
	Weight        float32      `db:"weight"`
	Cacheable     bool         `db:"cacheable"`
	MustSerialize bool         `db:"must_serialize"`
	Attributes    NullableJSON `db:"attributes"`
}

// ItemTypeAttributes holds everything that moved out of columns into
// the attributes JSON blob. Flat structure — abilities merged in.
type ItemTypeAttributes struct {
	Name          string `json:"name,omitempty"`
	Article       string `json:"article,omitempty"`
	Plural        string `json:"plural,omitempty"`
	Description   string `json:"description,omitempty"`
	RuneSpellName string `json:"runeSpellName,omitempty"`
	Text          string `json:"text,omitempty"`
	Writer        string `json:"writer,omitempty"`

	Armor           int16  `json:"armor,omitempty"`
	Defense         int16  `json:"defense,omitempty"`
	ExtraDefense    int16  `json:"extraDefense,omitempty"`
	Attack          int16  `json:"attack,omitempty"`
	ExtraAttack     int16  `json:"extraAttack,omitempty"`
	AttackSpeed     uint32 `json:"attackSpeed,omitempty"`
	RotateTo        uint16 `json:"rotateTo,omitempty"`
	MaxTextLength   uint16 `json:"maxTextLength,omitempty"`
	WriteOnceItemID uint16 `json:"writeOnceItemId,omitempty"`
	Worth           uint32 `json:"worth,omitempty"`
	LevelDoor       uint32 `json:"levelDoor,omitempty"`
	ShootRange      uint8  `json:"shootRange,omitempty"`
	HitChance       int8   `json:"hitChance,omitempty"`
	MaxHitChance    int8   `json:"maxHitChance,omitempty"`
	BreakChance     int8   `json:"breakChance,omitempty"`

	ShowDuration    bool `json:"showDuration,omitempty"`
	ShowCharges     bool `json:"showCharges,omitempty"`
	ShowCount       bool `json:"showCount,omitempty"`
	ShowAttributes  bool `json:"showAttributes,omitempty"`
	DualWield       bool `json:"dualWield,omitempty"`
	SpecialDoor     bool `json:"specialDoor,omitempty"`
	ClosingDoor     bool `json:"closingDoor,omitempty"`
	Replaceable     bool `json:"replaceable,omitempty"`
	AllowPickupable bool `json:"allowPickupable,omitempty"`
	StopTime        bool `json:"stopTime,omitempty"`
	CanWriteText    bool `json:"canWriteText,omitempty"`

	WeaponType    uint8 `json:"weaponType,omitempty"`
	AmmoType      uint8 `json:"ammoType,omitempty"`
	AmmoAction    uint8 `json:"ammoAction,omitempty"`
	ShootType     uint8 `json:"shootType,omitempty"`
	MagicEffect   uint8 `json:"magicEffect,omitempty"`
	CorpseType    uint8 `json:"corpseType,omitempty"`
	BedPartnerDir uint8 `json:"bedPartnerDir,omitempty"`

	SlotPosition  uint32 `json:"slotPosition,omitempty"`
	WieldPosition uint32 `json:"wieldPosition,omitempty"`
	WareID        uint16 `json:"wareId,omitempty"`
	MinimapColor  uint16 `json:"minimapColor,omitempty"`

	TransformEquipTo   uint16 `json:"transformEquipTo,omitempty"`
	TransformDeequipTo uint16 `json:"transformDeEquipTo,omitempty"`
	TransformUseTo     uint16 `json:"transformUseTo,omitempty"`
	MaleTransformTo    uint16 `json:"maleTransformTo,omitempty"`
	MaleLooktype       uint16 `json:"maleLooktype,omitempty"`
	FemaleTransformTo  uint16 `json:"femaleTransformTo,omitempty"`
	FemaleLooktype     uint16 `json:"femaleLooktype,omitempty"`

	Date int32 `json:"date,omitempty"`

	// Abilities (merged flat)
	AbilitySpeed          int32 `json:"abilitySpeed,omitempty"`
	HealthGain            int32 `json:"healthGain,omitempty"`
	HealthTicks           int32 `json:"healthTicks,omitempty"`
	ManaGain              int32 `json:"manaGain,omitempty"`
	ManaTicks             int32 `json:"manaTicks,omitempty"`
	ManaShield            bool  `json:"manaShield,omitempty"`
	Invisible             bool  `json:"invisible,omitempty"`
	Regeneration          bool  `json:"regeneration,omitempty"`
	PreventLoss           bool  `json:"preventLoss,omitempty"`
	PreventDrop           bool  `json:"preventDrop,omitempty"`
	ConditionSuppressions int32 `json:"conditionSuppressions,omitempty"`
	ElementType           int   `json:"elementType,omitempty"`
	ElementDamage         int16 `json:"elementDamage,omitempty"`

	Skills        [7]int32 `json:"skills,omitempty"`
	SkillsPercent [7]int32 `json:"skillsPercent,omitempty"`
	Stats         [4]int32 `json:"stats,omitempty"`
	StatsPercent  [4]int32 `json:"statsPercent,omitempty"`
	Increment     [4]int16 `json:"increment,omitempty"`

	Absorb         map[int]int16 `json:"absorb,omitempty"`
	FieldAbsorb    map[int]int16 `json:"fieldAbsorb,omitempty"`
	ReflectPercent map[int]int16 `json:"reflectPercent,omitempty"`
	ReflectChance  map[int]int16 `json:"reflectChance,omitempty"`

	FieldCombatType int   `json:"fieldCombatType,omitempty"`
	FieldTicks      int32 `json:"fieldTicks,omitempty"`
	FieldCount      int32 `json:"fieldCount,omitempty"`
	FieldStart      int32 `json:"fieldStart,omitempty"`
	FieldDamage     int32 `json:"fieldDamage,omitempty"`
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

// Item flag bit positions (from OTB itemflags_t).
const (
	FlagBlockSolid      uint32 = 1 << 0
	FlagBlockProjectile uint32 = 1 << 1
	FlagBlockPathFind   uint32 = 1 << 2
	FlagHasHeight       uint32 = 1 << 3
	FlagUsable          uint32 = 1 << 4
	FlagPickupable      uint32 = 1 << 5
	FlagMovable         uint32 = 1 << 6
	FlagStackable       uint32 = 1 << 7
	FlagAlwaysOnTop     uint32 = 1 << 13
	FlagReadable        uint32 = 1 << 14
	FlagRotable         uint32 = 1 << 15
	FlagHangable        uint32 = 1 << 16
	FlagVertical        uint32 = 1 << 17
	FlagHorizontal      uint32 = 1 << 18
	FlagAllowDistRead   uint32 = 1 << 20
	FlagLookThrough     uint32 = 1 << 23
	FlagAnimation       uint32 = 1 << 24
	FlagWalkStack       uint32 = 1 << 25
)

func (it *ItemType) HasFlag(f uint32) bool { return it.Flags&f != 0 }

func (it *ItemType) SetFlag(f uint32, v bool) {
	if v {
		it.Flags |= f
	} else {
		it.Flags &^= f
	}
}

// NullableJSON wraps json.RawMessage for nullable JSON columns.
type NullableJSON struct {
	json.RawMessage
	Valid bool
}

func (n NullableJSON) Value() (driver.Value, error) {
	if !n.Valid || len(n.RawMessage) == 0 {
		return nil, nil
	}
	return []byte(n.RawMessage), nil
}

func (n *NullableJSON) Scan(src interface{}) error {
	if src == nil {
		n.Valid = false
		n.RawMessage = nil
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("gamedata: cannot scan %T into NullableJSON", src)
	}
	n.Valid = true
	n.RawMessage = data
	return nil
}

func (n *NullableJSON) SetAttributes(a *ItemTypeAttributes) error {
	if a == nil {
		n.Valid = false
		n.RawMessage = nil
		return nil
	}
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	n.Valid = true
	n.RawMessage = data
	return nil
}

func (n *NullableJSON) GetAttributes() (*ItemTypeAttributes, error) {
	if !n.Valid || len(n.RawMessage) == 0 {
		return nil, nil
	}
	var a ItemTypeAttributes
	if err := json.Unmarshal(n.RawMessage, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func MarshalAttributes(a *ItemTypeAttributes) (sql.NullString, error) {
	if a == nil {
		return sql.NullString{}, nil
	}
	data, err := json.Marshal(a)
	if err != nil {
		return sql.NullString{}, err
	}
	return sql.NullString{String: string(data), Valid: true}, nil
}
