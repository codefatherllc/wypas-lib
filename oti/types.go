package oti

type ItemDatabase struct {
	Version uint32
	Build   string
	Items   []ItemType
}

type ItemType struct {
	ServerID  uint16
	ClientID  uint16
	Group     uint8
	Type      uint8
	Flags     uint32
	Speed     uint16
	TopOrder  int32

	Name          string
	Article       string
	Plural        string
	Description   string
	RuneSpellName string

	Weight          float32
	Armor           int32
	Defense         int32
	ExtraDefense    int32
	Attack          int32
	ExtraAttack     int32
	AttackSpeed     uint32
	RotateTo        int32
	ContainerSize   uint16
	MaxTextLength   uint16
	WriteOnceItemID uint16
	Charges         uint32
	DecayTo         int32
	DecayTime       uint32
	TransformEquipTo   uint16
	TransformDeEquipTo uint16
	TransformUseTo     uint16
	Duration        uint32
	ShowDuration    bool
	ShowCharges     bool
	ShowCount       bool
	ShowAttributes  bool
	BreakChance     int32
	HitChance       int32
	MaxHitChance    int32
	DualWield       bool
	ShootRange      uint32
	Worth           uint32
	LevelDoor       uint32
	SpecialDoor     bool
	ClosingDoor     bool
	WareID          uint16
	ForceSerialize  bool

	WeaponType    uint8
	AmmoType      uint8
	AmmoAction    uint8
	ShootType     uint8
	MagicEffect   uint8
	SlotPosition  uint32
	WieldPosition uint32
	FluidSource   uint8
	CorpseType    uint8

	LightLevel   int32
	LightColor   int32
	MinimapColor uint16

	BlockSolid      bool
	BlockProjectile bool
	BlockPathFind   bool
	AllowDistRead   bool
	Movable         bool
	Pickupable      bool
	AllowPickupable bool
	IsVertical      bool
	IsHorizontal    bool
	WalkStack       bool
	Replaceable     bool
	CanWriteText    bool
	CanReadText     bool
	StopTime        bool
	Cache           bool

	FloorchangeDown    bool
	FloorchangeNorth   bool
	FloorchangeSouth   bool
	FloorchangeEast    bool
	FloorchangeWest    bool
	FloorchangeNorthEx bool
	FloorchangeSouthEx bool
	FloorchangeEastEx  bool
	FloorchangeWestEx  bool

	BedPartnerDir     uint8
	MaleTransformTo   uint16
	MaleLooktype      uint16
	FemaleTransformTo uint16
	FemaleLooktype    uint16

	Abilities *Abilities
	Field     *FieldDefinition

	RandomizeFrom   uint16
	RandomizeTo     uint16
	RandomizeChance int32
}

type Abilities struct {
	Speed        int32
	HealthGain   int32
	HealthTicks  int32
	ManaGain     int32
	ManaTicks    int32
	ManaShield   bool
	Invisible    bool
	Regeneration bool
	PreventLoss  bool
	PreventDrop  bool

	SkillSword  int32
	SkillAxe    int32
	SkillClub   int32
	SkillDist   int32
	SkillFish   int32
	SkillShield int32
	SkillFist   int32

	MaxHealthPoints  int32
	MaxHealthPercent int32
	MaxManaPoints    int32
	MaxManaPercent   int32
	SoulPoints       int32
	SoulPercent      int32
	MagicPoints      int32
	MagicPercent     int32

	IncreaseMagicValue     int32
	IncreaseMagicPercent   int32
	IncreaseHealingValue   int32
	IncreaseHealingPercent int32

	Absorb         CombatValues
	FieldAbsorb    CombatValues
	ReflectPercent CombatValues
	ReflectChance  CombatValues

	SuppressEnergy       bool
	SuppressFire         bool
	SuppressPoison       bool
	SuppressIce          bool
	SuppressHoly         bool
	SuppressDeath        bool
	SuppressDrown        bool
	SuppressPhysical     bool
	SuppressHaste        bool
	SuppressParalyze     bool
	SuppressDrunk        bool
	SuppressRegeneration bool
	SuppressSoul         bool
	SuppressOutfit       bool
	SuppressInvisible    bool
	SuppressInfight      bool
	SuppressExhaust      bool
	SuppressMuted        bool
	SuppressPacified     bool
	SuppressLight        bool
	SuppressAttributes   bool
	SuppressManashield   bool

	ElementType   uint8
	ElementDamage int32
}

type CombatValues struct {
	Physical  int32
	Energy    int32
	Earth     int32
	Fire      int32
	Undefined int32
	LifeDrain int32
	ManaDrain int32
	Healing   int32
	Drown     int32
	Ice       int32
	Holy      int32
	Death     int32
}

type FieldDefinition struct {
	CombatType uint8
	Damages    []FieldDamage
}

type FieldDamage struct {
	Ticks  int32
	Count  int32
	Start  int32
	Damage int32
}
