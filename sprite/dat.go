package sprite

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type DatFile struct {
	Signature    uint32
	ItemCount    uint16
	OutfitCount  uint16
	EffectCount  uint16
	MissileCount uint16
	Items        map[uint16]*DatItem
	Outfits      map[uint16]*DatItem
	Effects      map[uint16]*DatItem
	Missiles     map[uint16]*DatItem
}

type DatItem struct {
	ClientID    uint16
	Width       uint8
	Height      uint8
	ColorLayers uint8
	XDiv        uint8
	YDiv        uint8
	ZDiv        uint8
	AnimCount   uint8
	SpriteIDs   []uint32
	Flags       map[uint8][]byte
}

func parseDAT(path string) (*DatFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open dat: %w", err)
	}
	defer f.Close()

	var sig uint32
	if err := binary.Read(f, binary.LittleEndian, &sig); err != nil {
		return nil, fmt.Errorf("read dat signature: %w", err)
	}

	var itemCount, outfitCount, effectCount, missileCount uint16
	for _, p := range []*uint16{&itemCount, &outfitCount, &effectCount, &missileCount} {
		if err := binary.Read(f, binary.LittleEndian, p); err != nil {
			return nil, fmt.Errorf("read dat header: %w", err)
		}
	}

	dat := &DatFile{
		Signature:    sig,
		ItemCount:    itemCount,
		OutfitCount:  outfitCount,
		EffectCount:  effectCount,
		MissileCount: missileCount,
		Items:        make(map[uint16]*DatItem),
	}

	for id := uint16(100); id <= itemCount; id++ {
		item, err := readDatItem(f, id)
		if err != nil {
			return nil, fmt.Errorf("read item %d: %w", id, err)
		}
		dat.Items[id] = item
	}

	dat.Outfits = make(map[uint16]*DatItem, outfitCount)
	for id := uint16(1); id <= outfitCount; id++ {
		outfit, err := readDatItem(f, id)
		if err != nil {
			return nil, fmt.Errorf("read outfit %d: %w", id, err)
		}
		dat.Outfits[id] = outfit
	}

	dat.Effects = make(map[uint16]*DatItem, effectCount)
	for id := uint16(1); id <= effectCount; id++ {
		effect, err := readDatItem(f, id)
		if err != nil {
			return nil, fmt.Errorf("read effect %d: %w", id, err)
		}
		dat.Effects[id] = effect
	}

	dat.Missiles = make(map[uint16]*DatItem, missileCount)
	for id := uint16(1); id <= missileCount; id++ {
		missile, err := readDatItem(f, id)
		if err != nil {
			return nil, fmt.Errorf("read missile %d: %w", id, err)
		}
		dat.Missiles[id] = missile
	}

	return dat, nil
}

func ParseDATFromReader(r io.Reader) (*DatFile, error) {
	var sig uint32
	if err := binary.Read(r, binary.LittleEndian, &sig); err != nil {
		return nil, fmt.Errorf("read dat signature: %w", err)
	}

	var itemCount, outfitCount, effectCount, missileCount uint16
	for _, p := range []*uint16{&itemCount, &outfitCount, &effectCount, &missileCount} {
		if err := binary.Read(r, binary.LittleEndian, p); err != nil {
			return nil, fmt.Errorf("read dat header: %w", err)
		}
	}

	dat := &DatFile{
		Signature:    sig,
		ItemCount:    itemCount,
		OutfitCount:  outfitCount,
		EffectCount:  effectCount,
		MissileCount: missileCount,
		Items:        make(map[uint16]*DatItem),
	}

	for id := uint16(100); id <= itemCount; id++ {
		item, err := readDatItem(r, id)
		if err != nil {
			return nil, fmt.Errorf("read item %d: %w", id, err)
		}
		dat.Items[id] = item
	}

	dat.Outfits = make(map[uint16]*DatItem, outfitCount)
	for id := uint16(1); id <= outfitCount; id++ {
		outfit, err := readDatItem(r, id)
		if err != nil {
			return nil, fmt.Errorf("read outfit %d: %w", id, err)
		}
		dat.Outfits[id] = outfit
	}

	dat.Effects = make(map[uint16]*DatItem, effectCount)
	for id := uint16(1); id <= effectCount; id++ {
		effect, err := readDatItem(r, id)
		if err != nil {
			return nil, fmt.Errorf("read effect %d: %w", id, err)
		}
		dat.Effects[id] = effect
	}

	dat.Missiles = make(map[uint16]*DatItem, missileCount)
	for id := uint16(1); id <= missileCount; id++ {
		missile, err := readDatItem(r, id)
		if err != nil {
			return nil, fmt.Errorf("read missile %d: %w", id, err)
		}
		dat.Missiles[id] = missile
	}

	return dat, nil
}

func readDatItem(r io.Reader, clientID uint16) (*DatItem, error) {
	flags := make(map[uint8][]byte)

	for {
		var flag uint8
		if err := binary.Read(r, binary.LittleEndian, &flag); err != nil {
			return nil, err
		}
		if flag == 0xFF {
			break
		}

		val, err := readFlagData(r, flag)
		if err != nil {
			return nil, fmt.Errorf("flag 0x%02x: %w", flag, err)
		}
		flags[flag] = val
	}

	var width, height uint8
	if err := binary.Read(r, binary.LittleEndian, &width); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &height); err != nil {
		return nil, err
	}

	if width > 1 || height > 1 {
		var skip uint8
		if err := binary.Read(r, binary.LittleEndian, &skip); err != nil {
			return nil, err
		}
	}

	var colorLayers, xDiv, yDiv, zDiv, animCount uint8
	for _, p := range []*uint8{&colorLayers, &xDiv, &yDiv, &zDiv, &animCount} {
		if err := binary.Read(r, binary.LittleEndian, p); err != nil {
			return nil, err
		}
	}

	spriteCount := int(width) * int(height) * int(colorLayers) * int(xDiv) * int(yDiv) * int(zDiv) * int(animCount)
	spriteIDs := make([]uint32, spriteCount)
	if err := binary.Read(r, binary.LittleEndian, spriteIDs); err != nil {
		return nil, fmt.Errorf("read %d sprite ids: %w", spriteCount, err)
	}

	return &DatItem{
		ClientID:    clientID,
		Width:       width,
		Height:      height,
		ColorLayers: colorLayers,
		XDiv:        xDiv,
		YDiv:        yDiv,
		ZDiv:        zDiv,
		AnimCount:   animCount,
		SpriteIDs:   spriteIDs,
		Flags:       flags,
	}, nil
}

// readFlagData reads the data bytes for a .dat flag attribute.
// Matches the proven PHP parser from htdocs/class.dat.php (works with Wypas custom .dat).
// Uses OTClient flag numbering (not ObjectBuilder format 6).
func readFlagData(r io.Reader, flag uint8) ([]byte, error) {
	switch flag {
	// No data (boolean flags)
	case 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		0x10, 0x11, 0x12, 0x13, 0x14,
		0x16, 0x17,
		0x1A, 0x1B,
		0x1E, 0x1F:
		return nil, nil

	// 2 bytes (single uint16)
	case 0x00, // Ground: speed
		0x08, 0x09, // Writable, WritableOnce: maxTextLen
		0x19,       // Elevation
		0x1C, 0x1D, // MinimapColor, LensHelp
		0x20, // Cloth
		0x22: // Usable
		buf := make([]byte, 2)
		_, err := io.ReadFull(r, buf)
		return buf, err

	// 4 bytes (two uint16)
	case 0x15, // Light: intensity + color
		0x18: // Displacement: x + y
		buf := make([]byte, 4)
		_, err := io.ReadFull(r, buf)
		return buf, err

	// Variable (Market: 3×u16 + nameLen u16 + name + 2×u16)
	case 0x21:
		buf := make([]byte, 6)
		if _, err := io.ReadFull(r, buf); err != nil {
			return buf, err
		}
		var nameLen uint16
		if err := binary.Read(r, binary.LittleEndian, &nameLen); err != nil {
			return buf, err
		}
		nameLenBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(nameLenBytes, nameLen)
		buf = append(buf, nameLenBytes...)
		name := make([]byte, nameLen)
		if _, err := io.ReadFull(r, name); err != nil {
			return append(buf, name...), err
		}
		buf = append(buf, name...)
		tail := make([]byte, 4)
		_, err := io.ReadFull(r, tail)
		return append(buf, tail...), err

	default:
		return nil, fmt.Errorf("unknown dat flag 0x%02x", flag)
	}
}
