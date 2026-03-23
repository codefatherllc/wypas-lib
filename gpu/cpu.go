package gpu

import (
	"image"
	"math"

	xdraw "golang.org/x/image/draw"
)

type cpuRenderer struct{}

func NewCPU() Renderer { return &cpuRenderer{} }

func (c *cpuRenderer) Close() {}

func (c *cpuRenderer) Composite(dstW, dstH int, ops []BlitOp) ([]byte, error) {
	n := dstW * dstH * 4
	dst := GetPix(n)
	for i := range dst {
		dst[i] = 0
	}

	for _, op := range ops {
		for sy := 0; sy < op.SrcH; sy++ {
			dy := op.DstY + sy
			if dy < 0 || dy >= dstH {
				continue
			}
			for sx := 0; sx < op.SrcW; sx++ {
				dx := op.DstX + sx
				if dx < 0 || dx >= dstW {
					continue
				}
				si := (sy*op.SrcW + sx) * 4
				di := (dy*dstW + dx) * 4

				sa := uint32(op.SrcPixels[si+3])
				if sa == 0 {
					continue
				}

				if sa == 255 {
					copy(dst[di:di+4], op.SrcPixels[si:si+4])
					continue
				}

				da := uint32(dst[di+3])
				outA := sa + da*(255-sa)/255
				if outA == 0 {
					continue
				}

				dst[di+0] = uint8((uint32(op.SrcPixels[si+0])*sa + uint32(dst[di+0])*da*(255-sa)/255) / outA)
				dst[di+1] = uint8((uint32(op.SrcPixels[si+1])*sa + uint32(dst[di+1])*da*(255-sa)/255) / outA)
				dst[di+2] = uint8((uint32(op.SrcPixels[si+2])*sa + uint32(dst[di+2])*da*(255-sa)/255) / outA)
				dst[di+3] = uint8(outA)
			}
		}
	}

	return dst, nil
}

func (c *cpuRenderer) TintOutfit(p OutfitTintParams) ([]byte, error) {
	const size = 32
	n := size * size * 4
	out := GetPix(n)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			i := (y*size + x) * 4
			br, bg, bb, ba := p.Base[i], p.Base[i+1], p.Base[i+2], p.Base[i+3]

			if ba == 0 {
				out[i], out[i+1], out[i+2], out[i+3] = 0, 0, 0, 0
				continue
			}

			if br == 252 && bg == 0 && bb == 252 {
				out[i], out[i+1], out[i+2], out[i+3] = 0, 0, 0, 0
				continue
			}

			if p.Overlay == nil {
				out[i], out[i+1], out[i+2], out[i+3] = br, bg, bb, ba
				continue
			}

			or, og, ob := p.Overlay[i], p.Overlay[i+1], p.Overlay[i+2]
			var pal [3]uint8

			matched := true
			switch {
			case or == 255 && og == 255 && ob == 0:
				pal = p.Head
			case or == 255 && og == 0 && ob == 0:
				pal = p.Body
			case or == 0 && og == 255 && ob == 0:
				pal = p.Legs
			case or == 0 && og == 0 && ob == 255:
				pal = p.Feet
			default:
				matched = false
			}

			if p.IsAddon {
				out[i], out[i+1], out[i+2], out[i+3] = br, bg, bb, 255
				if matched {
					out[i+0] = mulCh(br, pal[0])
					out[i+1] = mulCh(bg, pal[1])
					out[i+2] = mulCh(bb, pal[2])
				}
			} else {
				if matched {
					out[i+0] = mulCh(br, pal[0])
					out[i+1] = mulCh(bg, pal[1])
					out[i+2] = mulCh(bb, pal[2])
					out[i+3] = 255
				} else {
					out[i], out[i+1], out[i+2], out[i+3] = br, bg, bb, 255
				}
			}
		}
	}

	return out, nil
}

func mulCh(base, pal uint8) uint8 {
	return uint8(uint16(base) * uint16(pal) / 255)
}

func (c *cpuRenderer) ResizeNN(src []byte, srcW, srcH, dstW, dstH int) ([]byte, error) {
	n := dstW * dstH * 4
	dst := GetPix(n)

	for y := 0; y < dstH; y++ {
		sy := y * srcH / dstH
		for x := 0; x < dstW; x++ {
			sx := x * srcW / dstW
			si := (sy*srcW + sx) * 4
			di := (y*dstW + x) * 4
			copy(dst[di:di+4], src[si:si+4])
		}
	}

	return dst, nil
}

func (c *cpuRenderer) ResizeCatmullRom(src []byte, srcW, srcH, dstW, dstH int) ([]byte, error) {
	srcImg := &image.RGBA{
		Pix:    src,
		Stride: srcW * 4,
		Rect:   image.Rect(0, 0, srcW, srcH),
	}
	dstImg := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	xdraw.CatmullRom.Scale(dstImg, dstImg.Bounds(), srcImg, srcImg.Bounds(), xdraw.Over, nil)

	return dstImg.Pix, nil
}

func (c *cpuRenderer) RenderHeatmap(p HeatmapParams) ([]byte, error) {
	n := p.ImgW * p.ImgH * 4
	out := GetPix(n)

	if len(p.Freq) == 0 {
		return out, nil
	}

	invR2 := 1.0 / float64(p.Radius*p.Radius)
	heat := make([]float64, p.ImgW*p.ImgH)

	var maxVal float64
	for coord, count := range p.Freq {
		cx, cy := coord[0], coord[1]
		fc := float64(count)
		y0 := cy - p.Radius
		if y0 < 0 {
			y0 = 0
		}
		y1 := cy + p.Radius
		if y1 >= p.ImgH {
			y1 = p.ImgH - 1
		}
		x0 := cx - p.Radius
		if x0 < 0 {
			x0 = 0
		}
		x1 := cx + p.Radius
		if x1 >= p.ImgW {
			x1 = p.ImgW - 1
		}
		for py := y0; py <= y1; py++ {
			dy := float64(py - cy)
			for px := x0; px <= x1; px++ {
				dx := float64(px - cx)
				d2 := dx*dx + dy*dy
				r2 := float64(p.Radius * p.Radius)
				if d2 > r2 {
					continue
				}
				w := math.Exp(-3.0 * d2 * invR2)
				idx := py*p.ImgW + px
				heat[idx] += fc * w
				if heat[idx] > maxVal {
					maxVal = heat[idx]
				}
			}
		}
	}

	if maxVal <= 0 {
		return out, nil
	}

	type stop struct {
		pos     float64
		r, g, b uint8
	}
	stops := [4]stop{
		{0.0, 0, 80, 120},
		{0.33, 0, 255, 255},
		{0.66, 255, 0, 0},
		{1.0, 255, 255, 0},
	}

	for i := 0; i < p.ImgW*p.ImgH; i++ {
		t := heat[i] / maxVal
		if t <= 0 {
			continue
		}
		if t > 1 {
			t = 1
		}

		lo, hi := 0, 1
		for j := 1; j < len(stops); j++ {
			if t <= stops[j].pos {
				lo = j - 1
				hi = j
				break
			}
		}
		f := (t - stops[lo].pos) / (stops[hi].pos - stops[lo].pos)
		lerp := func(a, b uint8) uint8 {
			return uint8(float64(a) + f*(float64(b)-float64(a)))
		}
		alpha := uint8(160 + int(95*t))

		off := i * 4
		out[off+0] = lerp(stops[lo].r, stops[hi].r)
		out[off+1] = lerp(stops[lo].g, stops[hi].g)
		out[off+2] = lerp(stops[lo].b, stops[hi].b)
		out[off+3] = alpha
	}

	return out, nil
}

func (c *cpuRenderer) ApplyShadow(pixels []byte, w, h int, intensity float64) ([]byte, error) {
	n := w * h * 4
	out := GetPix(n)
	copy(out, pixels)

	mult := 1.0 - intensity
	for i := 0; i < n; i += 4 {
		if out[i+3] == 0 {
			continue
		}
		out[i+0] = uint8(float64(out[i+0]) * mult)
		out[i+1] = uint8(float64(out[i+1]) * mult)
		out[i+2] = uint8(float64(out[i+2]) * mult)
	}

	return out, nil
}

func (c *cpuRenderer) RenderMinimap(w, h int, tiles []MinimapTile) ([]byte, error) {
	n := w * h * 4
	dst := GetPix(n)
	for i := range dst {
		dst[i] = 0
	}

	for _, t := range tiles {
		if t.X < 0 || t.X >= w || t.Y < 0 || t.Y >= h {
			continue
		}
		ci := int(t.ColorIndex)
		r := uint8((ci / 36 % 6) * 51)
		g := uint8((ci / 6 % 6) * 51)
		b := uint8((ci % 6) * 51)

		off := (t.Y*w + t.X) * 4
		dst[off+0] = r
		dst[off+1] = g
		dst[off+2] = b
		dst[off+3] = 255
	}

	return dst, nil
}

var _ Renderer = (*cpuRenderer)(nil)
