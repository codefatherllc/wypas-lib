package oti

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	fbOti "github.com/codefatherllc/wypas-proto/oti"
)

var magicOTI = [4]byte{'O', 'T', 'I', 0}

const flagGzip = 1

func ReadFile(path string) (*ItemDatabase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("oti: read file: %w", err)
	}
	if len(data) < 8 {
		return nil, fmt.Errorf("oti: file too short")
	}

	var magic [4]byte
	copy(magic[:], data[:4])
	if magic != magicOTI {
		return nil, fmt.Errorf("oti: invalid magic %q", magic)
	}

	flags := binary.LittleEndian.Uint16(data[6:8])
	payload := data[8:]

	if flags&flagGzip != 0 {
		payload, err = decompressGzip(payload)
		if err != nil {
			return nil, fmt.Errorf("oti: decompress: %w", err)
		}
	}

	fb := fbOti.GetRootAsItemDatabase(payload, 0)
	return convertItemDatabase(fb), nil
}

func decompressGzip(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func convertItemDatabase(fb *fbOti.ItemDatabase) *ItemDatabase {
	db := &ItemDatabase{
		Version: fb.Version(),
		Build:   string(fb.Build()),
	}

	n := fb.ItemsLength()
	if n > 0 {
		db.Items = make([]ItemType, n)
		var fbItem fbOti.ItemType
		for i := 0; i < n; i++ {
			if fb.Items(&fbItem, i) {
				db.Items[i] = convertItemType(&fbItem)
			}
		}
	}
	return db
}

func convertItemType(fb *fbOti.ItemType) ItemType {
	it := ItemType{
		ServerID:           fb.ServerId(),
		ClientID:           fb.ClientId(),
		Group:              uint8(fb.Group()),
		Type:               uint8(fb.SpecialType()),
		Flags:              fb.Flags(),
		Speed:              fb.Speed(),
		TopOrder:           fb.TopOrder(),
		Name:               string(fb.Name()),
		Article:            string(fb.Article()),
		Plural:             string(fb.Plural()),
		Description:        string(fb.Description()),
		RuneSpellName:      string(fb.RuneSpellName()),
		Weight:             fb.Weight(),
		Armor:              fb.Armor(),
		Defense:            fb.Defense(),
		ExtraDefense:       fb.ExtraDefense(),
		Attack:             fb.Attack(),
		ExtraAttack:        fb.ExtraAttack(),
		AttackSpeed:        fb.AttackSpeed(),
		RotateTo:           fb.RotateTo(),
		ContainerSize:      fb.ContainerSize(),
		MaxTextLength:      fb.MaxTextLength(),
		WriteOnceItemID:    fb.WriteOnceItemId(),
		Charges:            fb.Charges(),
		DecayTo:            fb.DecayTo(),
		DecayTime:          fb.DecayTime(),
		TransformEquipTo:   fb.TransformEquipTo(),
		TransformDeEquipTo: fb.TransformDeequipTo(),
		TransformUseTo:     fb.TransformUseTo(),
		Duration:           fb.Duration(),
		ShowDuration:       fb.ShowDuration(),
		ShowCharges:        fb.ShowCharges(),
		ShowCount:          fb.ShowCount(),
		ShowAttributes:     fb.ShowAttributes(),
		BreakChance:        fb.BreakChance(),
		HitChance:          fb.HitChance(),
		MaxHitChance:       fb.MaxHitChance(),
		DualWield:          fb.DualWield(),
		ShootRange:         fb.ShootRange(),
		Worth:              fb.Worth(),
		LevelDoor:          fb.LevelDoor(),
		SpecialDoor:        fb.SpecialDoor(),
		ClosingDoor:        fb.ClosingDoor(),
		WareID:             fb.WareId(),
		ForceSerialize:     fb.ForceSerialize(),
		WeaponType:         uint8(fb.WeaponType()),
		AmmoType:           uint8(fb.AmmoType()),
		AmmoAction:         uint8(fb.AmmoAction()),
		ShootType:          uint8(fb.ShootType()),
		MagicEffect:        uint8(fb.MagicEffect()),
		SlotPosition:       fb.SlotPosition(),
		WieldPosition:      fb.WieldPosition(),
		FluidSource:        uint8(fb.FluidSource()),
		CorpseType:         uint8(fb.CorpseType()),
		LightLevel:         fb.LightLevel(),
		LightColor:         fb.LightColor(),
		MinimapColor:       fb.MinimapColor(),
		BlockSolid:         fb.BlockSolid(),
		BlockProjectile:    fb.BlockProjectile(),
		BlockPathFind:      fb.BlockPathFind(),
		AllowDistRead:      fb.AllowDistRead(),
		Movable:            fb.Movable(),
		Pickupable:         fb.Pickupable(),
		AllowPickupable:    fb.AllowPickupable(),
		IsVertical:         fb.IsVertical(),
		IsHorizontal:       fb.IsHorizontal(),
		WalkStack:          fb.WalkStack(),
		Replaceable:        fb.Replaceable(),
		CanWriteText:       fb.CanWriteText(),
		CanReadText:        fb.CanReadText(),
		StopTime:           fb.StopTime(),
		Cache:              fb.Cache(),
		FloorchangeDown:    fb.FloorchangeDown(),
		FloorchangeNorth:   fb.FloorchangeNorth(),
		FloorchangeSouth:   fb.FloorchangeSouth(),
		FloorchangeEast:    fb.FloorchangeEast(),
		FloorchangeWest:    fb.FloorchangeWest(),
		FloorchangeNorthEx: fb.FloorchangeNorthEx(),
		FloorchangeSouthEx: fb.FloorchangeSouthEx(),
		FloorchangeEastEx:  fb.FloorchangeEastEx(),
		FloorchangeWestEx:  fb.FloorchangeWestEx(),
		BedPartnerDir:      uint8(fb.BedPartnerDir()),
		MaleTransformTo:    fb.MaleTransformTo(),
		MaleLooktype:       fb.MaleLooktype(),
		FemaleTransformTo:  fb.FemaleTransformTo(),
		FemaleLooktype:     fb.FemaleLooktype(),
		RandomizeFrom:      fb.RandomizeFrom(),
		RandomizeTo:        fb.RandomizeTo(),
		RandomizeChance:    fb.RandomizeChance(),
	}

	if fbAb := fb.Abilities(nil); fbAb != nil {
		it.Abilities = convertAbilities(fbAb)
	}
	if fbFd := fb.Field(nil); fbFd != nil {
		it.Field = convertField(fbFd)
	}
	return it
}

func convertAbilities(fb *fbOti.Abilities) *Abilities {
	ab := &Abilities{
		Speed:                  fb.Speed(),
		HealthGain:             fb.HealthGain(),
		HealthTicks:            fb.HealthTicks(),
		ManaGain:               fb.ManaGain(),
		ManaTicks:              fb.ManaTicks(),
		ManaShield:             fb.ManaShield(),
		Invisible:              fb.Invisible(),
		Regeneration:           fb.Regeneration(),
		PreventLoss:            fb.PreventLoss(),
		PreventDrop:            fb.PreventDrop(),
		SkillSword:             fb.SkillSword(),
		SkillAxe:               fb.SkillAxe(),
		SkillClub:              fb.SkillClub(),
		SkillDist:              fb.SkillDist(),
		SkillFish:              fb.SkillFish(),
		SkillShield:            fb.SkillShield(),
		SkillFist:              fb.SkillFist(),
		MaxHealthPoints:        fb.MaxHealthPoints(),
		MaxHealthPercent:       fb.MaxHealthPercent(),
		MaxManaPoints:          fb.MaxManaPoints(),
		MaxManaPercent:         fb.MaxManaPercent(),
		SoulPoints:             fb.SoulPoints(),
		SoulPercent:            fb.SoulPercent(),
		MagicPoints:            fb.MagicPoints(),
		MagicPercent:           fb.MagicPercent(),
		IncreaseMagicValue:     fb.IncreaseMagicValue(),
		IncreaseMagicPercent:   fb.IncreaseMagicPercent(),
		IncreaseHealingValue:   fb.IncreaseHealingValue(),
		IncreaseHealingPercent: fb.IncreaseHealingPercent(),
		SuppressEnergy:         fb.SuppressEnergy(),
		SuppressFire:           fb.SuppressFire(),
		SuppressPoison:         fb.SuppressPoison(),
		SuppressIce:            fb.SuppressIce(),
		SuppressHoly:           fb.SuppressHoly(),
		SuppressDeath:          fb.SuppressDeath(),
		SuppressDrown:          fb.SuppressDrown(),
		SuppressPhysical:       fb.SuppressPhysical(),
		SuppressHaste:          fb.SuppressHaste(),
		SuppressParalyze:       fb.SuppressParalyze(),
		SuppressDrunk:          fb.SuppressDrunk(),
		SuppressRegeneration:   fb.SuppressRegeneration(),
		SuppressSoul:           fb.SuppressSoul(),
		SuppressOutfit:         fb.SuppressOutfit(),
		SuppressInvisible:      fb.SuppressInvisible(),
		SuppressInfight:        fb.SuppressInfight(),
		SuppressExhaust:        fb.SuppressExhaust(),
		SuppressMuted:          fb.SuppressMuted(),
		SuppressPacified:       fb.SuppressPacified(),
		SuppressLight:          fb.SuppressLight(),
		SuppressAttributes:     fb.SuppressAttributes(),
		SuppressManashield:     fb.SuppressManashield(),
		ElementType:            uint8(fb.ElementType()),
		ElementDamage:          fb.ElementDamage(),
	}

	if cv := fb.Absorb(nil); cv != nil {
		ab.Absorb = convertCombatValues(cv)
	}
	if cv := fb.FieldAbsorb(nil); cv != nil {
		ab.FieldAbsorb = convertCombatValues(cv)
	}
	if cv := fb.ReflectPercent(nil); cv != nil {
		ab.ReflectPercent = convertCombatValues(cv)
	}
	if cv := fb.ReflectChance(nil); cv != nil {
		ab.ReflectChance = convertCombatValues(cv)
	}
	return ab
}

func convertCombatValues(fb *fbOti.CombatValues) CombatValues {
	return CombatValues{
		Physical:  fb.Physical(),
		Energy:    fb.Energy(),
		Earth:     fb.Earth(),
		Fire:      fb.Fire(),
		Undefined: fb.Undefined(),
		LifeDrain: fb.LifeDrain(),
		ManaDrain: fb.ManaDrain(),
		Healing:   fb.Healing(),
		Drown:     fb.Drown(),
		Ice:       fb.Ice(),
		Holy:      fb.Holy(),
		Death:     fb.Death(),
	}
}

func convertField(fb *fbOti.FieldDefinition) *FieldDefinition {
	fd := &FieldDefinition{
		CombatType: uint8(fb.CombatType()),
	}
	n := fb.DamagesLength()
	if n > 0 {
		fd.Damages = make([]FieldDamage, n)
		var fbDmg fbOti.FieldDamage
		for i := 0; i < n; i++ {
			if fb.Damages(&fbDmg, i) {
				fd.Damages[i] = FieldDamage{
					Ticks:  fbDmg.Ticks(),
					Count:  fbDmg.Count(),
					Start:  fbDmg.Start(),
					Damage: fbDmg.Damage(),
				}
			}
		}
	}
	return fd
}
