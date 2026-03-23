package gpu

import (
	"testing"
)

func backends(t *testing.T) map[string]Renderer {
	t.Helper()
	m := map[string]Renderer{
		"cpu": NewCPU(),
	}
	metal, err := newMetal()
	if err == nil {
		m["metal"] = metal
		t.Cleanup(metal.Close)
	}
	return m
}

func TestComposite(t *testing.T) {
	bg := make([]byte, 4*4*4)
	for i := 0; i < len(bg); i += 4 {
		bg[i+0] = 100
		bg[i+1] = 150
		bg[i+2] = 200
		bg[i+3] = 255
	}

	fg := make([]byte, 2*2*4)
	for i := 0; i < len(fg); i += 4 {
		fg[i+0] = 255
		fg[i+1] = 0
		fg[i+2] = 0
		fg[i+3] = 128
	}

	ops := []BlitOp{
		{SrcPixels: bg, SrcW: 4, SrcH: 4, DstX: 0, DstY: 0},
		{SrcPixels: fg, SrcW: 2, SrcH: 2, DstX: 1, DstY: 1},
	}

	for name, r := range backends(t) {
		t.Run(name, func(t *testing.T) {
			out, err := r.Composite(4, 4, ops)
			if err != nil {
				t.Fatal(err)
			}
			if len(out) != 4*4*4 {
				t.Fatalf("expected %d bytes, got %d", 4*4*4, len(out))
			}

			// pixel (0,0) should be pure bg
			off := 0
			if out[off] != 100 || out[off+1] != 150 || out[off+2] != 200 || out[off+3] != 255 {
				t.Errorf("(0,0) expected bg, got %v", out[off:off+4])
			}

			// pixel (1,1) should be blended (fg alpha=128 over bg alpha=255)
			off = (1*4 + 1) * 4
			sa := uint32(128)
			da := uint32(255)
			outA := sa + da*(255-sa)/255
			expR := uint8((255*sa + 100*da*(255-sa)/255) / outA)
			if abs8(out[off], expR) > 1 {
				t.Errorf("(1,1).R expected ~%d, got %d", expR, out[off])
			}

			// pixel (3,3) should be pure bg
			off = (3*4 + 3) * 4
			if out[off] != 100 || out[off+1] != 150 || out[off+2] != 200 {
				t.Errorf("(3,3) expected bg, got %v", out[off:off+4])
			}
		})
	}
}

func TestTintOutfit(t *testing.T) {
	const n = 32 * 32 * 4
	base := make([]byte, n)
	overlay := make([]byte, n)

	// pixel (0,0): base white, overlay yellow → head tint
	base[0], base[1], base[2], base[3] = 200, 180, 160, 255
	overlay[0], overlay[1], overlay[2], overlay[3] = 255, 255, 0, 255

	// pixel (1,0): base white, overlay red → body tint
	base[4], base[5], base[6], base[7] = 200, 180, 160, 255
	overlay[4], overlay[5], overlay[6], overlay[7] = 255, 0, 0, 255

	// pixel (2,0): base, overlay has non-matching color → no tint
	base[8], base[9], base[10], base[11] = 200, 180, 160, 255
	overlay[8], overlay[9], overlay[10], overlay[11] = 128, 128, 128, 255

	// pixel (3,0): transparent base → skip
	base[12], base[13], base[14], base[15] = 0, 0, 0, 0

	// pixel (4,0): magenta base → skip
	base[16], base[17], base[18], base[19] = 252, 0, 252, 255

	p := OutfitTintParams{
		Base:    base,
		Overlay: overlay,
		Head:    [3]uint8{255, 128, 64},
		Body:    [3]uint8{64, 128, 255},
		Legs:    [3]uint8{100, 200, 50},
		Feet:    [3]uint8{200, 100, 150},
	}

	for name, r := range backends(t) {
		t.Run(name, func(t *testing.T) {
			out, err := r.TintOutfit(p)
			if err != nil {
				t.Fatal(err)
			}

			// head tint: 200*255/255=200, 180*128/255=90, 160*64/255=40
			expR := mulCh(200, 255)
			expG := mulCh(180, 128)
			expB := mulCh(160, 64)
			if abs8(out[0], expR) > 1 || abs8(out[1], expG) > 1 || abs8(out[2], expB) > 1 {
				t.Errorf("head tint: expected ~(%d,%d,%d), got (%d,%d,%d)", expR, expG, expB, out[0], out[1], out[2])
			}

			// body tint
			expR = mulCh(200, 64)
			expG = mulCh(180, 128)
			expB = mulCh(160, 255)
			if abs8(out[4], expR) > 1 || abs8(out[5], expG) > 1 || abs8(out[6], expB) > 1 {
				t.Errorf("body tint: expected ~(%d,%d,%d), got (%d,%d,%d)", expR, expG, expB, out[4], out[5], out[6])
			}

			// no tint: base pass-through
			if out[8] != 200 || out[9] != 180 || out[10] != 160 {
				t.Errorf("no-tint: expected (200,180,160), got (%d,%d,%d)", out[8], out[9], out[10])
			}

			// transparent: all zero
			if out[12] != 0 || out[13] != 0 || out[14] != 0 || out[15] != 0 {
				t.Errorf("transparent: expected all 0, got %v", out[12:16])
			}

			// magenta: all zero
			if out[16] != 0 || out[17] != 0 || out[18] != 0 || out[19] != 0 {
				t.Errorf("magenta: expected all 0, got %v", out[16:20])
			}
		})
	}
}

func TestResizeNN(t *testing.T) {
	// 4x4 image with distinct quadrant colors
	src := make([]byte, 4*4*4)
	colors := [][4]byte{
		{255, 0, 0, 255},     // top-left red
		{0, 255, 0, 255},     // top-right green
		{0, 0, 255, 255},     // bottom-left blue
		{255, 255, 0, 255},   // bottom-right yellow
	}
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			qi := 0
			if x >= 2 {
				qi += 1
			}
			if y >= 2 {
				qi += 2
			}
			off := (y*4 + x) * 4
			copy(src[off:off+4], colors[qi][:])
		}
	}

	for name, r := range backends(t) {
		t.Run(name, func(t *testing.T) {
			out, err := r.ResizeNN(src, 4, 4, 2, 2)
			if err != nil {
				t.Fatal(err)
			}
			if len(out) != 2*2*4 {
				t.Fatalf("expected %d bytes, got %d", 2*2*4, len(out))
			}
			// (0,0) → src(0,0) = red
			if out[0] != 255 || out[1] != 0 || out[2] != 0 {
				t.Errorf("(0,0) expected red, got (%d,%d,%d)", out[0], out[1], out[2])
			}
			// (1,0) → src(2,0) = green
			if out[4] != 0 || out[5] != 255 || out[6] != 0 {
				t.Errorf("(1,0) expected green, got (%d,%d,%d)", out[4], out[5], out[6])
			}
			// (0,1) → src(0,2) = blue
			if out[8] != 0 || out[9] != 0 || out[10] != 255 {
				t.Errorf("(0,1) expected blue, got (%d,%d,%d)", out[8], out[9], out[10])
			}
			// (1,1) → src(2,2) = yellow
			if out[12] != 255 || out[13] != 255 || out[14] != 0 {
				t.Errorf("(1,1) expected yellow, got (%d,%d,%d)", out[12], out[13], out[14])
			}
		})
	}
}

func TestResizeCatmullRom(t *testing.T) {
	// 4x4 solid red → 8x8 should still be mostly red
	src := make([]byte, 4*4*4)
	for i := 0; i < len(src); i += 4 {
		src[i], src[i+1], src[i+2], src[i+3] = 255, 0, 0, 255
	}

	for name, r := range backends(t) {
		t.Run(name, func(t *testing.T) {
			out, err := r.ResizeCatmullRom(src, 4, 4, 8, 8)
			if err != nil {
				t.Fatal(err)
			}
			if len(out) != 8*8*4 {
				t.Fatalf("expected %d bytes, got %d", 8*8*4, len(out))
			}
			// center pixel should be red
			off := (4*8 + 4) * 4
			if out[off] < 200 {
				t.Errorf("center R expected >200, got %d", out[off])
			}
		})
	}
}

func TestRenderHeatmap(t *testing.T) {
	freq := map[[2]int]int{
		{5, 5}: 100,
	}
	p := HeatmapParams{
		Freq:   freq,
		ImgW:   10,
		ImgH:   10,
		Radius: 3,
	}

	for name, r := range backends(t) {
		t.Run(name, func(t *testing.T) {
			out, err := r.RenderHeatmap(p)
			if err != nil {
				t.Fatal(err)
			}
			if len(out) != 10*10*4 {
				t.Fatalf("expected %d bytes, got %d", 10*10*4, len(out))
			}

			// center pixel (5,5) should have nonzero alpha
			off := (5*10 + 5) * 4
			if out[off+3] == 0 {
				t.Error("center pixel should have nonzero alpha")
			}

			// far corner (0,0) should be transparent (outside radius)
			if out[3] != 0 {
				t.Errorf("corner pixel expected transparent, got alpha %d", out[3])
			}
		})
	}
}

func TestApplyShadow(t *testing.T) {
	pixels := make([]byte, 2*2*4)
	for i := 0; i < len(pixels); i += 4 {
		pixels[i], pixels[i+1], pixels[i+2], pixels[i+3] = 200, 100, 50, 255
	}

	for name, r := range backends(t) {
		t.Run(name, func(t *testing.T) {
			out, err := r.ApplyShadow(pixels, 2, 2, 0.5)
			if err != nil {
				t.Fatal(err)
			}

			// 200 * 0.5 = 100, 100 * 0.5 = 50, 50 * 0.5 = 25
			if abs8(out[0], 100) > 1 {
				t.Errorf("R expected ~100, got %d", out[0])
			}
			if abs8(out[1], 50) > 1 {
				t.Errorf("G expected ~50, got %d", out[1])
			}
			if abs8(out[2], 25) > 1 {
				t.Errorf("B expected ~25, got %d", out[2])
			}
			if out[3] != 255 {
				t.Errorf("A expected 255, got %d", out[3])
			}
		})
	}
}

func TestRenderMinimap(t *testing.T) {
	tiles := []MinimapTile{
		{X: 0, Y: 0, ColorIndex: 0},
		{X: 1, Y: 0, ColorIndex: 215}, // 5*36 + 5*6 + 5 = 215 → (255,255,255)
		{X: 2, Y: 0, ColorIndex: 36},  // 1*36 → (51,0,0)
	}

	for name, r := range backends(t) {
		t.Run(name, func(t *testing.T) {
			out, err := r.RenderMinimap(4, 2, tiles)
			if err != nil {
				t.Fatal(err)
			}
			if len(out) != 4*2*4 {
				t.Fatalf("expected %d bytes, got %d", 4*2*4, len(out))
			}

			// colorIndex 0: (0,0,0,255)
			if out[0] != 0 || out[1] != 0 || out[2] != 0 || out[3] != 255 {
				t.Errorf("tile 0: expected (0,0,0,255), got %v", out[0:4])
			}

			// colorIndex 215: 5*51=255 each
			off := 4
			if out[off] != 255 || out[off+1] != 255 || out[off+2] != 255 || out[off+3] != 255 {
				t.Errorf("tile 215: expected (255,255,255,255), got %v", out[off:off+4])
			}

			// colorIndex 36: r=51, g=0, b=0
			off = 8
			if out[off] != 51 || out[off+1] != 0 || out[off+2] != 0 || out[off+3] != 255 {
				t.Errorf("tile 36: expected (51,0,0,255), got %v", out[off:off+4])
			}

			// empty tile (3,0): all zero
			off = 12
			if out[off] != 0 || out[off+1] != 0 || out[off+2] != 0 || out[off+3] != 0 {
				t.Errorf("empty tile: expected all 0, got %v", out[off:off+4])
			}
		})
	}
}

func abs8(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}
