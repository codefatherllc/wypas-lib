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

func (c *Cache) Dat() *DatFile    { return c.dat }
func (c *Cache) Spr() *SpriteFile { return c.spr }
