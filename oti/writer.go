package oti

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"os"

	flatbuffers "github.com/google/flatbuffers/go"

	fbItems "github.com/codefatherllc/wypas-proto/items"
)

func WriteFile(path string, db *ItemDatabase) error {
	builder := flatbuffers.NewBuilder(1024 * 1024)

	buildStr := builder.CreateString(db.Build)

	itemOffsets := make([]flatbuffers.UOffsetT, len(db.Items))
	for i := range db.Items {
		itemOffsets[i] = buildItemType(builder, &db.Items[i])
	}

	fbItems.ItemDatabaseStartItemsVector(builder, len(itemOffsets))
	for i := len(itemOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(itemOffsets[i])
	}
	itemsVec := builder.EndVector(len(itemOffsets))

	fbItems.ItemDatabaseStart(builder)
	fbItems.ItemDatabaseAddVersion(builder, db.Version)
	fbItems.ItemDatabaseAddBuild(builder, buildStr)
	fbItems.ItemDatabaseAddItems(builder, itemsVec)
	root := fbItems.ItemDatabaseEnd(builder)
	builder.Finish(root)

	payload := builder.FinishedBytes()

	compressed, err := compressGzip(payload)
	if err != nil {
		return fmt.Errorf("oti: compress: %w", err)
	}

	var header [8]byte
	copy(header[:4], magicOTI[:])
	binary.LittleEndian.PutUint16(header[4:6], 1)
	binary.LittleEndian.PutUint16(header[6:8], flagGzip)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("oti: create file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(header[:]); err != nil {
		return fmt.Errorf("oti: write header: %w", err)
	}
	if _, err := f.Write(compressed); err != nil {
		return fmt.Errorf("oti: write payload: %w", err)
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

func buildItemType(b *flatbuffers.Builder, it *ItemType) flatbuffers.UOffsetT {
	nameOff := b.CreateString(it.Name)
	articleOff := b.CreateString(it.Article)
	pluralOff := b.CreateString(it.Plural)
	descOff := b.CreateString(it.Description)
	runeOff := b.CreateString(it.RuneSpellName)

	var abilitiesOff flatbuffers.UOffsetT
	if it.Abilities != nil {
		abilitiesOff = buildAbilities(b, it.Abilities)
	}

	var fieldOff flatbuffers.UOffsetT
	if it.Field != nil {
		fieldOff = buildField(b, it.Field)
	}

	fbItems.ItemTypeStart(b)
	fbItems.ItemTypeAddServerId(b, it.ServerID)
	fbItems.ItemTypeAddClientId(b, it.ClientID)
	fbItems.ItemTypeAddGroup(b, fbItems.ItemGroup(it.Group))
	fbItems.ItemTypeAddSpecialType(b, fbItems.ItemSpecialType(it.Type))
	fbItems.ItemTypeAddFlags(b, it.Flags)
	fbItems.ItemTypeAddSpeed(b, it.Speed)
	fbItems.ItemTypeAddTopOrder(b, it.TopOrder)
	fbItems.ItemTypeAddName(b, nameOff)
	fbItems.ItemTypeAddArticle(b, articleOff)
	fbItems.ItemTypeAddPlural(b, pluralOff)
	fbItems.ItemTypeAddDescription(b, descOff)
	fbItems.ItemTypeAddRuneSpellName(b, runeOff)
	fbItems.ItemTypeAddWeight(b, it.Weight)
	fbItems.ItemTypeAddArmor(b, it.Armor)
	fbItems.ItemTypeAddDefense(b, it.Defense)
	fbItems.ItemTypeAddExtraDefense(b, it.ExtraDefense)
	fbItems.ItemTypeAddAttack(b, it.Attack)
	fbItems.ItemTypeAddExtraAttack(b, it.ExtraAttack)
	fbItems.ItemTypeAddAttackSpeed(b, it.AttackSpeed)
	fbItems.ItemTypeAddRotateTo(b, it.RotateTo)
	fbItems.ItemTypeAddContainerSize(b, it.ContainerSize)
	fbItems.ItemTypeAddMaxTextLength(b, it.MaxTextLength)
	fbItems.ItemTypeAddWriteOnceItemId(b, it.WriteOnceItemID)
	fbItems.ItemTypeAddCharges(b, it.Charges)
	fbItems.ItemTypeAddDecayTo(b, it.DecayTo)
	fbItems.ItemTypeAddDecayTime(b, it.DecayTime)
	fbItems.ItemTypeAddTransformEquipTo(b, it.TransformEquipTo)
	fbItems.ItemTypeAddTransformDeequipTo(b, it.TransformDeEquipTo)
	fbItems.ItemTypeAddTransformUseTo(b, it.TransformUseTo)
	fbItems.ItemTypeAddDuration(b, it.Duration)
	fbItems.ItemTypeAddShowDuration(b, it.ShowDuration)
	fbItems.ItemTypeAddShowCharges(b, it.ShowCharges)
	fbItems.ItemTypeAddShowCount(b, it.ShowCount)
	fbItems.ItemTypeAddShowAttributes(b, it.ShowAttributes)
	fbItems.ItemTypeAddBreakChance(b, it.BreakChance)
	fbItems.ItemTypeAddHitChance(b, it.HitChance)
	fbItems.ItemTypeAddMaxHitChance(b, it.MaxHitChance)
	fbItems.ItemTypeAddDualWield(b, it.DualWield)
	fbItems.ItemTypeAddShootRange(b, it.ShootRange)
	fbItems.ItemTypeAddWorth(b, it.Worth)
	fbItems.ItemTypeAddLevelDoor(b, it.LevelDoor)
	fbItems.ItemTypeAddSpecialDoor(b, it.SpecialDoor)
	fbItems.ItemTypeAddClosingDoor(b, it.ClosingDoor)
	fbItems.ItemTypeAddWareId(b, it.WareID)
	fbItems.ItemTypeAddForceSerialize(b, it.ForceSerialize)
	fbItems.ItemTypeAddWeaponType(b, fbItems.WeaponType(it.WeaponType))
	fbItems.ItemTypeAddAmmoType(b, fbItems.AmmoType(it.AmmoType))
	fbItems.ItemTypeAddAmmoAction(b, fbItems.AmmoAction(it.AmmoAction))
	fbItems.ItemTypeAddShootType(b, fbItems.ShootEffect(it.ShootType))
	fbItems.ItemTypeAddMagicEffect(b, fbItems.MagicEffect(it.MagicEffect))
	fbItems.ItemTypeAddSlotPosition(b, it.SlotPosition)
	fbItems.ItemTypeAddWieldPosition(b, it.WieldPosition)
	fbItems.ItemTypeAddFluidSource(b, fbItems.FluidType(it.FluidSource))
	fbItems.ItemTypeAddCorpseType(b, fbItems.CorpseType(it.CorpseType))
	fbItems.ItemTypeAddLightLevel(b, it.LightLevel)
	fbItems.ItemTypeAddLightColor(b, it.LightColor)
	fbItems.ItemTypeAddMinimapColor(b, it.MinimapColor)
	fbItems.ItemTypeAddBlockSolid(b, it.BlockSolid)
	fbItems.ItemTypeAddBlockProjectile(b, it.BlockProjectile)
	fbItems.ItemTypeAddBlockPathFind(b, it.BlockPathFind)
	fbItems.ItemTypeAddAllowDistRead(b, it.AllowDistRead)
	fbItems.ItemTypeAddMovable(b, it.Movable)
	fbItems.ItemTypeAddPickupable(b, it.Pickupable)
	fbItems.ItemTypeAddAllowPickupable(b, it.AllowPickupable)
	fbItems.ItemTypeAddIsVertical(b, it.IsVertical)
	fbItems.ItemTypeAddIsHorizontal(b, it.IsHorizontal)
	fbItems.ItemTypeAddWalkStack(b, it.WalkStack)
	fbItems.ItemTypeAddReplaceable(b, it.Replaceable)
	fbItems.ItemTypeAddCanWriteText(b, it.CanWriteText)
	fbItems.ItemTypeAddCanReadText(b, it.CanReadText)
	fbItems.ItemTypeAddStopTime(b, it.StopTime)
	fbItems.ItemTypeAddCache(b, it.Cache)
	fbItems.ItemTypeAddFloorchangeDown(b, it.FloorchangeDown)
	fbItems.ItemTypeAddFloorchangeNorth(b, it.FloorchangeNorth)
	fbItems.ItemTypeAddFloorchangeSouth(b, it.FloorchangeSouth)
	fbItems.ItemTypeAddFloorchangeEast(b, it.FloorchangeEast)
	fbItems.ItemTypeAddFloorchangeWest(b, it.FloorchangeWest)
	fbItems.ItemTypeAddFloorchangeNorthEx(b, it.FloorchangeNorthEx)
	fbItems.ItemTypeAddFloorchangeSouthEx(b, it.FloorchangeSouthEx)
	fbItems.ItemTypeAddFloorchangeEastEx(b, it.FloorchangeEastEx)
	fbItems.ItemTypeAddFloorchangeWestEx(b, it.FloorchangeWestEx)
	fbItems.ItemTypeAddBedPartnerDir(b, int8(it.BedPartnerDir))
	fbItems.ItemTypeAddMaleTransformTo(b, it.MaleTransformTo)
	fbItems.ItemTypeAddMaleLooktype(b, it.MaleLooktype)
	fbItems.ItemTypeAddFemaleTransformTo(b, it.FemaleTransformTo)
	fbItems.ItemTypeAddFemaleLooktype(b, it.FemaleLooktype)
	if it.Abilities != nil {
		fbItems.ItemTypeAddAbilities(b, abilitiesOff)
	}
	if it.Field != nil {
		fbItems.ItemTypeAddField(b, fieldOff)
	}
	fbItems.ItemTypeAddRandomizeFrom(b, it.RandomizeFrom)
	fbItems.ItemTypeAddRandomizeTo(b, it.RandomizeTo)
	fbItems.ItemTypeAddRandomizeChance(b, it.RandomizeChance)
	return fbItems.ItemTypeEnd(b)
}

func buildAbilities(b *flatbuffers.Builder, ab *Abilities) flatbuffers.UOffsetT {
	absorbOff := buildCombatValues(b, &ab.Absorb)
	fieldAbsorbOff := buildCombatValues(b, &ab.FieldAbsorb)
	reflectPercentOff := buildCombatValues(b, &ab.ReflectPercent)
	reflectChanceOff := buildCombatValues(b, &ab.ReflectChance)

	fbItems.AbilitiesStart(b)
	fbItems.AbilitiesAddSpeed(b, ab.Speed)
	fbItems.AbilitiesAddHealthGain(b, ab.HealthGain)
	fbItems.AbilitiesAddHealthTicks(b, ab.HealthTicks)
	fbItems.AbilitiesAddManaGain(b, ab.ManaGain)
	fbItems.AbilitiesAddManaTicks(b, ab.ManaTicks)
	fbItems.AbilitiesAddManaShield(b, ab.ManaShield)
	fbItems.AbilitiesAddInvisible(b, ab.Invisible)
	fbItems.AbilitiesAddRegeneration(b, ab.Regeneration)
	fbItems.AbilitiesAddPreventLoss(b, ab.PreventLoss)
	fbItems.AbilitiesAddPreventDrop(b, ab.PreventDrop)
	fbItems.AbilitiesAddSkillSword(b, ab.SkillSword)
	fbItems.AbilitiesAddSkillAxe(b, ab.SkillAxe)
	fbItems.AbilitiesAddSkillClub(b, ab.SkillClub)
	fbItems.AbilitiesAddSkillDist(b, ab.SkillDist)
	fbItems.AbilitiesAddSkillFish(b, ab.SkillFish)
	fbItems.AbilitiesAddSkillShield(b, ab.SkillShield)
	fbItems.AbilitiesAddSkillFist(b, ab.SkillFist)
	fbItems.AbilitiesAddMaxHealthPoints(b, ab.MaxHealthPoints)
	fbItems.AbilitiesAddMaxHealthPercent(b, ab.MaxHealthPercent)
	fbItems.AbilitiesAddMaxManaPoints(b, ab.MaxManaPoints)
	fbItems.AbilitiesAddMaxManaPercent(b, ab.MaxManaPercent)
	fbItems.AbilitiesAddSoulPoints(b, ab.SoulPoints)
	fbItems.AbilitiesAddSoulPercent(b, ab.SoulPercent)
	fbItems.AbilitiesAddMagicPoints(b, ab.MagicPoints)
	fbItems.AbilitiesAddMagicPercent(b, ab.MagicPercent)
	fbItems.AbilitiesAddIncreaseMagicValue(b, ab.IncreaseMagicValue)
	fbItems.AbilitiesAddIncreaseMagicPercent(b, ab.IncreaseMagicPercent)
	fbItems.AbilitiesAddIncreaseHealingValue(b, ab.IncreaseHealingValue)
	fbItems.AbilitiesAddIncreaseHealingPercent(b, ab.IncreaseHealingPercent)
	fbItems.AbilitiesAddAbsorb(b, absorbOff)
	fbItems.AbilitiesAddFieldAbsorb(b, fieldAbsorbOff)
	fbItems.AbilitiesAddReflectPercent(b, reflectPercentOff)
	fbItems.AbilitiesAddReflectChance(b, reflectChanceOff)
	fbItems.AbilitiesAddSuppressEnergy(b, ab.SuppressEnergy)
	fbItems.AbilitiesAddSuppressFire(b, ab.SuppressFire)
	fbItems.AbilitiesAddSuppressPoison(b, ab.SuppressPoison)
	fbItems.AbilitiesAddSuppressIce(b, ab.SuppressIce)
	fbItems.AbilitiesAddSuppressHoly(b, ab.SuppressHoly)
	fbItems.AbilitiesAddSuppressDeath(b, ab.SuppressDeath)
	fbItems.AbilitiesAddSuppressDrown(b, ab.SuppressDrown)
	fbItems.AbilitiesAddSuppressPhysical(b, ab.SuppressPhysical)
	fbItems.AbilitiesAddSuppressHaste(b, ab.SuppressHaste)
	fbItems.AbilitiesAddSuppressParalyze(b, ab.SuppressParalyze)
	fbItems.AbilitiesAddSuppressDrunk(b, ab.SuppressDrunk)
	fbItems.AbilitiesAddSuppressRegeneration(b, ab.SuppressRegeneration)
	fbItems.AbilitiesAddSuppressSoul(b, ab.SuppressSoul)
	fbItems.AbilitiesAddSuppressOutfit(b, ab.SuppressOutfit)
	fbItems.AbilitiesAddSuppressInvisible(b, ab.SuppressInvisible)
	fbItems.AbilitiesAddSuppressInfight(b, ab.SuppressInfight)
	fbItems.AbilitiesAddSuppressExhaust(b, ab.SuppressExhaust)
	fbItems.AbilitiesAddSuppressMuted(b, ab.SuppressMuted)
	fbItems.AbilitiesAddSuppressPacified(b, ab.SuppressPacified)
	fbItems.AbilitiesAddSuppressLight(b, ab.SuppressLight)
	fbItems.AbilitiesAddSuppressAttributes(b, ab.SuppressAttributes)
	fbItems.AbilitiesAddSuppressManashield(b, ab.SuppressManashield)
	fbItems.AbilitiesAddElementType(b, fbItems.CombatType(ab.ElementType))
	fbItems.AbilitiesAddElementDamage(b, ab.ElementDamage)
	return fbItems.AbilitiesEnd(b)
}

func buildCombatValues(b *flatbuffers.Builder, cv *CombatValues) flatbuffers.UOffsetT {
	fbItems.CombatValuesStart(b)
	fbItems.CombatValuesAddPhysical(b, cv.Physical)
	fbItems.CombatValuesAddEnergy(b, cv.Energy)
	fbItems.CombatValuesAddEarth(b, cv.Earth)
	fbItems.CombatValuesAddFire(b, cv.Fire)
	fbItems.CombatValuesAddUndefined(b, cv.Undefined)
	fbItems.CombatValuesAddLifeDrain(b, cv.LifeDrain)
	fbItems.CombatValuesAddManaDrain(b, cv.ManaDrain)
	fbItems.CombatValuesAddHealing(b, cv.Healing)
	fbItems.CombatValuesAddDrown(b, cv.Drown)
	fbItems.CombatValuesAddIce(b, cv.Ice)
	fbItems.CombatValuesAddHoly(b, cv.Holy)
	fbItems.CombatValuesAddDeath(b, cv.Death)
	return fbItems.CombatValuesEnd(b)
}

func buildField(b *flatbuffers.Builder, fd *FieldDefinition) flatbuffers.UOffsetT {
	dmgOffsets := make([]flatbuffers.UOffsetT, len(fd.Damages))
	for i := range fd.Damages {
		fbItems.FieldDamageStart(b)
		fbItems.FieldDamageAddTicks(b, fd.Damages[i].Ticks)
		fbItems.FieldDamageAddCount(b, fd.Damages[i].Count)
		fbItems.FieldDamageAddStart(b, fd.Damages[i].Start)
		fbItems.FieldDamageAddDamage(b, fd.Damages[i].Damage)
		dmgOffsets[i] = fbItems.FieldDamageEnd(b)
	}

	fbItems.FieldDefinitionStartDamagesVector(b, len(dmgOffsets))
	for i := len(dmgOffsets) - 1; i >= 0; i-- {
		b.PrependUOffsetT(dmgOffsets[i])
	}
	dmgVec := b.EndVector(len(dmgOffsets))

	fbItems.FieldDefinitionStart(b)
	fbItems.FieldDefinitionAddCombatType(b, fbItems.CombatType(fd.CombatType))
	fbItems.FieldDefinitionAddDamages(b, dmgVec)
	return fbItems.FieldDefinitionEnd(b)
}
