package sprite

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

const (
	hsiHSteps   = 19
	hsiSIValues = 7
)

var OutfitPalette [hsiHSteps * hsiSIValues][3]uint8

func init() {
	for i := 0; i < hsiHSteps*hsiSIValues; i++ {
		OutfitPalette[i] = computeOutfitColor(i)
	}
}

func OutfitPaletteRGB() [][3]uint8 {
	out := make([][3]uint8, len(OutfitPalette))
	copy(out, OutfitPalette[:])
	return out
}

func computeOutfitColor(c int) [3]uint8 {
	if c >= hsiHSteps*hsiSIValues {
		c = 0
	}

	var loc1, loc2, loc3 float64

	if c%hsiHSteps != 0 {
		loc1 = float64(c%hsiHSteps) / 18.0
		loc2 = 1
		loc3 = 1
		switch c / hsiHSteps {
		case 0:
			loc2 = 0.25
			loc3 = 1.00
		case 1:
			loc2 = 0.25
			loc3 = 0.75
		case 2:
			loc2 = 0.50
			loc3 = 0.75
		case 3:
			loc2 = 0.667
			loc3 = 0.75
		case 4:
			loc2 = 1.00
			loc3 = 1.00
		case 5:
			loc2 = 1.00
			loc3 = 0.75
		case 6:
			loc2 = 1.00
			loc3 = 0.50
		}
	} else {
		loc3 = 1.0 - float64(c)/float64(hsiHSteps)/float64(hsiSIValues)
	}

	if loc3 == 0 {
		return [3]uint8{0, 0, 0}
	}
	if loc2 == 0 {
		v := uint8(math.Floor(loc3 * 255))
		return [3]uint8{v, v, v}
	}

	var red, green, blue float64

	switch {
	case loc1 < 1.0/6.0:
		red = loc3
		blue = loc3 * (1 - loc2)
		green = blue + (loc3-blue)*6*loc1
	case loc1 < 2.0/6.0:
		green = loc3
		blue = loc3 * (1 - loc2)
		red = green - (loc3-blue)*(6*loc1-1)
	case loc1 < 3.0/6.0:
		green = loc3
		red = loc3 * (1 - loc2)
		blue = red + (loc3-red)*(6*loc1-2)
	case loc1 < 4.0/6.0:
		blue = loc3
		red = loc3 * (1 - loc2)
		green = blue - (loc3-red)*(6*loc1-3)
	case loc1 < 5.0/6.0:
		blue = loc3
		green = loc3 * (1 - loc2)
		red = green + (loc3-green)*(6*loc1-4)
	default:
		red = loc3
		green = loc3 * (1 - loc2)
		blue = red - (loc3-green)*(6*loc1-5)
	}

	return [3]uint8{
		uint8(math.Floor(red * 255)),
		uint8(math.Floor(green * 255)),
		uint8(math.Floor(blue * 255)),
	}
}

func clampColor(v int) int {
	if v < 0 {
		return 0
	}
	if v > 132 {
		return 132
	}
	return v
}

func multiplyChannel(base uint8, pal uint8) uint8 {
	return uint8(uint16(base) * uint16(pal) / 255)
}

func spriteIndex(anim, z, y, x, l, h, w int, item *DatItem) int {
	ac := int(item.AnimCount)
	if ac == 0 {
		ac = 1
	}
	return ((((((anim%ac)*int(item.ZDiv)+z)*int(item.YDiv)+y)*int(item.XDiv)+x)*int(item.ColorLayers)+l)*int(item.Height)+h)*int(item.Width) + w
}

func (c *Cache) OutfitPNG(looktype uint16, head, body, legs, feet, addons, direction int, mount uint16) ([]byte, error) {
	head = clampColor(head)
	body = clampColor(body)
	legs = clampColor(legs)
	feet = clampColor(feet)

	outfit, ok := c.dat.Outfits[looktype]
	if !ok {
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		return renderPNG(img)
	}

	headRGB := OutfitPalette[head]
	bodyRGB := OutfitPalette[body]
	legsRGB := OutfitPalette[legs]
	feetRGB := OutfitPalette[feet]

	if direction < 0 || direction > 3 {
		direction = 2
	}

	canvasW := int(outfit.Width) * 32
	canvasH := int(outfit.Height) * 32

	var mountItem *DatItem
	if mount > 0 {
		if mi, ok := c.dat.Outfits[mount]; ok {
			mountItem = mi
			mw := int(mi.Width) * 32
			mh := int(mi.Height) * 32
			if mw > canvasW {
				canvasW = mw
			}
			if mh > canvasH {
				canvasH = mh
			}
		}
	}

	canvas := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))

	if mountItem != nil {
		if err := renderOutfitLayer(c, canvas, mountItem, canvasW, canvasH, 0, direction, 0, headRGB, bodyRGB, legsRGB, feetRGB, false); err != nil {
			return nil, fmt.Errorf("render mount: %w", err)
		}
		zPattern := 0
		if int(outfit.ZDiv) > 1 {
			zPattern = 1
		}
		if err := renderOutfitLayer(c, canvas, outfit, canvasW, canvasH, 0, direction, zPattern, headRGB, bodyRGB, legsRGB, feetRGB, false); err != nil {
			return nil, fmt.Errorf("render base outfit: %w", err)
		}
		if addons&1 != 0 && int(outfit.YDiv) > 1 {
			if err := renderOutfitLayer(c, canvas, outfit, canvasW, canvasH, 1, direction, zPattern, headRGB, bodyRGB, legsRGB, feetRGB, true); err != nil {
				return nil, fmt.Errorf("render addon1: %w", err)
			}
		}
		if addons&2 != 0 && int(outfit.YDiv) > 2 {
			if err := renderOutfitLayer(c, canvas, outfit, canvasW, canvasH, 2, direction, zPattern, headRGB, bodyRGB, legsRGB, feetRGB, true); err != nil {
				return nil, fmt.Errorf("render addon2: %w", err)
			}
		}
	} else {
		if err := renderOutfitLayer(c, canvas, outfit, canvasW, canvasH, 0, direction, 0, headRGB, bodyRGB, legsRGB, feetRGB, false); err != nil {
			return nil, fmt.Errorf("render base outfit: %w", err)
		}
		if addons&1 != 0 && int(outfit.YDiv) > 1 {
			if err := renderOutfitLayer(c, canvas, outfit, canvasW, canvasH, 1, direction, 0, headRGB, bodyRGB, legsRGB, feetRGB, true); err != nil {
				return nil, fmt.Errorf("render addon1: %w", err)
			}
		}
		if addons&2 != 0 && int(outfit.YDiv) > 2 {
			if err := renderOutfitLayer(c, canvas, outfit, canvasW, canvasH, 2, direction, 0, headRGB, bodyRGB, legsRGB, feetRGB, true); err != nil {
				return nil, fmt.Errorf("render addon2: %w", err)
			}
		}
	}

	return renderPNG(autocropRGBA(canvas))
}

func renderOutfitLayer(
	c *Cache,
	canvas *image.RGBA,
	item *DatItem,
	canvasW, canvasH int,
	ydiv int,
	dir int,
	zPattern int,
	headRGB, bodyRGB, legsRGB, feetRGB [3]uint8,
	isAddon bool,
) error {
	w := int(item.Width)
	h := int(item.Height)
	cl := int(item.ColorLayers)

	offsetX := canvasW - w*32
	offsetY := canvasH - h*32

	for htile := 0; htile < h; htile++ {
		for wtile := 0; wtile < w; wtile++ {
			canvasX := offsetX + (w-1-wtile)*32
			canvasY := offsetY + (h-1-htile)*32

			baseIdx := spriteIndex(0, zPattern, ydiv, dir, 0, htile, wtile, item)
			if baseIdx < 0 || baseIdx >= len(item.SpriteIDs) {
				continue
			}
			baseSprID := item.SpriteIDs[baseIdx]

			var overlayImg *image.RGBA
			if cl >= 2 {
				overlayIdx := spriteIndex(0, zPattern, ydiv, dir, 1, htile, wtile, item)
				if overlayIdx >= 0 && overlayIdx < len(item.SpriteIDs) {
					overlaySprID := item.SpriteIDs[overlayIdx]
					if overlaySprID != 0 {
						img, err := c.spr.GetRGBA(overlaySprID)
						if err == nil {
							overlayImg = img
						}
					}
				}
			}

			if baseSprID == 0 {
				continue
			}

			baseImg, err := c.spr.GetRGBA(baseSprID)
			if err != nil {
				continue
			}

			compositeTile(canvas, baseImg, overlayImg, canvasX, canvasY, headRGB, bodyRGB, legsRGB, feetRGB, isAddon)
		}
	}

	return nil
}

func compositeTile(
	canvas *image.RGBA,
	baseImg *image.RGBA,
	overlayImg *image.RGBA,
	cx, cy int,
	headRGB, bodyRGB, legsRGB, feetRGB [3]uint8,
	isAddon bool,
) {
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			bp := baseImg.RGBAAt(x, y)

			if bp.A == 0 {
				continue
			}

			if bp.R == 252 && bp.G == 0 && bp.B == 252 {
				continue
			}

			var out color.RGBA

			if overlayImg != nil && !isAddon {
				op := overlayImg.RGBAAt(x, y)

				switch {
				case op.R == 255 && op.G == 255 && op.B == 0:
					out = color.RGBA{
						R: multiplyChannel(bp.R, headRGB[0]),
						G: multiplyChannel(bp.G, headRGB[1]),
						B: multiplyChannel(bp.B, headRGB[2]),
						A: 255,
					}
				case op.R == 255 && op.G == 0 && op.B == 0:
					out = color.RGBA{
						R: multiplyChannel(bp.R, bodyRGB[0]),
						G: multiplyChannel(bp.G, bodyRGB[1]),
						B: multiplyChannel(bp.B, bodyRGB[2]),
						A: 255,
					}
				case op.R == 0 && op.G == 255 && op.B == 0:
					out = color.RGBA{
						R: multiplyChannel(bp.R, legsRGB[0]),
						G: multiplyChannel(bp.G, legsRGB[1]),
						B: multiplyChannel(bp.B, legsRGB[2]),
						A: 255,
					}
				case op.R == 0 && op.G == 0 && op.B == 255:
					out = color.RGBA{
						R: multiplyChannel(bp.R, feetRGB[0]),
						G: multiplyChannel(bp.G, feetRGB[1]),
						B: multiplyChannel(bp.B, feetRGB[2]),
						A: 255,
					}
				default:
					out = color.RGBA{R: bp.R, G: bp.G, B: bp.B, A: 255}
				}
			} else if overlayImg != nil && isAddon {
				out = color.RGBA{R: bp.R, G: bp.G, B: bp.B, A: 255}
				canvas.SetRGBA(cx+x, cy+y, out)

				op := overlayImg.RGBAAt(x, y)
				switch {
				case op.R == 255 && op.G == 255 && op.B == 0:
					out = color.RGBA{
						R: multiplyChannel(bp.R, headRGB[0]),
						G: multiplyChannel(bp.G, headRGB[1]),
						B: multiplyChannel(bp.B, headRGB[2]),
						A: 255,
					}
					canvas.SetRGBA(cx+x, cy+y, out)
				case op.R == 255 && op.G == 0 && op.B == 0:
					out = color.RGBA{
						R: multiplyChannel(bp.R, bodyRGB[0]),
						G: multiplyChannel(bp.G, bodyRGB[1]),
						B: multiplyChannel(bp.B, bodyRGB[2]),
						A: 255,
					}
					canvas.SetRGBA(cx+x, cy+y, out)
				case op.R == 0 && op.G == 255 && op.B == 0:
					out = color.RGBA{
						R: multiplyChannel(bp.R, legsRGB[0]),
						G: multiplyChannel(bp.G, legsRGB[1]),
						B: multiplyChannel(bp.B, legsRGB[2]),
						A: 255,
					}
					canvas.SetRGBA(cx+x, cy+y, out)
				case op.R == 0 && op.G == 0 && op.B == 255:
					out = color.RGBA{
						R: multiplyChannel(bp.R, feetRGB[0]),
						G: multiplyChannel(bp.G, feetRGB[1]),
						B: multiplyChannel(bp.B, feetRGB[2]),
						A: 255,
					}
					canvas.SetRGBA(cx+x, cy+y, out)
				}
				continue
			} else {
				out = color.RGBA{R: bp.R, G: bp.G, B: bp.B, A: 255}
			}

			canvas.SetRGBA(cx+x, cy+y, out)
		}
	}
}

func autocropRGBA(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if minX > maxX || minY > maxY {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	cropped := image.NewRGBA(image.Rect(0, 0, maxX-minX+1, maxY-minY+1))
	for y := minY; y <= maxY; y++ {
		copy(cropped.Pix[(y-minY)*cropped.Stride:], img.Pix[y*img.Stride+minX*4:(y*img.Stride+(maxX+1)*4)])
	}
	return cropped
}
