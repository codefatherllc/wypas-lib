package sprite

import (
	"image"
	"image/color"
	"strconv"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// stackPattern maps an item count to its stack-sprite pattern index,
// matching the classic recount() tiers: 1, 2, 3, 4, 5+, 10+, 25+, 50+.
func stackPattern(count int) int {
	switch {
	case count >= 50:
		return 7
	case count >= 25:
		return 6
	case count >= 10:
		return 5
	case count >= 5:
		return 4
	case count >= 4:
		return 3
	case count >= 3:
		return 2
	case count >= 2:
		return 1
	default:
		return 0
	}
}

// renderItemRGBA composes the full Width×Height canvas for an item,
// selecting the stack-variant pattern for the given count.
func (c *Cache) renderItemRGBA(item *DatItem, count int) *image.RGBA {
	w := int(item.Width)
	h := int(item.Height)
	if w == 0 {
		w = 1
	}
	if h == 0 {
		h = 1
	}

	xPat, yPat := 0, 0
	if count > 1 {
		xd := int(item.XDiv)
		yd := int(item.YDiv)
		if xd == 0 {
			xd = 1
		}
		if yd == 0 {
			yd = 1
		}
		p := stackPattern(count)
		if p >= xd*yd {
			p = xd*yd - 1
		}
		xPat = p % xd
		yPat = (p / xd) % yd
	}

	canvas := image.NewRGBA(image.Rect(0, 0, w*32, h*32))
	for ht := 0; ht < h; ht++ {
		for wt := 0; wt < w; wt++ {
			idx := spriteIndex(0, 0, yPat, xPat, 0, ht, wt, item)
			if idx < 0 || idx >= len(item.SpriteIDs) {
				continue
			}
			sprID := item.SpriteIDs[idx]
			if sprID == 0 {
				continue
			}
			tile, err := c.spr.GetRGBA(sprID)
			if err != nil {
				continue
			}
			dx := (w - 1 - wt) * 32
			dy := (h - 1 - ht) * 32
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					px := tile.RGBAAt(x, y)
					if px.A == 0 {
						continue
					}
					canvas.SetRGBA(dx+x, dy+y, px)
				}
			}
		}
	}
	return canvas
}

// drawItemCount draws the stack count bottom-right, matching the classic
// GD font-3 style: near-black outline around a light-grey number.
func drawItemCount(canvas *image.RGBA, count int) {
	text := strconv.Itoa(count)
	face := basicfont.Face7x13
	w := canvas.Bounds().Dx()
	h := canvas.Bounds().Dy()
	x := w - (len(text)*face.Advance + 1)
	baseline := h - 13 + face.Ascent

	drawAt := func(dx, dy int, col color.Color) {
		d := &font.Drawer{
			Dst:  canvas,
			Src:  image.NewUniform(col),
			Face: face,
			Dot:  fixed.P(x+dx, baseline+dy),
		}
		d.DrawString(text)
	}

	outline := color.RGBA{R: 1, G: 1, B: 1, A: 255}
	drawAt(-1, -1, outline)
	drawAt(0, -1, outline)
	drawAt(-1, 0, outline)
	drawAt(0, 1, outline)
	drawAt(1, 0, outline)
	drawAt(1, 1, outline)
	drawAt(0, 0, color.RGBA{R: 219, G: 219, B: 219, A: 255})
}

// ItemPNG renders an item by client ID. Counts above 1 select the
// stack-variant sprite and draw the count number bottom-right.
// Count is clamped to [1, 100]. Unknown IDs yield a 1×1 transparent PNG.
func (c *Cache) ItemPNG(clientID uint16, count int) ([]byte, error) {
	if count < 1 {
		count = 1
	}
	if count > 100 {
		count = 100
	}

	if count == 1 {
		return c.PNG(clientID)
	}

	key := uint32(clientID)<<8 | uint32(count)
	c.mu.RLock()
	data, ok := c.itemPNG[key]
	c.mu.RUnlock()
	if ok {
		return data, nil
	}

	item, exists := c.dat.Items[clientID]
	if !exists {
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		return renderPNG(img)
	}

	canvas := c.renderItemRGBA(item, count)
	drawItemCount(canvas, count)
	data, err := renderPNG(canvas)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.itemPNG[key] = data
	c.mu.Unlock()
	return data, nil
}

// hasVisibleTile reports whether any layer-0 tile of the item's first
// pattern has a sprite, i.e. whether a composed render would be non-empty.
func hasVisibleTile(item *DatItem) bool {
	w := int(item.Width)
	h := int(item.Height)
	if w == 0 {
		w = 1
	}
	if h == 0 {
		h = 1
	}
	for ht := 0; ht < h; ht++ {
		for wt := 0; wt < w; wt++ {
			idx := spriteIndex(0, 0, 0, 0, 0, ht, wt, item)
			if idx >= 0 && idx < len(item.SpriteIDs) && item.SpriteIDs[idx] != 0 {
				return true
			}
		}
	}
	return false
}

// ItemPNGByServerID renders an item by server ID (items.otb sid),
// translating to the client ID first. See ItemPNG for count semantics.
func (c *Cache) ItemPNGByServerID(serverID uint16, count int) ([]byte, error) {
	clientID, ok := c.serverToClient[serverID]
	if !ok {
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		return renderPNG(img)
	}
	return c.ItemPNG(clientID, count)
}
