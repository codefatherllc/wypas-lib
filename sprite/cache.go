package sprite

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"sync"
)

type Cache struct {
	dat            *DatFile
	spr            *SpriteFile
	mu             sync.RWMutex
	png            map[uint16][]byte
	pngByServerID  map[uint16][]byte
}

func NewCache(datPath, sprPath string, serverToClient map[uint16]uint16) (*Cache, error) {
	dat, err := parseDAT(datPath)
	if err != nil {
		return nil, fmt.Errorf("parse dat: %w", err)
	}

	spr, err := parseSPR(sprPath)
	if err != nil {
		return nil, fmt.Errorf("parse spr: %w", err)
	}

	c := &Cache{
		dat: dat,
		spr: spr,
		png: make(map[uint16][]byte, len(dat.Items)),
	}

	for id, item := range dat.Items {
		if len(item.SpriteIDs) == 0 || item.SpriteIDs[0] == 0 {
			continue
		}

		img, err := spr.GetRGBA(uint32(item.SpriteIDs[0]))
		if err != nil {
			log.Printf("sprite cache: skip item %d: %v", id, err)
			continue
		}

		data, err := renderPNG(img)
		if err != nil {
			log.Printf("sprite cache: encode item %d: %v", id, err)
			continue
		}

		c.png[id] = data
	}

	c.pngByServerID = make(map[uint16][]byte, len(serverToClient))
	for serverID, clientID := range serverToClient {
		if data, ok := c.png[clientID]; ok {
			c.pngByServerID[serverID] = data
		}
	}

	log.Printf("sprite cache: loaded %d client sprites, %d server sprites", len(c.png), len(c.pngByServerID))
	return c, nil
}

func (c *Cache) PNG(id uint16) ([]byte, error) {
	c.mu.RLock()
	data, ok := c.png[id]
	c.mu.RUnlock()

	if ok {
		return data, nil
	}

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{})
	return renderPNG(img)
}

func (c *Cache) PNGByServerID(id uint16) ([]byte, error) {
	c.mu.RLock()
	data, ok := c.pngByServerID[id]
	c.mu.RUnlock()

	if ok {
		return data, nil
	}

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{})
	return renderPNG(img)
}

func (c *Cache) EffectPNG(id uint16, frame int) ([]byte, error) {
	effect, ok := c.dat.Effects[id]
	if !ok || len(effect.SpriteIDs) == 0 {
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		return renderPNG(img)
	}

	w := int(effect.Width)
	h := int(effect.Height)
	ac := int(effect.AnimCount)
	if ac == 0 { ac = 1 }
	frame = frame % ac

	canvasW := w * 32
	canvasH := h * 32
	canvas := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))

	for ht := 0; ht < h; ht++ {
		for wt := 0; wt < w; wt++ {
			idx := spriteIndex(frame, 0, 0, 0, 0, ht, wt, effect)
			if idx >= len(effect.SpriteIDs) {
				continue
			}
			sprID := effect.SpriteIDs[idx]
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
					if px.A > 0 {
						canvas.SetRGBA(dx+x, dy+y, px)
					}
				}
			}
		}
	}

	return renderPNG(canvas)
}

func (c *Cache) MissilePNG(id uint16, direction int) ([]byte, error) {
	missile, ok := c.dat.Missiles[id]
	if !ok || len(missile.SpriteIDs) == 0 {
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		return renderPNG(img)
	}

	w := int(missile.Width)
	h := int(missile.Height)
	xd := int(missile.XDiv)
	yd := int(missile.YDiv)
	if xd == 0 { xd = 1 }
	if yd == 0 { yd = 1 }
	totalDirs := xd * yd
	direction = direction % totalDirs
	xPat := direction % xd
	yPat := direction / xd

	canvasW := w * 32
	canvasH := h * 32
	canvas := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))

	for ht := 0; ht < h; ht++ {
		for wt := 0; wt < w; wt++ {
			idx := spriteIndex(0, 0, yPat, xPat, 0, ht, wt, missile)
			if idx >= len(missile.SpriteIDs) {
				continue
			}
			sprID := missile.SpriteIDs[idx]
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
					if px.A > 0 {
						canvas.SetRGBA(dx+x, dy+y, px)
					}
				}
			}
		}
	}

	return renderPNG(canvas)
}

func (c *Cache) Dat() *DatFile    { return c.dat }
func (c *Cache) Spr() *SpriteFile { return c.spr }
