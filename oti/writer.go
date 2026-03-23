package oti

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"os"

	flatbuffers "github.com/google/flatbuffers/go"

	fbOti "github.com/codefatherllc/wypas-proto/oti"
)

func WriteFile(path string, db *ItemDatabase) error {
	builder := flatbuffers.NewBuilder(1024 * 1024)

	buildStr := builder.CreateString(db.Build)

	itemOffsets := make([]flatbuffers.UOffsetT, len(db.Items))
	for i := range db.Items {
		itemOffsets[i] = buildItemType(builder, &db.Items[i])
	}

	fbOti.ItemDatabaseStartItemsVector(builder, len(itemOffsets))
	for i := len(itemOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(itemOffsets[i])
	}
	itemsVec := builder.EndVector(len(itemOffsets))

	fbOti.ItemDatabaseStart(builder)
	fbOti.ItemDatabaseAddVersion(builder, db.Version)
	fbOti.ItemDatabaseAddBuild(builder, buildStr)
	fbOti.ItemDatabaseAddItems(builder, itemsVec)
	root := fbOti.ItemDatabaseEnd(builder)
	builder.Finish(root)

	payload := builder.FinishedBytes()

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("oti: create file: %w", err)
	}
	defer f.Close()

	var header [8]byte
	copy(header[:4], magicOTI[:])
	binary.LittleEndian.PutUint16(header[4:6], 1)
	binary.LittleEndian.PutUint16(header[6:8], flagGzip)
	if _, err := f.Write(header[:]); err != nil {
		return fmt.Errorf("oti: write header: %w", err)
	}

	gz, _ := gzip.NewWriterLevel(f, gzip.BestCompression)
	if _, err := gz.Write(payload); err != nil {
		return fmt.Errorf("oti: write compressed: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("oti: close gzip: %w", err)
	}

	return nil
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

	fbOti.ItemTypeStart(b)
	fbOti.ItemTypeAddServerId(b, it.ServerID)
	fbOti.ItemTypeAddClientId(b, it.ClientID)
	fbOti.ItemTypeAddGroup(b, fbOti.ItemGroup(it.Group))
	fbOti.ItemTypeAddSpecialType(b, fbOti.ItemSpecialType(it.Type))
	fbOti.ItemTypeAddFlags(b, it.Flags)
	fbOti.ItemTypeAddSpeed(b, it.Speed)
	fbOti.ItemTypeAddTopOrder(b, it.TopOrder)
	fbOti.ItemTypeAddName(b, nameOff)
	fbOti.ItemTypeAddArticle(b, articleOff)
	fbOti.ItemTypeAddPlural(b, pluralOff)
	fbOti.ItemTypeAddDescription(b, descOff)
	fbOti.ItemTypeAddRuneSpellName(b, runeOff)
	fbOti.ItemTypeAddWeight(b, it.Weight)
	fbOti.ItemTypeAddArmor(b, it.Armor)
	fbOti.ItemTypeAddDefense(b, it.Defense)
	fbOti.ItemTypeAddExtraDefense(b, it.ExtraDefense)
	fbOti.ItemTypeAddAttack(b, it.Attack)
	fbOti.ItemTypeAddExtraAttack(b, it.ExtraAttack)
	fbOti.ItemTypeAddAttackSpeed(b, it.AttackSpeed)
	fbOti.ItemTypeAddRotateTo(b, it.RotateTo)
	fbOti.ItemTypeAddContainerSize(b, it.ContainerSize)
	fbOti.ItemTypeAddMaxTextLength(b, it.MaxTextLength)
	fbOti.ItemTypeAddWriteOnceItemId(b, it.WriteOnceItemID)
	fbOti.ItemTypeAddCharges(b, it.Charges)
	fbOti.ItemTypeAddDecayTo(b, it.DecayTo)
	fbOti.ItemTypeAddDecayTime(b, it.DecayTime)
	fbOti.ItemTypeAddTransformEquipTo(b, it.TransformEquipTo)
	fbOti.ItemTypeAddTransformDeequipTo(b, it.TransformDeEquipTo)
	fbOti.ItemTypeAddTransformUseTo(b, it.TransformUseTo)
	fbOti.ItemTypeAddDuration(b, it.Duration)
	fbOti.ItemTypeAddShowDuration(b, it.ShowDuration)
	fbOti.ItemTypeAddShowCharges(b, it.ShowCharges)
	fbOti.ItemTypeAddShowCount(b, it.ShowCount)
	fbOti.ItemTypeAddShowAttributes(b, it.ShowAttributes)
	fbOti.ItemTypeAddBreakChance(b, it.BreakChance)
	fbOti.ItemTypeAddHitChance(b, it.HitChance)
	fbOti.ItemTypeAddMaxHitChance(b, it.MaxHitChance)
	fbOti.ItemTypeAddDualWield(b, it.DualWield)
	fbOti.ItemTypeAddShootRange(b, it.ShootRange)
	fbOti.ItemTypeAddWorth(b, it.Worth)
	fbOti.ItemTypeAddLevelDoor(b, it.LevelDoor)
	fbOti.ItemTypeAddSpecialDoor(b, it.SpecialDoor)
	fbOti.ItemTypeAddClosingDoor(b, it.ClosingDoor)
	fbOti.ItemTypeAddWareId(b, it.WareID)
	fbOti.ItemTypeAddForceSerialize(b, it.ForceSerialize)
	fbOti.ItemTypeAddWeaponType(b, fbOti.WeaponType(it.WeaponType))
	fbOti.ItemTypeAddAmmoType(b, fbOti.AmmoType(it.AmmoType))
	fbOti.ItemTypeAddAmmoAction(b, fbOti.AmmoAction(it.AmmoAction))
	fbOti.ItemTypeAddShootType(b, fbOti.ShootEffect(it.ShootType))
	fbOti.ItemTypeAddMagicEffect(b, fbOti.MagicEffect(it.MagicEffect))
	fbOti.ItemTypeAddSlotPosition(b, it.SlotPosition)
	fbOti.ItemTypeAddWieldPosition(b, it.WieldPosition)
	fbOti.ItemTypeAddFluidSource(b, fbOti.FluidType(it.FluidSource))
	fbOti.ItemTypeAddCorpseType(b, fbOti.CorpseType(it.CorpseType))
	fbOti.ItemTypeAddLightLevel(b, it.LightLevel)
	fbOti.ItemTypeAddLightColor(b, it.LightColor)
	fbOti.ItemTypeAddMinimapColor(b, it.MinimapColor)
	fbOti.ItemTypeAddBlockSolid(b, it.BlockSolid)
	fbOti.ItemTypeAddBlockProjectile(b, it.BlockProjectile)
	fbOti.ItemTypeAddBlockPathFind(b, it.BlockPathFind)
	fbOti.ItemTypeAddAllowDistRead(b, it.AllowDistRead)
	fbOti.ItemTypeAddMovable(b, it.Movable)
	fbOti.ItemTypeAddPickupable(b, it.Pickupable)
	fbOti.ItemTypeAddAllowPickupable(b, it.AllowPickupable)
	fbOti.ItemTypeAddIsVertical(b, it.IsVertical)
	fbOti.ItemTypeAddIsHorizontal(b, it.IsHorizontal)
	fbOti.ItemTypeAddWalkStack(b, it.WalkStack)
	fbOti.ItemTypeAddReplaceable(b, it.Replaceable)
	fbOti.ItemTypeAddCanWriteText(b, it.CanWriteText)
	fbOti.ItemTypeAddCanReadText(b, it.CanReadText)
	fbOti.ItemTypeAddStopTime(b, it.StopTime)
	fbOti.ItemTypeAddCache(b, it.Cache)
	fbOti.ItemTypeAddFloorchangeDown(b, it.FloorchangeDown)
	fbOti.ItemTypeAddFloorchangeNorth(b, it.FloorchangeNorth)
	fbOti.ItemTypeAddFloorchangeSouth(b, it.FloorchangeSouth)
	fbOti.ItemTypeAddFloorchangeEast(b, it.FloorchangeEast)
	fbOti.ItemTypeAddFloorchangeWest(b, it.FloorchangeWest)
	fbOti.ItemTypeAddFloorchangeNorthEx(b, it.FloorchangeNorthEx)
	fbOti.ItemTypeAddFloorchangeSouthEx(b, it.FloorchangeSouthEx)
	fbOti.ItemTypeAddFloorchangeEastEx(b, it.FloorchangeEastEx)
	fbOti.ItemTypeAddFloorchangeWestEx(b, it.FloorchangeWestEx)
	fbOti.ItemTypeAddBedPartnerDir(b, int8(it.BedPartnerDir))
	fbOti.ItemTypeAddMaleTransformTo(b, it.MaleTransformTo)
	fbOti.ItemTypeAddMaleLooktype(b, it.MaleLooktype)
	fbOti.ItemTypeAddFemaleTransformTo(b, it.FemaleTransformTo)
	fbOti.ItemTypeAddFemaleLooktype(b, it.FemaleLooktype)
	if it.Abilities != nil {
		fbOti.ItemTypeAddAbilities(b, abilitiesOff)
	}
	if it.Field != nil {
		fbOti.ItemTypeAddField(b, fieldOff)
	}
	fbOti.ItemTypeAddRandomizeFrom(b, it.RandomizeFrom)
	fbOti.ItemTypeAddRandomizeTo(b, it.RandomizeTo)
	fbOti.ItemTypeAddRandomizeChance(b, it.RandomizeChance)
	return fbOti.ItemTypeEnd(b)
}

func buildAbilities(b *flatbuffers.Builder, ab *Abilities) flatbuffers.UOffsetT {
	absorbOff := buildCombatValues(b, &ab.Absorb)
	fieldAbsorbOff := buildCombatValues(b, &ab.FieldAbsorb)
	reflectPercentOff := buildCombatValues(b, &ab.ReflectPercent)
	reflectChanceOff := buildCombatValues(b, &ab.ReflectChance)

	fbOti.AbilitiesStart(b)
	fbOti.AbilitiesAddSpeed(b, ab.Speed)
	fbOti.AbilitiesAddHealthGain(b, ab.HealthGain)
	fbOti.AbilitiesAddHealthTicks(b, ab.HealthTicks)
	fbOti.AbilitiesAddManaGain(b, ab.ManaGain)
	fbOti.AbilitiesAddManaTicks(b, ab.ManaTicks)
	fbOti.AbilitiesAddManaShield(b, ab.ManaShield)
	fbOti.AbilitiesAddInvisible(b, ab.Invisible)
	fbOti.AbilitiesAddRegeneration(b, ab.Regeneration)
	fbOti.AbilitiesAddPreventLoss(b, ab.PreventLoss)
	fbOti.AbilitiesAddPreventDrop(b, ab.PreventDrop)
	fbOti.AbilitiesAddSkillSword(b, ab.SkillSword)
	fbOti.AbilitiesAddSkillAxe(b, ab.SkillAxe)
	fbOti.AbilitiesAddSkillClub(b, ab.SkillClub)
	fbOti.AbilitiesAddSkillDist(b, ab.SkillDist)
	fbOti.AbilitiesAddSkillFish(b, ab.SkillFish)
	fbOti.AbilitiesAddSkillShield(b, ab.SkillShield)
	fbOti.AbilitiesAddSkillFist(b, ab.SkillFist)
	fbOti.AbilitiesAddMaxHealthPoints(b, ab.MaxHealthPoints)
	fbOti.AbilitiesAddMaxHealthPercent(b, ab.MaxHealthPercent)
	fbOti.AbilitiesAddMaxManaPoints(b, ab.MaxManaPoints)
	fbOti.AbilitiesAddMaxManaPercent(b, ab.MaxManaPercent)
	fbOti.AbilitiesAddSoulPoints(b, ab.SoulPoints)
	fbOti.AbilitiesAddSoulPercent(b, ab.SoulPercent)
	fbOti.AbilitiesAddMagicPoints(b, ab.MagicPoints)
	fbOti.AbilitiesAddMagicPercent(b, ab.MagicPercent)
	fbOti.AbilitiesAddIncreaseMagicValue(b, ab.IncreaseMagicValue)
	fbOti.AbilitiesAddIncreaseMagicPercent(b, ab.IncreaseMagicPercent)
	fbOti.AbilitiesAddIncreaseHealingValue(b, ab.IncreaseHealingValue)
	fbOti.AbilitiesAddIncreaseHealingPercent(b, ab.IncreaseHealingPercent)
	fbOti.AbilitiesAddAbsorb(b, absorbOff)
	fbOti.AbilitiesAddFieldAbsorb(b, fieldAbsorbOff)
	fbOti.AbilitiesAddReflectPercent(b, reflectPercentOff)
	fbOti.AbilitiesAddReflectChance(b, reflectChanceOff)
	fbOti.AbilitiesAddSuppressEnergy(b, ab.SuppressEnergy)
	fbOti.AbilitiesAddSuppressFire(b, ab.SuppressFire)
	fbOti.AbilitiesAddSuppressPoison(b, ab.SuppressPoison)
	fbOti.AbilitiesAddSuppressIce(b, ab.SuppressIce)
	fbOti.AbilitiesAddSuppressHoly(b, ab.SuppressHoly)
	fbOti.AbilitiesAddSuppressDeath(b, ab.SuppressDeath)
	fbOti.AbilitiesAddSuppressDrown(b, ab.SuppressDrown)
	fbOti.AbilitiesAddSuppressPhysical(b, ab.SuppressPhysical)
	fbOti.AbilitiesAddSuppressHaste(b, ab.SuppressHaste)
	fbOti.AbilitiesAddSuppressParalyze(b, ab.SuppressParalyze)
	fbOti.AbilitiesAddSuppressDrunk(b, ab.SuppressDrunk)
	fbOti.AbilitiesAddSuppressRegeneration(b, ab.SuppressRegeneration)
	fbOti.AbilitiesAddSuppressSoul(b, ab.SuppressSoul)
	fbOti.AbilitiesAddSuppressOutfit(b, ab.SuppressOutfit)
	fbOti.AbilitiesAddSuppressInvisible(b, ab.SuppressInvisible)
	fbOti.AbilitiesAddSuppressInfight(b, ab.SuppressInfight)
	fbOti.AbilitiesAddSuppressExhaust(b, ab.SuppressExhaust)
	fbOti.AbilitiesAddSuppressMuted(b, ab.SuppressMuted)
	fbOti.AbilitiesAddSuppressPacified(b, ab.SuppressPacified)
	fbOti.AbilitiesAddSuppressLight(b, ab.SuppressLight)
	fbOti.AbilitiesAddSuppressAttributes(b, ab.SuppressAttributes)
	fbOti.AbilitiesAddSuppressManashield(b, ab.SuppressManashield)
	fbOti.AbilitiesAddElementType(b, fbOti.CombatType(ab.ElementType))
	fbOti.AbilitiesAddElementDamage(b, ab.ElementDamage)
	return fbOti.AbilitiesEnd(b)
}

func buildCombatValues(b *flatbuffers.Builder, cv *CombatValues) flatbuffers.UOffsetT {
	fbOti.CombatValuesStart(b)
	fbOti.CombatValuesAddPhysical(b, cv.Physical)
	fbOti.CombatValuesAddEnergy(b, cv.Energy)
	fbOti.CombatValuesAddEarth(b, cv.Earth)
	fbOti.CombatValuesAddFire(b, cv.Fire)
	fbOti.CombatValuesAddUndefined(b, cv.Undefined)
	fbOti.CombatValuesAddLifeDrain(b, cv.LifeDrain)
	fbOti.CombatValuesAddManaDrain(b, cv.ManaDrain)
	fbOti.CombatValuesAddHealing(b, cv.Healing)
	fbOti.CombatValuesAddDrown(b, cv.Drown)
	fbOti.CombatValuesAddIce(b, cv.Ice)
	fbOti.CombatValuesAddHoly(b, cv.Holy)
	fbOti.CombatValuesAddDeath(b, cv.Death)
	return fbOti.CombatValuesEnd(b)
}

func buildField(b *flatbuffers.Builder, fd *FieldDefinition) flatbuffers.UOffsetT {
	dmgOffsets := make([]flatbuffers.UOffsetT, len(fd.Damages))
	for i := range fd.Damages {
		fbOti.FieldDamageStart(b)
		fbOti.FieldDamageAddTicks(b, fd.Damages[i].Ticks)
		fbOti.FieldDamageAddCount(b, fd.Damages[i].Count)
		fbOti.FieldDamageAddStart(b, fd.Damages[i].Start)
		fbOti.FieldDamageAddDamage(b, fd.Damages[i].Damage)
		dmgOffsets[i] = fbOti.FieldDamageEnd(b)
	}

	fbOti.FieldDefinitionStartDamagesVector(b, len(dmgOffsets))
	for i := len(dmgOffsets) - 1; i >= 0; i-- {
		b.PrependUOffsetT(dmgOffsets[i])
	}
	dmgVec := b.EndVector(len(dmgOffsets))

	fbOti.FieldDefinitionStart(b)
	fbOti.FieldDefinitionAddCombatType(b, fbOti.CombatType(fd.CombatType))
	fbOti.FieldDefinitionAddDamages(b, dmgVec)
	return fbOti.FieldDefinitionEnd(b)
}
