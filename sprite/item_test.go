package sprite

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func emptySPR(count int) *SpriteFile {
	buf := make([]byte, 8+count*4)
	binary.LittleEndian.PutUint32(buf[0:], 0x12345678)
	binary.LittleEndian.PutUint32(buf[4:], uint32(count))
	spr, err := ParseSPRFromBytes(buf)
	if err != nil {
		panic(err)
	}
	return spr
}

func solidTile(c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

func TestStackPattern(t *testing.T) {
	cases := map[int]int{
		1: 0, 2: 1, 3: 2, 4: 3,
		5: 4, 9: 4,
		10: 5, 24: 5,
		25: 6, 49: 6,
		50: 7, 100: 7,
	}
	for count, want := range cases {
		if got := stackPattern(count); got != want {
			t.Errorf("stackPattern(%d) = %d, want %d", count, got, want)
		}
	}
}

func TestRenderItemMultiTilePlacement(t *testing.T) {
	spr := emptySPR(4)
	colors := []color.RGBA{
		{255, 0, 0, 255},
		{0, 255, 0, 255},
		{0, 0, 255, 255},
		{255, 255, 0, 255},
	}
	for i, col := range colors {
		spr.SetRGBA(uint32(i+1), solidTile(col))
	}

	item := &DatItem{
		Width: 2, Height: 2, ColorLayers: 1,
		XDiv: 1, YDiv: 1, ZDiv: 1, AnimCount: 1,
		SpriteIDs: []uint32{1, 2, 3, 4},
	}
	c := &Cache{spr: spr}

	canvas := c.renderItemRGBA(item, 1)
	if got := canvas.Bounds(); got.Dx() != 64 || got.Dy() != 64 {
		t.Fatalf("canvas = %dx%d, want 64x64", got.Dx(), got.Dy())
	}

	// .dat tile order is width-fastest; tile (ht, wt) lands at
	// ((W-1-wt)*32, (H-1-ht)*32) — sprite 1 bottom-right … sprite 4 top-left.
	quadrants := map[[2]int]color.RGBA{
		{48, 48}: colors[0],
		{16, 48}: colors[1],
		{48, 16}: colors[2],
		{16, 16}: colors[3],
	}
	for pt, want := range quadrants {
		if got := canvas.RGBAAt(pt[0], pt[1]); got != want {
			t.Errorf("pixel at %v = %v, want %v", pt, got, want)
		}
	}
}

func TestRenderItemStackVariant(t *testing.T) {
	spr := emptySPR(8)
	for i := 1; i <= 8; i++ {
		spr.SetRGBA(uint32(i), solidTile(color.RGBA{uint8(i * 10), 0, 0, 255}))
	}

	item := &DatItem{
		Width: 1, Height: 1, ColorLayers: 1,
		XDiv: 4, YDiv: 2, ZDiv: 1, AnimCount: 1,
		SpriteIDs: []uint32{1, 2, 3, 4, 5, 6, 7, 8},
	}
	c := &Cache{spr: spr}

	// count 100 → pattern 7 → xPat 3, yPat 1 → sprite index 7 → id 8.
	canvas := c.renderItemRGBA(item, 100)
	if got := canvas.RGBAAt(4, 4); got != (color.RGBA{80, 0, 0, 255}) {
		t.Errorf("stack-100 pixel = %v, want sprite 8 color {80 0 0 255}", got)
	}

	// count 1 → pattern 0 → sprite id 1.
	canvas = c.renderItemRGBA(item, 1)
	if got := canvas.RGBAAt(4, 4); got != (color.RGBA{10, 0, 0, 255}) {
		t.Errorf("stack-1 pixel = %v, want sprite 1 color {10 0 0 255}", got)
	}
}

func TestItemPNGCountOverlay(t *testing.T) {
	spr := emptySPR(8)
	for i := 1; i <= 8; i++ {
		spr.SetRGBA(uint32(i), solidTile(color.RGBA{0, 0, 128, 255}))
	}
	item := &DatItem{
		Width: 1, Height: 1, ColorLayers: 1,
		XDiv: 4, YDiv: 2, ZDiv: 1, AnimCount: 1,
		SpriteIDs: []uint32{1, 2, 3, 4, 5, 6, 7, 8},
	}
	c := &Cache{
		spr:     spr,
		dat:     &DatFile{Items: map[uint16]*DatItem{100: item}},
		png:     map[uint16][]byte{},
		itemPNG: map[uint32][]byte{},
	}

	data, err := c.ItemPNG(100, 90)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	found := false
	b := img.Bounds()
	for y := b.Max.Y - 14; y < b.Max.Y && !found; y++ {
		for x := 0; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			if r>>8 == 219 && g>>8 == 219 && bl>>8 == 219 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("count overlay color {219 219 219} not found in bottom rows")
	}
}

func TestItemPNGByServerIDTranslates(t *testing.T) {
	spr := emptySPR(1)
	spr.SetRGBA(1, solidTile(color.RGBA{7, 7, 7, 255}))
	item := &DatItem{
		Width: 1, Height: 1, ColorLayers: 1,
		XDiv: 1, YDiv: 1, ZDiv: 1, AnimCount: 1,
		SpriteIDs: []uint32{1},
	}
	clientPNG, _ := renderPNG(solidTile(color.RGBA{7, 7, 7, 255}))
	c := &Cache{
		spr:            spr,
		dat:            &DatFile{Items: map[uint16]*DatItem{3079: item}},
		png:            map[uint16][]byte{3079: clientPNG},
		itemPNG:        map[uint32][]byte{},
		serverToClient: map[uint16]uint16{2195: 3079},
	}

	data, err := c.ItemPNGByServerID(2195, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, clientPNG) {
		t.Error("ItemPNGByServerID(2195) did not resolve to client sprite 3079")
	}

	// Unknown server id falls back to a 1×1 transparent PNG, never errors.
	data, err = c.ItemPNGByServerID(9999, 1)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 1 || img.Bounds().Dy() != 1 {
		t.Errorf("unknown sid image = %v, want 1x1", img.Bounds())
	}
}
