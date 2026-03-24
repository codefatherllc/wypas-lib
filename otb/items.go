package otb

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/codefatherllc/wypas-lib/gamedata"
	"github.com/codefatherllc/wypas-lib/otbm"
)

// OTB itemflags_t (from fs/item_loader.hpp)
const (
	flagBlockSolid      = 1 << 0
	flagBlockProjectile = 1 << 1
	flagBlockPathFind   = 1 << 2
	flagHasHeight       = 1 << 3
	flagUsable          = 1 << 4
	flagPickupable      = 1 << 5
	flagMovable         = 1 << 6
	flagStackable       = 1 << 7
	flagFloorchangeDown = 1 << 8
	flagFloorchangeN    = 1 << 9
	flagFloorchangeE    = 1 << 10
	flagFloorchangeS    = 1 << 11
	flagFloorchangeW    = 1 << 12
	flagAlwaysOnTop     = 1 << 13
	flagReadable        = 1 << 14
	flagRotable         = 1 << 15
	flagHangable        = 1 << 16
	flagVertical        = 1 << 17
	flagHorizontal      = 1 << 18
	flagAllowDistRead   = 1 << 20
	flagLookThrough     = 1 << 23
	flagWalkStack       = 1 << 25
)

// OTB item attribute IDs (from fs/item_loader.hpp)
const (
	otbAttrServerID    = 0x10
	otbAttrClientID    = 0x11
	otbAttrSpeed       = 0x14
	otbAttrTopOrder    = 0x2B
	otbAttrMinimapColor = 0x23
	otbAttrLight       = 0x2A
	otbAttrLight2      = 0x2A
	otbAttrWareID      = 0x2C
)

// SLOTP_* constants (from entity/items.hpp) — bitfield positions
const (
	slotpHead     uint32 = 1 << 0
	slotpNecklace uint32 = 1 << 1
	slotpBackpack uint32 = 1 << 2
	slotpArmor    uint32 = 1 << 3
	slotpRight    uint32 = 1 << 4
	slotpLeft     uint32 = 1 << 5
	slotpLegs     uint32 = 1 << 6
	slotpFeet     uint32 = 1 << 7
	slotpRing     uint32 = 1 << 8
	slotpAmmo     uint32 = 1 << 9
	slotpTwoHand  uint32 = 1 << 10
	slotpHand            = slotpLeft | slotpRight
)

// SLOT_* wield positions (from entity/items.hpp)
const (
	slotHead    uint32 = 1
	slotNecklace uint32 = 2
	slotBackpack uint32 = 3
	slotArmor   uint32 = 4
	slotRight   uint32 = 5
	slotLeft    uint32 = 6
	slotLegs    uint32 = 7
	slotFeet    uint32 = 8
	slotRing    uint32 = 9
	slotAmmo    uint32 = 10
	slotHand    uint32 = 11
	slotTwoHand uint32 = 12
)

// Condition bits for suppress (from combat/condition.hpp)
const (
	condPoison      int32 = 1 << 0
	condFire        int32 = 1 << 1
	condEnergy      int32 = 1 << 2
	condBleeding    int32 = 1 << 3
	condHaste       int32 = 1 << 4
	condParalyze    int32 = 1 << 5
	condOutfit      int32 = 1 << 6
	condInvisible   int32 = 1 << 7
	condLight       int32 = 1 << 8
	condManaShield  int32 = 1 << 9
	condInfight     int32 = 1 << 10
	condDrunk       int32 = 1 << 11
	condExhaust     int32 = 1 << 12
	condRegen       int32 = 1 << 13
	condSoul        int32 = 1 << 14
	condDrown       int32 = 1 << 15
	condMuted       int32 = 1 << 16
	condAttributes  int32 = 1 << 17
	condFreezing    int32 = 1 << 18
	condDazzled     int32 = 1 << 19
	condCursed      int32 = 1 << 20
	condPacified    int32 = 1 << 21
)

type otbItem struct {
	group    uint8
	flags    uint32
	serverID uint16
	clientID uint16
	speed    uint16
	topOrder int8
	lightLvl uint16
	lightCol uint16
	wareID   uint16
}

func LoadItems(otbPath, xmlPath string) ([]gamedata.ItemType, error) {
	otbItems, err := parseOTBFull(otbPath)
	if err != nil {
		return nil, fmt.Errorf("otb: parse %s: %w", otbPath, err)
	}

	xmlData, err := os.ReadFile(xmlPath)
	if err != nil {
		return nil, fmt.Errorf("otb: read %s: %w", xmlPath, err)
	}
	var xi xmlItems
	if err := xml.Unmarshal(xmlData, &xi); err != nil {
		return nil, fmt.Errorf("otb: parse xml: %w", err)
	}

	xmlMap := buildXMLMap(xi.Items)
	result := make([]gamedata.ItemType, 0, len(otbItems))

	for _, oi := range otbItems {
		it := gamedata.ItemType{
			ServerID:    oi.serverID,
			ClientID:    oi.clientID,
			ItemGroup:   oi.group,
			Flags:       oi.flags,
			Speed:       oi.speed,
			TopOrder:    oi.topOrder,
			LightLevel:  int16(oi.lightLvl),
			LightColor:  int16(oi.lightCol),
			WareID:      oi.wareID,
			Movable:     true,
			WalkStack:   true,
			ShowCount:   true,
			Replaceable: true,
			ShootRange:  1,
			SlotPosition: uint32(slotpHand),
			WieldPosition: slotHand,
		}

		applyOTBFlags(&it, oi.flags)

		if xi, ok := xmlMap[oi.serverID]; ok {
			it.Name = xi.Name
			it.Article = xi.Article
			it.Plural = xi.Plural
			applyXMLAttrs(&it, xi.Attrs)
		}

		result = append(result, it)
	}

	return result, nil
}

func parseOTBFull(path string) ([]otbItem, error) {
	root, err := otbm.ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("parse tree: %w", err)
	}

	root.ResetPos()
	if _, err := root.GetU32(); err != nil {
		return nil, fmt.Errorf("read root signature: %w", err)
	}
	if root.Remaining() > 0 {
		attr, _ := root.GetU8()
		if attr == 0x01 {
			size, err := root.GetU16()
			if err != nil {
				return nil, fmt.Errorf("read version info size: %w", err)
			}
			if err := root.Skip(int(size)); err != nil {
				return nil, fmt.Errorf("skip version info: %w", err)
			}
		}
	}

	items := make([]otbItem, 0, len(root.Children))

	for _, child := range root.Children {
		child.ResetPos()

		oi := otbItem{
			group: child.Type,
		}

		flags, err := child.GetU32()
		if err != nil {
			continue
		}
		oi.flags = flags

		for child.Remaining() > 0 {
			attrType, err := child.GetU8()
			if err != nil || attrType == 0 || attrType == 0xFF {
				break
			}
			attrLen, err := child.GetU16()
			if err != nil {
				break
			}

			switch attrType {
			case otbAttrServerID:
				if attrLen >= 2 {
					oi.serverID, _ = child.GetU16()
					if attrLen > 2 {
						child.Skip(int(attrLen) - 2)
					}
				} else {
					child.Skip(int(attrLen))
				}
			case otbAttrClientID:
				if attrLen >= 2 {
					oi.clientID, _ = child.GetU16()
					if attrLen > 2 {
						child.Skip(int(attrLen) - 2)
					}
				} else {
					child.Skip(int(attrLen))
				}
			case 0x14: // ITEM_ATTR_SPEED
				if attrLen >= 2 {
					oi.speed, _ = child.GetU16()
					if attrLen > 2 {
						child.Skip(int(attrLen) - 2)
					}
				} else {
					child.Skip(int(attrLen))
				}
			case 0x2B: // ITEM_ATTR_TOPORDER
				if attrLen >= 1 {
					v, _ := child.GetU8()
					oi.topOrder = int8(v)
					if attrLen > 1 {
						child.Skip(int(attrLen) - 1)
					}
				} else {
					child.Skip(int(attrLen))
				}
			case 0x2A: // ITEM_ATTR_LIGHT2
				if attrLen >= 4 {
					oi.lightLvl, _ = child.GetU16()
					oi.lightCol, _ = child.GetU16()
					if attrLen > 4 {
						child.Skip(int(attrLen) - 4)
					}
				} else {
					child.Skip(int(attrLen))
				}
			case 0x2C: // ITEM_ATTR_WAREID
				if attrLen >= 2 {
					oi.wareID, _ = child.GetU16()
					if attrLen > 2 {
						child.Skip(int(attrLen) - 2)
					}
				} else {
					child.Skip(int(attrLen))
				}
			default:
				child.Skip(int(attrLen))
			}
		}

		if oi.serverID > 0 {
			items = append(items, oi)
		}
	}

	return items, nil
}

func applyOTBFlags(it *gamedata.ItemType, flags uint32) {
	it.BlockSolid = flags&flagBlockSolid != 0
	it.BlockProjectile = flags&flagBlockProjectile != 0
	it.BlockPathFind = flags&flagBlockPathFind != 0
	it.Pickupable = flags&flagPickupable != 0
	it.IsVertical = flags&flagVertical != 0
	it.IsHorizontal = flags&flagHorizontal != 0
	it.AllowDistRead = flags&flagAllowDistRead != 0
	it.CanReadText = flags&flagReadable != 0
	it.WalkStack = flags&flagWalkStack != 0

	if flags&flagMovable != 0 {
		it.Movable = true
	} else {
		it.Movable = false
	}

	var fc uint16
	if flags&flagFloorchangeDown != 0 {
		fc |= gamedata.FloorchangeDown
	}
	if flags&flagFloorchangeN != 0 {
		fc |= gamedata.FloorchangeNorth
	}
	if flags&flagFloorchangeS != 0 {
		fc |= gamedata.FloorchangeSouth
	}
	if flags&flagFloorchangeE != 0 {
		fc |= gamedata.FloorchangeEast
	}
	if flags&flagFloorchangeW != 0 {
		fc |= gamedata.FloorchangeWest
	}
	if fc != 0 {
		it.Floorchange = fc
	}
}

func buildXMLMap(xmlItems []xmlItem) map[uint16]*xmlItem {
	m := make(map[uint16]*xmlItem, len(xmlItems))
	for i := range xmlItems {
		item := &xmlItems[i]
		if item.ID != 0 {
			m[item.ID] = item
		} else if item.FromID != 0 {
			for id := item.FromID; id <= item.ToID; id++ {
				clone := *item
				clone.ID = id
				m[id] = &clone
			}
		}
	}
	return m
}

func applyXMLAttrs(it *gamedata.ItemType, attrs []xmlAttr) {
	var ab *gamedata.Abilities
	hasAbilities := false

	ensureAbilities := func() *gamedata.Abilities {
		if ab == nil {
			ab = &gamedata.Abilities{}
		}
		hasAbilities = true
		return ab
	}

	for _, a := range attrs {
		key := strings.ToLower(a.Key)
		val := a.Value
		switch key {
		case "weight":
			it.Weight = float32(atoi(val)) / 100.0
		case "attack":
			it.Attack = int16(atoi(val))
		case "defense":
			it.Defense = int16(atoi(val))
		case "extradefense", "extradef":
			it.ExtraDefense = int16(atoi(val))
		case "armor":
			it.Armor = int16(atoi(val))
		case "rotateto":
			it.RotateTo = uint16(atoi(val))
		case "containersize":
			it.ContainerSize = uint8(atoi(val))
		case "charges":
			it.Charges = uint32(atoi(val))
		case "decayto":
			v := atoi(val)
			if v >= 0 {
				u := uint16(v)
				it.DecayTo = &u
			}
			// v < 0 means no decay => DecayTo stays nil (M2 fix)
		case "duration":
			it.Duration = uint32(atoi(val))
			it.DecayTime = uint32(atoi(val))
		case "transformequipto", "onequipto":
			it.TransformEquipTo = uint16(atoi(val))
		case "transformdeequipto", "ondeequipto":
			it.TransformDeequipTo = uint16(atoi(val))
		case "transformuseto", "transformto", "onuseto":
			it.TransformUseTo = uint16(atoi(val))
		case "maxhitchance":
			it.MaxHitChance = int8(atoi(val))
		case "hitchance":
			it.HitChance = int8(atoi(val))
		case "worth":
			it.Worth = uint32(atoi(val))
		case "shootrange", "range":
			it.ShootRange = uint8(atoi(val))
		case "breakchance":
			it.BreakChance = int8(atoi(val))
		case "leveldoor":
			it.LevelDoor = uint32(atoi(val))
		case "wareid":
			it.WareID = uint16(atoi(val))
		case "maxtextlen", "maxtextlength":
			it.MaxTextLength = uint16(atoi(val))
		case "writeonceitemid":
			it.WriteOnceItemID = uint16(atoi(val))
		case "attackspeed":
			it.AttackSpeed = uint32(atoi(val))
		case "extraattack", "extraatk":
			it.ExtraAttack = int16(atoi(val))
		case "description":
			it.Description = val
		case "runespellname":
			it.RuneSpellName = val
		case "minimapcolor":
			it.MinimapColor = uint16(atoi(val))
		case "showduration":
			it.ShowDuration = parseBool(val)
		case "showcharges":
			it.ShowCharges = parseBool(val)
		case "showcount":
			it.ShowCount = parseBool(val)
		case "showattributes":
			it.ShowAttributes = parseBool(val)
		case "forceserialize", "forceserialization", "forcesave":
			it.ForceSerialize = parseBool(val)
		case "dualwield":
			it.DualWield = parseBool(val)
		case "specialdoor":
			it.SpecialDoor = parseBool(val)
		case "closingdoor":
			it.ClosingDoor = parseBool(val)
		case "blocksolid", "blocking":
			it.BlockSolid = val != "0"
		case "blockprojectile":
			it.BlockProjectile = val != "0"
		case "blockpathfind", "blockpathing", "blockpath":
			it.BlockPathFind = val != "0"
		case "allowdistread", "allowdistanceread":
			it.AllowDistRead = val != "0"
		case "movable", "moveable":
			it.Movable = val != "0"
		case "pickupable":
			it.Pickupable = val != "0"
		case "allowpickupable":
			it.AllowPickupable = val != "0"
		case "vertical", "isvertical":
			it.IsVertical = val != "0"
		case "horizontal", "ishorizontal":
			it.IsHorizontal = val != "0"
		case "walkstack":
			it.WalkStack = val != "0"
		case "replacable", "replaceable":
			it.Replaceable = val != "0"
		case "writeable", "writable":
			it.CanWriteText = val != "0"
			it.CanReadText = val != "0"
		case "readable":
			it.CanReadText = val != "0"
		case "stopduration":
			it.StopTime = val != "0"
		case "lightlevel":
			it.LightLevel = int16(atoi(val))
		case "lightcolor":
			it.LightColor = int16(atoi(val))

		case "floorchange":
			it.Floorchange |= floorchangeBit(strings.ToLower(val))

		case "weapontype":
			it.WeaponType = uint8(weaponTypeVal(strings.ToLower(val)))
		case "ammotype":
			it.AmmoType = uint8(ammoTypeVal(strings.ToLower(val)))
		case "ammoaction":
			it.AmmoAction = uint8(ammoActionVal(strings.ToLower(val)))
		case "shoottype":
			it.ShootType = uint8(shootTypeVal(strings.ToLower(val)))
		case "effect":
			it.MagicEffect = uint8(magicEffectVal(strings.ToLower(val)))

		case "slottype":
			sp, wp := slotTypeVals(strings.ToLower(val))
			it.SlotPosition = sp
			it.WieldPosition = wp

		case "corpsetype":
			it.CorpseType = uint8(corpseTypeVal(strings.ToLower(val)))
		case "fluidsource":
			it.FluidSource = uint8(fluidTypeVal(strings.ToLower(val)))

		case "partnerdirection":
			it.BedPartnerDir = uint8(atoi(val))
		case "maletransformto":
			it.MaleTransformTo = uint16(atoi(val))
		case "femaletransformto":
			it.FemaleTransformTo = uint16(atoi(val))
		case "malelooktype":
			it.MaleLooktype = uint16(atoi(val))
		case "femalelooktype":
			it.FemaleLooktype = uint16(atoi(val))

		// Abilities: direct fields
		case "speed":
			ensureAbilities().Speed = int32(atoi(val))
		case "invisible":
			ensureAbilities().Invisible = val != "0"
		case "healthgain":
			ensureAbilities().HealthGain = int32(atoi(val))
		case "healthticks":
			ensureAbilities().HealthTicks = int32(atoi(val))
		case "managain":
			ensureAbilities().ManaGain = int32(atoi(val))
		case "manaticks":
			ensureAbilities().ManaTicks = int32(atoi(val))
		case "manashield":
			ensureAbilities().ManaShield = val != "0"
		case "regeneration":
			ensureAbilities().Regeneration = val != "0"
		case "preventloss":
			ensureAbilities().PreventLoss = val != "0"
		case "preventdrop":
			ensureAbilities().PreventDrop = val != "0"

		// Skills: [0]=fist, [1]=club, [2]=sword, [3]=axe, [4]=dist, [5]=shield, [6]=fish
		case "skillfist":
			ensureAbilities().Skills[0] = int32(atoi(val))
		case "skillclub":
			ensureAbilities().Skills[1] = int32(atoi(val))
		case "skillsword":
			ensureAbilities().Skills[2] = int32(atoi(val))
		case "skillaxe":
			ensureAbilities().Skills[3] = int32(atoi(val))
		case "skilldist":
			ensureAbilities().Skills[4] = int32(atoi(val))
		case "skillshield":
			ensureAbilities().Skills[5] = int32(atoi(val))
		case "skillfish":
			ensureAbilities().Skills[6] = int32(atoi(val))

		// Stats: [0]=maxHealth, [1]=maxMana, [2]=soul, [3]=magicLevel
		case "maxhealthpoints", "maxhitpoints":
			ensureAbilities().Stats[0] = int32(atoi(val))
		case "maxhealthpercent", "maxhitpointspercent":
			ensureAbilities().StatsPercent[0] = int32(atoi(val))
		case "maxmanapoints":
			ensureAbilities().Stats[1] = int32(atoi(val))
		case "maxmanapercent", "maxmanapointspercent":
			ensureAbilities().StatsPercent[1] = int32(atoi(val))
		case "soul":
			ensureAbilities().Stats[2] = int32(atoi(val))
		case "soulpercent":
			ensureAbilities().StatsPercent[2] = int32(atoi(val))
		case "magiclevelpoints", "magicpoints":
			ensureAbilities().Stats[3] = int32(atoi(val))
		case "magiclevelpercent", "magicpointspercent":
			ensureAbilities().StatsPercent[3] = int32(atoi(val))

		// Increment: [0]=healingValue, [1]=healingPercent, [2]=magicValue, [3]=magicPercent
		case "increasehealingvalue", "increasehealvalue":
			ensureAbilities().Increment[0] = int16(atoi(val))
		case "increasehealingpercent", "increasehealpercent":
			ensureAbilities().Increment[1] = int16(atoi(val))
		case "increasemagicvalue":
			ensureAbilities().Increment[2] = int16(atoi(val))
		case "increasemagicpercent":
			ensureAbilities().Increment[3] = int16(atoi(val))

		// Absorb: keyed by combat type bitfield
		case "absorbpercentall":
			v := int16(atoi(val))
			a := ensureAbilities()
			if a.Absorb == nil {
				a.Absorb = make(map[int]int16)
			}
			for _, ct := range allCombatTypes {
				a.Absorb[ct] = v
			}
		case "absorbpercentphysical":
			ensureAbsorbMap(ensureAbilities())[gamedata.CombatPhysical] = int16(atoi(val))
		case "absorbpercentenergy":
			ensureAbsorbMap(ensureAbilities())[gamedata.CombatEnergy] = int16(atoi(val))
		case "absorbpercentfire":
			ensureAbsorbMap(ensureAbilities())[gamedata.CombatFire] = int16(atoi(val))
		case "absorbpercentpoison", "absorbpercentearth":
			ensureAbsorbMap(ensureAbilities())[gamedata.CombatEarth] = int16(atoi(val))
		case "absorbpercentice":
			ensureAbsorbMap(ensureAbilities())[gamedata.CombatIce] = int16(atoi(val))
		case "absorbpercentholy":
			ensureAbsorbMap(ensureAbilities())[gamedata.CombatHoly] = int16(atoi(val))
		case "absorbpercentdeath":
			ensureAbsorbMap(ensureAbilities())[gamedata.CombatDeath] = int16(atoi(val))
		case "absorbpercentlifedrain":
			ensureAbsorbMap(ensureAbilities())[gamedata.CombatLifeDrain] = int16(atoi(val))
		case "absorbpercentmanadrain":
			ensureAbsorbMap(ensureAbilities())[gamedata.CombatManaDrain] = int16(atoi(val))
		case "absorbpercentdrown":
			ensureAbsorbMap(ensureAbilities())[gamedata.CombatDrown] = int16(atoi(val))
		case "absorbpercenthealing":
			ensureAbsorbMap(ensureAbilities())[gamedata.CombatHealing] = int16(atoi(val))
		case "absorbpercentundefined":
			ensureAbsorbMap(ensureAbilities())[gamedata.CombatUndefined] = int16(atoi(val))

		case "absorbpercentelements":
			v := int16(atoi(val))
			a := ensureAbilities()
			if a.Absorb == nil {
				a.Absorb = make(map[int]int16)
			}
			for _, ct := range elementCombatTypes {
				a.Absorb[ct] += v
			}
		case "absorbpercentmagic":
			v := int16(atoi(val))
			a := ensureAbilities()
			if a.Absorb == nil {
				a.Absorb = make(map[int]int16)
			}
			for _, ct := range magicCombatTypes {
				a.Absorb[ct] += v
			}

		// Field absorb
		case "fieldabsorbpercentenergy":
			ensureFieldAbsorbMap(ensureAbilities())[gamedata.CombatEnergy] = int16(atoi(val))
		case "fieldabsorbpercentfire":
			ensureFieldAbsorbMap(ensureAbilities())[gamedata.CombatFire] = int16(atoi(val))
		case "fieldabsorbpercentpoison", "fieldabsorbpercentearth":
			ensureFieldAbsorbMap(ensureAbilities())[gamedata.CombatEarth] = int16(atoi(val))

		// Reflect percent
		case "reflectpercentphysical":
			ensureReflectPercentMap(ensureAbilities())[gamedata.CombatPhysical] = int16(atoi(val))
		case "reflectpercentenergy":
			ensureReflectPercentMap(ensureAbilities())[gamedata.CombatEnergy] = int16(atoi(val))
		case "reflectpercentfire":
			ensureReflectPercentMap(ensureAbilities())[gamedata.CombatFire] = int16(atoi(val))
		case "reflectpercentearth":
			ensureReflectPercentMap(ensureAbilities())[gamedata.CombatEarth] = int16(atoi(val))
		case "reflectpercentice":
			ensureReflectPercentMap(ensureAbilities())[gamedata.CombatIce] = int16(atoi(val))
		case "reflectpercentholy":
			ensureReflectPercentMap(ensureAbilities())[gamedata.CombatHoly] = int16(atoi(val))
		case "reflectpercentdeath":
			ensureReflectPercentMap(ensureAbilities())[gamedata.CombatDeath] = int16(atoi(val))
		case "reflectpercentlifedrain":
			ensureReflectPercentMap(ensureAbilities())[gamedata.CombatLifeDrain] = int16(atoi(val))
		case "reflectpercentmanadrain":
			ensureReflectPercentMap(ensureAbilities())[gamedata.CombatManaDrain] = int16(atoi(val))
		case "reflectpercentdrown":
			ensureReflectPercentMap(ensureAbilities())[gamedata.CombatDrown] = int16(atoi(val))
		case "reflectpercenthealing":
			ensureReflectPercentMap(ensureAbilities())[gamedata.CombatHealing] = int16(atoi(val))
		case "reflectpercentundefined":
			ensureReflectPercentMap(ensureAbilities())[gamedata.CombatUndefined] = int16(atoi(val))
		case "reflectpercentall":
			v := int16(atoi(val))
			a := ensureAbilities()
			if a.ReflectPercent == nil {
				a.ReflectPercent = make(map[int]int16)
			}
			for _, ct := range allCombatTypes {
				a.ReflectPercent[ct] = v
			}
		case "reflectpercentelements":
			v := int16(atoi(val))
			a := ensureAbilities()
			if a.ReflectPercent == nil {
				a.ReflectPercent = make(map[int]int16)
			}
			for _, ct := range elementCombatTypes {
				a.ReflectPercent[ct] += v
			}
		case "reflectpercentmagic":
			v := int16(atoi(val))
			a := ensureAbilities()
			if a.ReflectPercent == nil {
				a.ReflectPercent = make(map[int]int16)
			}
			for _, ct := range magicCombatTypes {
				a.ReflectPercent[ct] += v
			}

		// Reflect chance
		case "reflectchancephysical":
			ensureReflectChanceMap(ensureAbilities())[gamedata.CombatPhysical] = int16(atoi(val))
		case "reflectchanceenergy":
			ensureReflectChanceMap(ensureAbilities())[gamedata.CombatEnergy] = int16(atoi(val))
		case "reflectchancefire":
			ensureReflectChanceMap(ensureAbilities())[gamedata.CombatFire] = int16(atoi(val))
		case "reflectchanceearth":
			ensureReflectChanceMap(ensureAbilities())[gamedata.CombatEarth] = int16(atoi(val))
		case "reflectchanceice":
			ensureReflectChanceMap(ensureAbilities())[gamedata.CombatIce] = int16(atoi(val))
		case "reflectchanceholy":
			ensureReflectChanceMap(ensureAbilities())[gamedata.CombatHoly] = int16(atoi(val))
		case "reflectchancedeath":
			ensureReflectChanceMap(ensureAbilities())[gamedata.CombatDeath] = int16(atoi(val))
		case "reflectchancelifedrain":
			ensureReflectChanceMap(ensureAbilities())[gamedata.CombatLifeDrain] = int16(atoi(val))
		case "reflectchancemanadrain":
			ensureReflectChanceMap(ensureAbilities())[gamedata.CombatManaDrain] = int16(atoi(val))
		case "reflectchancedrown":
			ensureReflectChanceMap(ensureAbilities())[gamedata.CombatDrown] = int16(atoi(val))
		case "reflectchancehealing":
			ensureReflectChanceMap(ensureAbilities())[gamedata.CombatHealing] = int16(atoi(val))
		case "reflectchanceundefined":
			ensureReflectChanceMap(ensureAbilities())[gamedata.CombatUndefined] = int16(atoi(val))
		case "reflectchanceall":
			v := int16(atoi(val))
			a := ensureAbilities()
			if a.ReflectChance == nil {
				a.ReflectChance = make(map[int]int16)
			}
			for _, ct := range allCombatTypes {
				a.ReflectChance[ct] = v
			}
		case "reflectchanceelements":
			v := int16(atoi(val))
			a := ensureAbilities()
			if a.ReflectChance == nil {
				a.ReflectChance = make(map[int]int16)
			}
			for _, ct := range elementCombatTypes {
				a.ReflectChance[ct] += v
			}
		case "reflectchancemagic":
			v := int16(atoi(val))
			a := ensureAbilities()
			if a.ReflectChance == nil {
				a.ReflectChance = make(map[int]int16)
			}
			for _, ct := range magicCombatTypes {
				a.ReflectChance[ct] += v
			}

		// Element damage
		case "elementphysical":
			a := ensureAbilities()
			a.ElementType = gamedata.CombatPhysical
			a.ElementDamage = int16(atoi(val))
		case "elementfire":
			a := ensureAbilities()
			a.ElementType = gamedata.CombatFire
			a.ElementDamage = int16(atoi(val))
		case "elementenergy":
			a := ensureAbilities()
			a.ElementType = gamedata.CombatEnergy
			a.ElementDamage = int16(atoi(val))
		case "elementearth":
			a := ensureAbilities()
			a.ElementType = gamedata.CombatEarth
			a.ElementDamage = int16(atoi(val))
		case "elementice":
			a := ensureAbilities()
			a.ElementType = gamedata.CombatIce
			a.ElementDamage = int16(atoi(val))
		case "elementholy":
			a := ensureAbilities()
			a.ElementType = gamedata.CombatHoly
			a.ElementDamage = int16(atoi(val))
		case "elementdeath":
			a := ensureAbilities()
			a.ElementType = gamedata.CombatDeath
			a.ElementDamage = int16(atoi(val))
		case "elementlifedrain":
			a := ensureAbilities()
			a.ElementType = gamedata.CombatLifeDrain
			a.ElementDamage = int16(atoi(val))
		case "elementmanadrain":
			a := ensureAbilities()
			a.ElementType = gamedata.CombatManaDrain
			a.ElementDamage = int16(atoi(val))
		case "elementdrown":
			a := ensureAbilities()
			a.ElementType = gamedata.CombatDrown
			a.ElementDamage = int16(atoi(val))
		case "elementundefined":
			a := ensureAbilities()
			a.ElementType = gamedata.CombatUndefined
			a.ElementDamage = int16(atoi(val))

		case "field":
			ab := ensureAbilities()
			ab.FieldCombatType = fieldCombatType(strings.ToLower(val))
			for _, sub := range a.Children {
				subKey := strings.ToLower(sub.Key)
				switch subKey {
				case "ticks":
					ab.FieldTicks = int32(atoi(sub.Value))
				case "count":
					ab.FieldCount = int32(atoi(sub.Value))
				case "start":
					ab.FieldStart = int32(atoi(sub.Value))
				case "damage":
					ab.FieldDamage = int32(atoi(sub.Value))
				}
			}

		// Suppress conditions
		case "suppresspoison", "suppressearth":
			ensureAbilities().ConditionSuppressions |= condPoison
		case "suppressfire", "suppressburn":
			ensureAbilities().ConditionSuppressions |= condFire
		case "suppressenergy", "suppressshock":
			ensureAbilities().ConditionSuppressions |= condEnergy
		case "suppressbleeding":
			ensureAbilities().ConditionSuppressions |= condBleeding
		case "suppresshaste":
			ensureAbilities().ConditionSuppressions |= condHaste
		case "suppressparalyze":
			ensureAbilities().ConditionSuppressions |= condParalyze
		case "suppressoutfit":
			ensureAbilities().ConditionSuppressions |= condOutfit
		case "suppressinvisible":
			ensureAbilities().ConditionSuppressions |= condInvisible
		case "suppresslight":
			ensureAbilities().ConditionSuppressions |= condLight
		case "suppressmanashield":
			ensureAbilities().ConditionSuppressions |= condManaShield
		case "suppressinfight":
			ensureAbilities().ConditionSuppressions |= condInfight
		case "suppressdrunk":
			ensureAbilities().ConditionSuppressions |= condDrunk
		case "suppressexhaust":
			ensureAbilities().ConditionSuppressions |= condExhaust
		case "suppressregeneration":
			ensureAbilities().ConditionSuppressions |= condRegen
		case "suppresssoul":
			ensureAbilities().ConditionSuppressions |= condSoul
		case "suppressdrown":
			ensureAbilities().ConditionSuppressions |= condDrown
		case "suppressmuted":
			ensureAbilities().ConditionSuppressions |= condMuted
		case "suppressattributes":
			ensureAbilities().ConditionSuppressions |= condAttributes
		case "suppressice", "suppressfreeze", "suppressfreezing":
			ensureAbilities().ConditionSuppressions |= condFreezing
		case "suppressholy", "suppressdazzle", "suppressdazzled":
			ensureAbilities().ConditionSuppressions |= condDazzled
		case "suppressdeath", "suppresscurse", "suppresscursed":
			ensureAbilities().ConditionSuppressions |= condCursed
		case "suppressphysical", "suppresspacified":
			ensureAbilities().ConditionSuppressions |= condPacified
		}
	}

	if hasAbilities {
		it.Abilities = gamedata.NullableAbilities{
			Abilities: *ab,
			Valid:     true,
		}
	}
}

var allCombatTypes = []int{
	gamedata.CombatPhysical, gamedata.CombatEnergy, gamedata.CombatEarth,
	gamedata.CombatFire, gamedata.CombatUndefined, gamedata.CombatLifeDrain,
	gamedata.CombatManaDrain, gamedata.CombatHealing, gamedata.CombatDrown,
	gamedata.CombatIce, gamedata.CombatHoly, gamedata.CombatDeath,
}

var elementCombatTypes = []int{
	gamedata.CombatEnergy, gamedata.CombatFire, gamedata.CombatEarth, gamedata.CombatIce,
}

var magicCombatTypes = []int{
	gamedata.CombatEnergy, gamedata.CombatFire, gamedata.CombatEarth,
	gamedata.CombatIce, gamedata.CombatHoly, gamedata.CombatDeath,
}

func ensureAbsorbMap(a *gamedata.Abilities) map[int]int16 {
	if a.Absorb == nil {
		a.Absorb = make(map[int]int16)
	}
	return a.Absorb
}

func ensureFieldAbsorbMap(a *gamedata.Abilities) map[int]int16 {
	if a.FieldAbsorb == nil {
		a.FieldAbsorb = make(map[int]int16)
	}
	return a.FieldAbsorb
}

func ensureReflectPercentMap(a *gamedata.Abilities) map[int]int16 {
	if a.ReflectPercent == nil {
		a.ReflectPercent = make(map[int]int16)
	}
	return a.ReflectPercent
}

func ensureReflectChanceMap(a *gamedata.Abilities) map[int]int16 {
	if a.ReflectChance == nil {
		a.ReflectChance = make(map[int]int16)
	}
	return a.ReflectChance
}

func floorchangeBit(val string) uint16 {
	switch val {
	case "down":
		return gamedata.FloorchangeDown
	case "north":
		return gamedata.FloorchangeNorth
	case "south":
		return gamedata.FloorchangeSouth
	case "east":
		return gamedata.FloorchangeEast
	case "west":
		return gamedata.FloorchangeWest
	case "northex":
		return gamedata.FloorchangeNorthEx
	case "southex":
		return gamedata.FloorchangeSouthEx
	case "eastex":
		return gamedata.FloorchangeEastEx
	case "westex":
		return gamedata.FloorchangeWestEx
	}
	return 0
}

func weaponTypeVal(s string) int {
	switch s {
	case "sword":
		return 1
	case "club":
		return 2
	case "axe":
		return 3
	case "shield":
		return 4
	case "distance", "dist":
		return 5
	case "wand", "rod":
		return 6
	case "ammunition", "ammo":
		return 7
	case "fist":
		return 8
	}
	return 0
}

func ammoTypeVal(s string) int {
	m := map[string]int{
		"bolt": 1, "arrow": 2, "poisonarrow": 3, "burstarrow": 4,
		"throwingstar": 5, "throwingknife": 6, "smallstone": 7,
		"largerock": 8, "snowball": 9, "powerbolt": 10, "spear": 11,
	}
	return m[s]
}

func ammoActionVal(s string) int {
	switch s {
	case "removecount", "remove count":
		return 1
	case "removecharge", "remove charge":
		return 2
	case "move":
		return 3
	case "moveback", "move back":
		return 4
	}
	return 0
}

func shootTypeVal(s string) int {
	m := map[string]int{
		"spear": 1, "bolt": 2, "arrow": 3, "fire": 4, "energy": 5,
		"poisonarrow": 6, "burstarrow": 7, "throwingstar": 8,
		"throwingknife": 9, "smallstone": 10, "death": 11,
		"largerock": 12, "snowball": 13, "powerbolt": 14,
		"poison": 15, "infernalbolt": 16, "huntingspear": 17,
		"enchantedspear": 18, "assassinstar": 19, "piercingbolt": 20,
		"earth": 21, "ice": 22, "flamearrow": 23, "holy": 24,
		"etherealspear": 25, "flamingarrow": 26, "shiverarrow": 27,
		"eartharrow": 28, "explosion": 29, "cake": 30,
	}
	return m[s]
}

func magicEffectVal(s string) int {
	m := map[string]int{
		"redspark": 1, "bluebubble": 2, "poff": 3, "yellowspark": 4,
		"explosionarea": 5, "explosiondamage": 6, "firearea": 7,
		"yellowrings": 8, "greenrings": 9, "hitarea": 10,
		"teleport": 11, "energydamage": 12, "energyarea": 13,
		"blackspark": 18, "blueshimmer": 27, "redshimmer": 28,
		"greenspark": 29, "mortarea": 30,
	}
	return m[s]
}

func corpseTypeVal(s string) int {
	switch s {
	case "venom":
		return 1
	case "blood":
		return 2
	case "undead":
		return 3
	case "fire":
		return 4
	case "energy":
		return 5
	}
	return 0
}

func fieldCombatType(s string) int {
	switch s {
	case "fire":
		return gamedata.CombatFire
	case "energy":
		return gamedata.CombatEnergy
	case "poison", "earth":
		return gamedata.CombatEarth
	case "drown":
		return gamedata.CombatDrown
	case "ice":
		return gamedata.CombatIce
	case "holy":
		return gamedata.CombatHoly
	case "death":
		return gamedata.CombatDeath
	}
	return 0
}

func fluidTypeVal(s string) int {
	m := map[string]int{
		"water": 1, "blood": 2, "beer": 3, "slime": 4,
		"lemonade": 5, "milk": 6, "mana": 7, "life": 8,
		"oil": 9, "urine": 10, "coconutmilk": 11, "wine": 12,
		"mud": 13, "fruitjuice": 14, "lava": 15, "rum": 16,
		"swamp": 17, "tea": 18, "mead": 19,
	}
	return m[s]
}

// slotTypeVals returns SLOTP_* bitfield and SLOT_* wield position (C4 fix)
func slotTypeVals(s string) (uint32, uint32) {
	switch s {
	case "head":
		return slotpHead, slotHead
	case "necklace":
		return slotpNecklace, slotNecklace
	case "backpack":
		return slotpBackpack, slotBackpack
	case "body", "armor":
		return slotpArmor, slotArmor
	case "right-hand":
		return slotpRight, slotRight
	case "left-hand":
		return slotpLeft, slotLeft
	case "legs":
		return slotpLegs, slotLegs
	case "feet":
		return slotpFeet, slotFeet
	case "ring":
		return slotpRing, slotRing
	case "ammo":
		return slotpAmmo, slotAmmo
	case "two-handed":
		return slotpTwoHand | slotpLeft | slotpRight, slotTwoHand
	case "hand":
		return slotpHand, slotHand
	}
	return slotpHand, slotHand
}

func parseBool(val string) bool {
	return val == "1" || strings.EqualFold(val, "true")
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
