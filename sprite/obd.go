package sprite

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
)

type OBDFile struct {
	Version uint16
	Entries []OBDEntry
}

type OBDEntry struct {
	Category string
	ID       uint16
	DatItem  *DatItem
	Sprites  map[uint32]*image.RGBA
}

var categoryNames = [4]string{"item", "outfit", "effect", "missile"}

func ParseOBD(r io.Reader) (*OBDFile, error) {
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, fmt.Errorf("read obd header: %w", err)
	}

	var version uint16
	if header[0] == 'O' && header[1] == 'B' && header[2] == 'D' {
		version = uint16(header[3])
	} else {
		version = binary.LittleEndian.Uint16(header[0:2])
	}

	obd := &OBDFile{Version: version}

	for {
		var cat uint8
		if err := binary.Read(r, binary.LittleEndian, &cat); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return nil, fmt.Errorf("read category: %w", err)
		}
		if int(cat) >= len(categoryNames) {
			return nil, fmt.Errorf("invalid category byte: %d", cat)
		}

		var id uint16
		if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
			return nil, fmt.Errorf("read entry id: %w", err)
		}

		datItem, err := readDatItem(r, id)
		if err != nil {
			return nil, fmt.Errorf("read obd entry %s #%d dat: %w", categoryNames[cat], id, err)
		}

		sprites := make(map[uint32]*image.RGBA, len(datItem.SpriteIDs))
		for _, sprID := range datItem.SpriteIDs {
			if sprID == 0 {
				continue
			}
			rgba, err := readOBDSprite(r)
			if err != nil {
				return nil, fmt.Errorf("read obd entry %s #%d sprite %d: %w", categoryNames[cat], id, sprID, err)
			}
			sprites[sprID] = rgba
		}

		obd.Entries = append(obd.Entries, OBDEntry{
			Category: categoryNames[cat],
			ID:       id,
			DatItem:  datItem,
			Sprites:  sprites,
		})
	}

	return obd, nil
}

func readOBDSprite(r io.Reader) (*image.RGBA, error) {
	const pixelDataSize = 32 * 32 * 4
	buf := make([]byte, pixelDataSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("read bgra data: %w", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for i := 0; i < 32*32; i++ {
		off := i * 4
		b, g, red, a := buf[off], buf[off+1], buf[off+2], buf[off+3]
		x := i % 32
		y := i / 32
		img.SetRGBA(x, y, color.RGBA{R: red, G: g, B: b, A: a})
	}
	return img, nil
}
