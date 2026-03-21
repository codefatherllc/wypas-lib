package sprite

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"os"
)

type SpriteFile struct {
	signature   uint32
	spriteCount uint32
	offsets     []uint32
	data        []byte
}

func parseSPR(path string) (*SpriteFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read spr: %w", err)
	}
	if len(data) < 8 {
		return nil, fmt.Errorf("spr file too small: %d bytes", len(data))
	}

	sig := binary.LittleEndian.Uint32(data[0:4])
	count := binary.LittleEndian.Uint32(data[4:8])

	headerSize := 8
	offsetsEnd := headerSize + int(count)*4
	if offsetsEnd > len(data) {
		return nil, fmt.Errorf("spr offsets exceed file size")
	}

	offsets := make([]uint32, count)
	for i := uint32(0); i < count; i++ {
		offsets[i] = binary.LittleEndian.Uint32(data[headerSize+int(i)*4:])
	}

	return &SpriteFile{
		signature:   sig,
		spriteCount: count,
		offsets:     offsets,
		data:        data,
	}, nil
}

func (s *SpriteFile) GetRGBA(id uint32) (*image.RGBA, error) {
	if id == 0 || id > s.spriteCount {
		return nil, fmt.Errorf("sprite id %d out of range [1, %d]", id, s.spriteCount)
	}

	offset := s.offsets[id-1]
	if offset == 0 {
		return image.NewRGBA(image.Rect(0, 0, 32, 32)), nil
	}

	pos := int(offset)
	if pos+5 > len(s.data) {
		return nil, fmt.Errorf("sprite %d: offset %d past end of data", id, offset)
	}

	pos += 3
	dataSize := int(binary.LittleEndian.Uint16(s.data[pos:]))
	pos += 2

	if pos+dataSize > len(s.data) {
		return nil, fmt.Errorf("sprite %d: data extends past end of file", id)
	}

	return decodeRLE(s.data[pos : pos+dataSize]), nil
}

func decodeRLE(data []byte) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	pixel := 0
	i := 0
	totalPixels := 32 * 32

	for pixel < totalPixels && i+3 < len(data) {
		transparentCount := int(binary.LittleEndian.Uint16(data[i:]))
		i += 2
		coloredCount := int(binary.LittleEndian.Uint16(data[i:]))
		i += 2

		pixel += transparentCount

		for c := 0; c < coloredCount && pixel < totalPixels; c++ {
			if i+2 >= len(data) {
				break
			}
			r, g, b := data[i], data[i+1], data[i+2]
			i += 3
			x := pixel % 32
			y := pixel / 32
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			pixel++
		}
	}

	return img
}

func ParseSPRFromBytes(data []byte) (*SpriteFile, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("spr file too small: %d bytes", len(data))
	}

	sig := binary.LittleEndian.Uint32(data[0:4])
	count := binary.LittleEndian.Uint32(data[4:8])

	headerSize := 8
	offsetsEnd := headerSize + int(count)*4
	if offsetsEnd > len(data) {
		return nil, fmt.Errorf("spr offsets exceed file size")
	}

	offsets := make([]uint32, count)
	for i := uint32(0); i < count; i++ {
		offsets[i] = binary.LittleEndian.Uint32(data[headerSize+int(i)*4:])
	}

	return &SpriteFile{
		signature:   sig,
		spriteCount: count,
		offsets:     offsets,
		data:        data,
	}, nil
}

func (s *SpriteFile) Close() error {
	return nil
}

func (s *SpriteFile) Signature() uint32   { return s.signature }
func (s *SpriteFile) SpriteCount() uint32 { return s.spriteCount }

func (s *SpriteFile) RawSpriteData(id uint32) ([]byte, error) {
	if id == 0 || id > s.spriteCount {
		return nil, fmt.Errorf("sprite id %d out of range [1, %d]", id, s.spriteCount)
	}
	offset := s.offsets[id-1]
	if offset == 0 {
		return nil, nil
	}
	pos := int(offset)
	if pos+5 > len(s.data) {
		return nil, fmt.Errorf("sprite %d: offset past end of data", id)
	}
	colorKey := s.data[pos : pos+3]
	pos += 3
	dataSize := int(binary.LittleEndian.Uint16(s.data[pos:]))
	pos += 2
	if pos+dataSize > len(s.data) {
		return nil, fmt.Errorf("sprite %d: data extends past end of file", id)
	}
	buf := make([]byte, 0, 3+2+dataSize)
	buf = append(buf, colorKey...)
	buf = binary.LittleEndian.AppendUint16(buf, uint16(dataSize))
	buf = append(buf, s.data[pos:pos+dataSize]...)
	return buf, nil
}
