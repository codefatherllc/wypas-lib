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
	flagAnimation       = 1 << 24
	flagWalkStack       = 1 << 25
)

// OTB item attribute IDs (from fs/item_loader.hpp)
const (
	otbAttrServerID     = 0x10
	otbAttrClientID     = 0x11
	otbAttrSpeed        = 0x14
	otbAttrTopOrder     = 0x2B
	otbAttrMinimapColor = 0x23
	otbAttrLight        = 0x2A
	otbAttrLight2       = 0x2A
	otbAttrWareID       = 0x2C
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
	slotHead     uint32 = 1
	slotNecklace uint32 = 2
	slotBackpack uint32 = 3
	slotArmor    uint32 = 4
	slotRight    uint32 = 5
	slotLeft     uint32 = 6
	slotLegs     uint32 = 7
	slotFeet     uint32 = 8
	slotRing     uint32 = 9
	slotAmmo     uint32 = 10
	slotPurse    uint32 = 11
)

// Condition bits for suppress (from combat/condition.hpp)
const (
	condPoison     int32 = 1 << 0
	condFire       int32 = 1 << 1
	condEnergy     int32 = 1 << 2
	condBleeding   int32 = 1 << 3
	condHaste      int32 = 1 << 4
	condParalyze   int32 = 1 << 5
	condOutfit     int32 = 1 << 6
	condInvisible  int32 = 1 << 7
	condLight      int32 = 1 << 8
	condManaShield int32 = 1 << 9
	condInfight    int32 = 1 << 10
	condDrunk      int32 = 1 << 11
	condExhaust    int32 = 1 << 12
	condRegen      int32 = 1 << 13
	condSoul       int32 = 1 << 14
	condDrown      int32 = 1 << 15
	condMuted      int32 = 1 << 16
	condAttributes int32 = 1 << 17
	condFreezing   int32 = 1 << 18
	condDazzled    int32 = 1 << 19
	condCursed     int32 = 1 << 20
	condPacified   int32 = 1 << 21
)

type otbItem struct {
	group        uint8
	flags        uint32
	serverID     uint16
	clientID     uint16
	speed        uint16
	topOrder     int8
	lightLvl     uint16
	lightCol     uint16
	wareID       uint16
	minimapColor uint16
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
			ServerID:   oi.serverID,
			ClientID:   oi.clientID,
			ItemGroup:  oi.group,
			Flags:      oi.flags,
			Speed:      oi.speed,
			TopOrder:   oi.topOrder,
			LightLevel: int16(oi.lightLvl),
			LightColor: int16(oi.lightCol),
		}

		attr := gamedata.ItemTypeAttributes{
			WareID:        oi.wareID,
			MinimapColor:  oi.minimapColor,
			ShowCount:     true,
			Replaceable:   true,
			ShootRange:    1,
			SlotPosition:  uint32(slotpHand),
			WieldPosition: slotpHand,
		}

		applyOTBFlags(&it, oi.flags)
		applyOTBGroupType(&it, oi.group)

		if xi, ok := xmlMap[oi.serverID]; ok {
			attr.Name = xi.Name
			attr.Article = xi.Article
			attr.Plural = xi.Plural
			applyXMLAttrs(&it, &attr, xi.Attrs)
		}

		if err := it.Attributes.SetAttributes(&attr); err != nil {
			return nil, fmt.Errorf("marshal attributes for %d: %w", it.ServerID, err)
		}

		result = append(result, it)
	}

	for i := range result {
		a, _ := result[i].Attributes.GetAttributes()
		if a != nil && a.Plural == "" && a.ShowCount && a.Name != "" {
			a.Plural = a.Name + "s"
			result[i].Attributes.SetAttributes(a)
		}
	}

	return result, nil
}

// applyOTBGroupType maps OTB item groups to ItemTypes_t values.
// ItemGroup enum: CONTAINER=2, TELEPORT=7, MAGICFIELD=8, DOOR=13
// ItemTypes_t enum: DEPOT=1, MAILBOX=2, TRASHHOLDER=3, CONTAINER=4, DOOR=5, MAGICFIELD=6, TELEPORT=7
func applyOTBGroupType(it *gamedata.ItemType, group uint8) {
	switch group {
	case 2: // ITEM_GROUP_CONTAINER
		it.ItemTypeVal = 4 // ITEM_TYPE_CONTAINER
	case 7: // ITEM_GROUP_TELEPORT
		it.ItemTypeVal = 7 // ITEM_TYPE_TELEPORT
	case 8: // ITEM_GROUP_MAGICFIELD
		it.ItemTypeVal = 6 // ITEM_TYPE_MAGICFIELD
	case 13: // ITEM_GROUP_DOOR
		it.ItemTypeVal = 5 // ITEM_TYPE_DOOR
	}
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
			case 0x23: // ITEM_ATTR_MINIMAP_COLOR
				if attrLen >= 2 {
					oi.minimapColor, _ = child.GetU16()
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
	if flags&flagAlwaysOnTop != 0 && it.TopOrder == 0 {
		it.TopOrder = 2
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
		it.FloorChange = fc
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

func applyXMLAttrs(it *gamedata.ItemType, attr *gamedata.ItemTypeAttributes, attrs []xmlAttr) {
	for _, a := range attrs {
		key := strings.ToLower(a.Key)
		val := a.Value
		switch key {
		case "weight":
			it.Weight = float32(atoi(val)) / 100.0
		case "attack":
			attr.Attack = int16(atoi(val))
		case "defense":
			attr.Defense = int16(atoi(val))
		case "extradefense", "extradef":
			attr.ExtraDefense = int16(atoi(val))
		case "armor":
			attr.Armor = int16(atoi(val))
		case "rotateto":
			attr.RotateTo = uint16(atoi(val))
		case "containersize":
			it.ContainerSize = uint8(atoi(val))
			if it.ItemGroup == 0 {
				it.ItemGroup = 2   // ITEM_GROUP_CONTAINER
				it.ItemTypeVal = 4 // ITEM_TYPE_CONTAINER
			}
		case "charges":
			it.Charges = uint32(atoi(val))
		case "decayto":
			v := atoi(val)
			if v >= 0 {
				u := uint16(v)
				it.DecayTo = &u
			}
		case "duration":
			it.DecayTime = uint32(atoi(val))
		case "transformequipto", "onequipto":
			attr.TransformEquipTo = uint16(atoi(val))
		case "transformdeequipto", "ondeequipto":
			attr.TransformDeequipTo = uint16(atoi(val))
		case "transformuseto", "transformto", "onuseto":
			attr.TransformUseTo = uint16(atoi(val))
		case "maxhitchance":
			v := atoi(val)
			if v < 0 {
				v = 0
			}
			if v > 100 {
				v = 100
			}
			attr.MaxHitChance = int8(v)
		case "hitchance":
			v := atoi(val)
			if v < -100 {
				v = -100
			}
			if v > 100 {
				v = 100
			}
			attr.HitChance = int8(v)
		case "worth":
			attr.Worth = uint32(atoi(val))
		case "shootrange", "range":
			attr.ShootRange = uint8(atoi(val))
		case "breakchance":
			v := atoi(val)
			if v < 0 {
				v = 0
			}
			if v > 100 {
				v = 100
			}
			attr.BreakChance = int8(v)
		case "leveldoor":
			attr.LevelDoor = uint32(atoi(val))
		case "wareid":
			attr.WareID = uint16(atoi(val))
		case "maxtextlen", "maxtextlength":
			attr.MaxTextLength = uint16(atoi(val))
		case "writeonceitemid":
			attr.WriteOnceItemID = uint16(atoi(val))
		case "attackspeed":
			attr.AttackSpeed = uint32(atoi(val))
		case "extraattack", "extraatk":
			attr.ExtraAttack = int16(atoi(val))
		case "type":
			switch strings.ToLower(val) {
			case "depot":
				it.ItemTypeVal = 1
			case "mailbox":
				it.ItemTypeVal = 2
			case "trashholder":
				it.ItemTypeVal = 3
			case "container":
				it.ItemTypeVal = 4
			case "door":
				it.ItemTypeVal = 5
			case "magicfield":
				it.ItemTypeVal = 6
			case "teleport":
				it.ItemTypeVal = 7
			case "bed":
				it.ItemTypeVal = 8
			case "key":
				it.ItemTypeVal = 9
			case "rune":
				it.ItemTypeVal = 10
			}
		case "clientid":
			it.ClientID = uint16(atoi(val))
			if it.ItemGroup == 14 { // un-deprecate
				it.ItemGroup = 0
			}
		case "description":
			attr.Description = val
		case "text":
			attr.Text = val
		case "writer", "author":
			attr.Writer = val
		case "date":
			attr.Date = int32(atoi(val))
		case "runespellname":
			attr.RuneSpellName = val
		case "minimapcolor":
			attr.MinimapColor = uint16(atoi(val))
		case "showduration":
			attr.ShowDuration = parseBool(val)
		case "showcharges":
			attr.ShowCharges = parseBool(val)
		case "showcount":
			attr.ShowCount = parseBool(val)
		case "showattributes":
			attr.ShowAttributes = parseBool(val)
		case "forceserialize", "forceserialization", "forcesave":
			it.MustSerialize = parseBool(val)
		case "dualwield":
			attr.DualWield = parseBool(val)
		case "specialdoor":
			attr.SpecialDoor = parseBool(val)
		case "closingdoor":
			attr.ClosingDoor = parseBool(val)
		case "cache":
			it.Cacheable = val != "0"
		case "blocksolid", "blocking":
			it.SetFlag(gamedata.FlagBlockSolid, val != "0")
		case "blockprojectile":
			it.SetFlag(gamedata.FlagBlockProjectile, val != "0")
		case "blockpathfind", "blockpathing", "blockpath":
			it.SetFlag(gamedata.FlagBlockPathFind, val != "0")
		case "allowdistread", "allowdistanceread":
			it.SetFlag(gamedata.FlagAllowDistRead, val != "0")
		case "movable", "moveable":
			it.SetFlag(gamedata.FlagMovable, val != "0")
		case "pickupable":
			it.SetFlag(gamedata.FlagPickupable, val != "0")
		case "allowpickupable":
			attr.AllowPickupable = val != "0"
		case "vertical", "isvertical":
			it.SetFlag(gamedata.FlagVertical, val != "0")
		case "horizontal", "ishorizontal":
			it.SetFlag(gamedata.FlagHorizontal, val != "0")
		case "walkstack":
			it.SetFlag(gamedata.FlagWalkStack, val != "0")
		case "replacable", "replaceable":
			attr.Replaceable = val != "0"
		case "writeable", "writable":
			attr.CanWriteText = val != "0"
			it.SetFlag(gamedata.FlagReadable, val != "0")
		case "readable":
			it.SetFlag(gamedata.FlagReadable, val != "0")
		case "stopduration":
			attr.StopTime = val != "0"
		case "lightlevel":
			it.LightLevel = int16(atoi(val))
		case "lightcolor":
			it.LightColor = int16(atoi(val))

		case "floorchange":
			it.FloorChange |= floorchangeBit(strings.ToLower(val))

		case "weapontype":
			attr.WeaponType = uint8(weaponTypeVal(strings.ToLower(val)))
		case "ammotype":
			attr.AmmoType = uint8(ammoTypeVal(strings.ToLower(val)))
		case "ammoaction":
			attr.AmmoAction = uint8(ammoActionVal(strings.ToLower(val)))
		case "shoottype":
			attr.ShootType = uint8(shootTypeVal(strings.ToLower(val)))
		case "effect":
			attr.MagicEffect = uint8(magicEffectVal(strings.ToLower(val)))

		case "slottype":
			sp, wp := slotTypeVals(strings.ToLower(val), attr.SlotPosition)
			attr.SlotPosition = sp
			attr.WieldPosition = wp

		case "corpsetype":
			attr.CorpseType = uint8(corpseTypeVal(strings.ToLower(val)))
		case "fluidsource":
			it.FluidSource = uint8(fluidTypeVal(strings.ToLower(val)))

		case "partnerdirection":
			attr.BedPartnerDir = directionFromString(strings.ToLower(val))
		case "maletransformto":
			attr.MaleTransformTo = uint16(atoi(val))
		case "femaletransformto":
			attr.FemaleTransformTo = uint16(atoi(val))
		case "malelooktype":
			attr.MaleLooktype = uint16(atoi(val))
		case "femalelooktype":
			attr.FemaleLooktype = uint16(atoi(val))

		case "speed":
			attr.AbilitySpeed = int32(atoi(val))
		case "invisible":
			attr.Invisible = val != "0"
		case "healthgain":
			attr.HealthGain = int32(atoi(val))
			attr.Regeneration = true
		case "healthticks":
			attr.HealthTicks = int32(atoi(val))
			attr.Regeneration = true
		case "managain":
			attr.ManaGain = int32(atoi(val))
			attr.Regeneration = true
		case "manaticks":
			attr.ManaTicks = int32(atoi(val))
			attr.Regeneration = true
		case "manashield":
			attr.ManaShield = val != "0"
		case "regeneration":
			attr.Regeneration = val != "0"
		case "preventloss":
			attr.PreventLoss = val != "0"
		case "preventdrop":
			attr.PreventDrop = val != "0"

		case "skillfist":
			attr.Skills[0] = int32(atoi(val))
		case "skillclub":
			attr.Skills[1] = int32(atoi(val))
		case "skillsword":
			attr.Skills[2] = int32(atoi(val))
		case "skillaxe":
			attr.Skills[3] = int32(atoi(val))
		case "skilldist":
			attr.Skills[4] = int32(atoi(val))
		case "skillshield":
			attr.Skills[5] = int32(atoi(val))
		case "skillfish":
			attr.Skills[6] = int32(atoi(val))

		case "maxhealthpoints", "maxhitpoints":
			attr.Stats[0] = int32(atoi(val))
		case "maxhealthpercent", "maxhitpointspercent":
			attr.StatsPercent[0] = int32(atoi(val))
		case "maxmanapoints":
			attr.Stats[1] = int32(atoi(val))
		case "maxmanapercent", "maxmanapointspercent":
			attr.StatsPercent[1] = int32(atoi(val))
		case "soul":
			attr.Stats[2] = int32(atoi(val))
		case "soulpercent":
			attr.StatsPercent[2] = int32(atoi(val))
		case "magiclevelpoints", "magicpoints":
			attr.Stats[3] = int32(atoi(val))
		case "magiclevelpercent", "magicpointspercent":
			attr.StatsPercent[3] = int32(atoi(val))

		case "increasehealingvalue", "increasehealvalue":
			attr.Increment[0] = int16(atoi(val))
		case "increasehealingpercent", "increasehealpercent":
			attr.Increment[1] = int16(atoi(val))
		case "increasemagicvalue":
			attr.Increment[2] = int16(atoi(val))
		case "increasemagicpercent":
			attr.Increment[3] = int16(atoi(val))

		// Absorb: keyed by combat type bitfield
		case "absorbpercentall":
			v := int16(atoi(val))
			if attr.Absorb == nil {
				attr.Absorb = make(map[int]int16)
			}
			for _, ct := range allCombatTypes {
				attr.Absorb[ct] += v
			}
		case "absorbpercentphysical":
			ensureAbsorbMap(attr)[gamedata.CombatPhysical] += int16(atoi(val))
		case "absorbpercentenergy":
			ensureAbsorbMap(attr)[gamedata.CombatEnergy] += int16(atoi(val))
		case "absorbpercentfire":
			ensureAbsorbMap(attr)[gamedata.CombatFire] += int16(atoi(val))
		case "absorbpercentpoison", "absorbpercentearth":
			ensureAbsorbMap(attr)[gamedata.CombatEarth] += int16(atoi(val))
		case "absorbpercentice":
			ensureAbsorbMap(attr)[gamedata.CombatIce] += int16(atoi(val))
		case "absorbpercentholy":
			ensureAbsorbMap(attr)[gamedata.CombatHoly] += int16(atoi(val))
		case "absorbpercentdeath":
			ensureAbsorbMap(attr)[gamedata.CombatDeath] += int16(atoi(val))
		case "absorbpercentlifedrain":
			ensureAbsorbMap(attr)[gamedata.CombatLifeDrain] += int16(atoi(val))
		case "absorbpercentmanadrain":
			ensureAbsorbMap(attr)[gamedata.CombatManaDrain] += int16(atoi(val))
		case "absorbpercentdrown":
			ensureAbsorbMap(attr)[gamedata.CombatDrown] += int16(atoi(val))
		case "absorbpercenthealing":
			ensureAbsorbMap(attr)[gamedata.CombatHealing] += int16(atoi(val))
		case "absorbpercentundefined":
			ensureAbsorbMap(attr)[gamedata.CombatUndefined] += int16(atoi(val))

		case "absorbpercentelements":
			v := int16(atoi(val))
			if attr.Absorb == nil {
				attr.Absorb = make(map[int]int16)
			}
			for _, ct := range elementCombatTypes {
				attr.Absorb[ct] += v
			}
		case "absorbpercentmagic":
			v := int16(atoi(val))
			if attr.Absorb == nil {
				attr.Absorb = make(map[int]int16)
			}
			for _, ct := range magicCombatTypes {
				attr.Absorb[ct] += v
			}

		// Field absorb
		case "fieldabsorbpercentenergy":
			ensureFieldAbsorbMap(attr)[gamedata.CombatEnergy] += int16(atoi(val))
		case "fieldabsorbpercentfire":
			ensureFieldAbsorbMap(attr)[gamedata.CombatFire] += int16(atoi(val))
		case "fieldabsorbpercentpoison", "fieldabsorbpercentearth":
			ensureFieldAbsorbMap(attr)[gamedata.CombatEarth] += int16(atoi(val))

		case "reflectpercentphysical":
			ensureReflectPercentMap(attr)[gamedata.CombatPhysical] += int16(atoi(val))
		case "reflectpercentenergy":
			ensureReflectPercentMap(attr)[gamedata.CombatEnergy] += int16(atoi(val))
		case "reflectpercentfire":
			ensureReflectPercentMap(attr)[gamedata.CombatFire] += int16(atoi(val))
		case "reflectpercentearth":
			ensureReflectPercentMap(attr)[gamedata.CombatEarth] += int16(atoi(val))
		case "reflectpercentice":
			ensureReflectPercentMap(attr)[gamedata.CombatIce] += int16(atoi(val))
		case "reflectpercentholy":
			ensureReflectPercentMap(attr)[gamedata.CombatHoly] += int16(atoi(val))
		case "reflectpercentdeath":
			ensureReflectPercentMap(attr)[gamedata.CombatDeath] += int16(atoi(val))
		case "reflectpercentlifedrain":
			ensureReflectPercentMap(attr)[gamedata.CombatLifeDrain] += int16(atoi(val))
		case "reflectpercentmanadrain":
			ensureReflectPercentMap(attr)[gamedata.CombatManaDrain] += int16(atoi(val))
		case "reflectpercentdrown":
			ensureReflectPercentMap(attr)[gamedata.CombatDrown] += int16(atoi(val))
		case "reflectpercenthealing":
			ensureReflectPercentMap(attr)[gamedata.CombatHealing] += int16(atoi(val))
		case "reflectpercentundefined":
			ensureReflectPercentMap(attr)[gamedata.CombatUndefined] += int16(atoi(val))
		case "reflectpercentall":
			v := int16(atoi(val))
			if attr.ReflectPercent == nil {
				attr.ReflectPercent = make(map[int]int16)
			}
			for _, ct := range allCombatTypes {
				attr.ReflectPercent[ct] += v
			}
		case "reflectpercentelements":
			v := int16(atoi(val))
			if attr.ReflectPercent == nil {
				attr.ReflectPercent = make(map[int]int16)
			}
			for _, ct := range elementCombatTypes {
				attr.ReflectPercent[ct] += v
			}
		case "reflectpercentmagic":
			v := int16(atoi(val))
			if attr.ReflectPercent == nil {
				attr.ReflectPercent = make(map[int]int16)
			}
			for _, ct := range magicCombatTypes {
				attr.ReflectPercent[ct] += v
			}

		// Reflect chance
		case "reflectchancephysical":
			ensureReflectChanceMap(attr)[gamedata.CombatPhysical] += int16(atoi(val))
		case "reflectchanceenergy":
			ensureReflectChanceMap(attr)[gamedata.CombatEnergy] += int16(atoi(val))
		case "reflectchancefire":
			ensureReflectChanceMap(attr)[gamedata.CombatFire] += int16(atoi(val))
		case "reflectchanceearth":
			ensureReflectChanceMap(attr)[gamedata.CombatEarth] += int16(atoi(val))
		case "reflectchanceice":
			ensureReflectChanceMap(attr)[gamedata.CombatIce] += int16(atoi(val))
		case "reflectchanceholy":
			ensureReflectChanceMap(attr)[gamedata.CombatHoly] += int16(atoi(val))
		case "reflectchancedeath":
			ensureReflectChanceMap(attr)[gamedata.CombatDeath] += int16(atoi(val))
		case "reflectchancelifedrain":
			ensureReflectChanceMap(attr)[gamedata.CombatLifeDrain] += int16(atoi(val))
		case "reflectchancemanadrain":
			ensureReflectChanceMap(attr)[gamedata.CombatManaDrain] += int16(atoi(val))
		case "reflectchancedrown":
			ensureReflectChanceMap(attr)[gamedata.CombatDrown] += int16(atoi(val))
		case "reflectchancehealing":
			ensureReflectChanceMap(attr)[gamedata.CombatHealing] += int16(atoi(val))
		case "reflectchanceundefined":
			ensureReflectChanceMap(attr)[gamedata.CombatUndefined] += int16(atoi(val))
		case "reflectchanceall":
			v := int16(atoi(val))
			if attr.ReflectChance == nil {
				attr.ReflectChance = make(map[int]int16)
			}
			for _, ct := range allCombatTypes {
				attr.ReflectChance[ct] += v
			}
		case "reflectchanceelements":
			v := int16(atoi(val))
			if attr.ReflectChance == nil {
				attr.ReflectChance = make(map[int]int16)
			}
			for _, ct := range elementCombatTypes {
				attr.ReflectChance[ct] += v
			}
		case "reflectchancemagic":
			v := int16(atoi(val))
			if attr.ReflectChance == nil {
				attr.ReflectChance = make(map[int]int16)
			}
			for _, ct := range magicCombatTypes {
				attr.ReflectChance[ct] += v
			}

		// Element damage
		case "elementphysical":
			attr.ElementType = gamedata.CombatPhysical
			attr.ElementDamage = int16(atoi(val))
		case "elementfire":
			attr.ElementType = gamedata.CombatFire
			attr.ElementDamage = int16(atoi(val))
		case "elementenergy":
			attr.ElementType = gamedata.CombatEnergy
			attr.ElementDamage = int16(atoi(val))
		case "elementearth":
			attr.ElementType = gamedata.CombatEarth
			attr.ElementDamage = int16(atoi(val))
		case "elementice":
			attr.ElementType = gamedata.CombatIce
			attr.ElementDamage = int16(atoi(val))
		case "elementholy":
			attr.ElementType = gamedata.CombatHoly
			attr.ElementDamage = int16(atoi(val))
		case "elementdeath":
			attr.ElementType = gamedata.CombatDeath
			attr.ElementDamage = int16(atoi(val))
		case "elementlifedrain":
			attr.ElementType = gamedata.CombatLifeDrain
			attr.ElementDamage = int16(atoi(val))
		case "elementmanadrain":
			attr.ElementType = gamedata.CombatManaDrain
			attr.ElementDamage = int16(atoi(val))
		case "elementdrown":
			attr.ElementType = gamedata.CombatDrown
			attr.ElementDamage = int16(atoi(val))
		case "elementundefined":
			attr.ElementType = gamedata.CombatUndefined
			attr.ElementDamage = int16(atoi(val))
		case "elementhealing":
			attr.ElementType = gamedata.CombatHealing
			attr.ElementDamage = int16(atoi(val))

		case "field":
			it.ItemGroup = 8   // ITEM_GROUP_MAGICFIELD
			it.ItemTypeVal = 6 // ITEM_TYPE_MAGICFIELD
			attr.FieldCombatType = fieldCombatType(strings.ToLower(val))
			for _, sub := range a.Children {
				subKey := strings.ToLower(sub.Key)
				switch subKey {
				case "ticks":
					attr.FieldTicks = int32(atoi(sub.Value))
				case "count":
					attr.FieldCount = int32(atoi(sub.Value))
				case "start":
					attr.FieldStart = int32(atoi(sub.Value))
				case "damage":
					attr.FieldDamage = int32(atoi(sub.Value))
				}
			}

		case "suppresspoison", "suppressearth":
			if val != "0" {
				attr.ConditionSuppressions |= condPoison
			}
		case "suppressfire", "suppressburn":
			if val != "0" {
				attr.ConditionSuppressions |= condFire
			}
		case "suppressenergy", "suppressshock":
			if val != "0" {
				attr.ConditionSuppressions |= condEnergy
			}
		case "suppressbleeding":
			if val != "0" {
				attr.ConditionSuppressions |= condBleeding
			}
		case "suppresshaste":
			if val != "0" {
				attr.ConditionSuppressions |= condHaste
			}
		case "suppressparalyze":
			if val != "0" {
				attr.ConditionSuppressions |= condParalyze
			}
		case "suppressoutfit":
			if val != "0" {
				attr.ConditionSuppressions |= condOutfit
			}
		case "suppressinvisible":
			if val != "0" {
				attr.ConditionSuppressions |= condInvisible
			}
		case "suppresslight":
			if val != "0" {
				attr.ConditionSuppressions |= condLight
			}
		case "suppressmanashield":
			if val != "0" {
				attr.ConditionSuppressions |= condManaShield
			}
		case "suppressinfight":
			if val != "0" {
				attr.ConditionSuppressions |= condInfight
			}
		case "suppressdrunk":
			if val != "0" {
				attr.ConditionSuppressions |= condDrunk
			}
		case "suppressexhaust":
			if val != "0" {
				attr.ConditionSuppressions |= condExhaust
			}
		case "suppressregeneration":
			if val != "0" {
				attr.ConditionSuppressions |= condRegen
			}
		case "suppresssoul":
			if val != "0" {
				attr.ConditionSuppressions |= condSoul
			}
		case "suppressdrown":
			if val != "0" {
				attr.ConditionSuppressions |= condDrown
			}
		case "suppressmuted":
			if val != "0" {
				attr.ConditionSuppressions |= condMuted
			}
		case "suppressattributes":
			if val != "0" {
				attr.ConditionSuppressions |= condAttributes
			}
		case "suppressice", "suppressfreeze", "suppressfreezing":
			if val != "0" {
				attr.ConditionSuppressions |= condFreezing
			}
		case "suppressholy", "suppressdazzle", "suppressdazzled":
			if val != "0" {
				attr.ConditionSuppressions |= condDazzled
			}
		case "suppressdeath", "suppresscurse", "suppresscursed":
			if val != "0" {
				attr.ConditionSuppressions |= condCursed
			}
		case "suppressphysical", "suppresspacified":
			if val != "0" {
				attr.ConditionSuppressions |= condPacified
			}
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

func ensureAbsorbMap(a *gamedata.ItemTypeAttributes) map[int]int16 {
	if a.Absorb == nil {
		a.Absorb = make(map[int]int16)
	}
	return a.Absorb
}

func ensureFieldAbsorbMap(a *gamedata.ItemTypeAttributes) map[int]int16 {
	if a.FieldAbsorb == nil {
		a.FieldAbsorb = make(map[int]int16)
	}
	return a.FieldAbsorb
}

func ensureReflectPercentMap(a *gamedata.ItemTypeAttributes) map[int]int16 {
	if a.ReflectPercent == nil {
		a.ReflectPercent = make(map[int]int16)
	}
	return a.ReflectPercent
}

func ensureReflectChanceMap(a *gamedata.ItemTypeAttributes) map[int]int16 {
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

// slotTypeVals applies SLOTP_* bitfield and SLOT_* wield position.
// For right-hand/left-hand the C++ clears the opposite bit instead of assigning.
func slotTypeVals(s string, current uint32) (uint32, uint32) {
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
		return current &^ slotpLeft, slotRight
	case "left-hand":
		return current &^ slotpRight, slotLeft
	case "legs":
		return slotpLegs, slotLegs
	case "feet":
		return slotpFeet, slotFeet
	case "ring":
		return slotpRing, slotRing
	case "ammo":
		return slotpAmmo, slotAmmo
	case "two-handed":
		return slotpTwoHand, 0
	case "hand":
		return slotpHand, 0
	case "purse":
		return 0, slotPurse
	}
	return 0, 0
}

func parseBool(val string) bool {
	return val == "1" || strings.EqualFold(val, "true")
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func directionFromString(s string) uint8 {
	switch s {
	case "north", "0":
		return 0
	case "east", "1":
		return 1
	case "south", "2":
		return 2
	case "west", "3":
		return 3
	}
	v, _ := strconv.Atoi(s)
	return uint8(v)
}
