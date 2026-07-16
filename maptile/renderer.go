package maptile

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"github.com/codefatherllc/wypas-lib/v2/gpu"
	"github.com/codefatherllc/wypas-lib/v2/otbm"
	"github.com/codefatherllc/wypas-lib/v2/sprite"
)

const TileSize = 256

// SpritePixels is the native sprite resolution in the SPR file.
const SpritePixels = 32

var pngEncoder = &png.Encoder{CompressionLevel: png.BestSpeed}

// Renderer produces 256x256 map tile PNGs.
// All methods are thread-safe (read-only access to shared data).
type Renderer struct {
	gameMap *otbm.GameMap
	otb     *otbm.OTB
	dat     *sprite.DatFile
	spr     *sprite.SpriteFile
	gpu     gpu.Renderer
}

func NewRenderer(gameMap *otbm.GameMap, otb *otbm.OTB, dat *sprite.DatFile, spr *sprite.SpriteFile, g gpu.Renderer) *Renderer {
	return &Renderer{
		gameMap: gameMap,
		otb:     otb,
		dat:     dat,
		spr:     spr,
		gpu:     g,
	}
}

// SpriteZoomThreshold is the zoom level at which sprite rendering begins.
// Below this, minimap colors (single floor, no stacking).
// At and above, actual sprites with floor stacking.
const SpriteZoomThreshold = 3

// MaxTileZoom is the highest zoom level for which we generate tiles server-side.
const MaxTileZoom = 5

// GameTilesForZoom returns how many game tiles one 256px tile covers at the given zoom.
func GameTilesForZoom(zoom int) int {
	if zoom >= 8 {
		return 1
	}
	if zoom >= 0 {
		return TileSize >> uint(zoom)
	}
	return TileSize << uint(-zoom)
}

// GameMap returns the underlying parsed map for metadata access.
func (r *Renderer) GameMap() *otbm.GameMap {
	return r.gameMap
}

// RenderTile renders a 256x256 PNG tile at the given floor, zoom, and tile coordinates.
// Returns nil if the tile has no data (empty region).
func (r *Renderer) RenderTile(floor int, zoom, x, y int) (data []byte, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("panic rendering tile f%d/z%d/%d_%d: %v", floor, zoom, x, y, rec)
			data = nil
		}
	}()

	if zoom > MaxTileZoom {
		return nil, nil
	}

	fb := r.gameMap.FloorBounds[uint8(floor)]
	if fb == nil {
		return nil, nil
	}

	gameTilesPerTile := GameTilesForZoom(zoom)

	startGameX := x * gameTilesPerTile
	startGameY := y * gameTilesPerTile
	endGameX := startGameX + gameTilesPerTile
	endGameY := startGameY + gameTilesPerTile

	if startGameX > int(fb.MaxX) || endGameX <= int(fb.MinX) ||
		startGameY > int(fb.MaxY) || endGameY <= int(fb.MinY) {
		return nil, nil
	}

	if zoom < SpriteZoomThreshold {
		return r.renderMinimap(floor, zoom, x, y)
	}
	return r.renderSprite(floor, zoom, x, y)
}

// RenderMinimapImage renders a minimap-colored image for the given tile data region.
// Useful for services that need an image.RGBA directly instead of PNG bytes.
func (r *Renderer) RenderMinimapImage(floor int, startX, startY, width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	z := uint8(floor)

	for ty := 0; ty < height; ty++ {
		gy := startY + ty
		for tx := 0; tx < width; tx++ {
			gx := startX + tx
			if gx < 0 || gy < 0 || gx > 0xFFFF || gy > 0xFFFF {
				continue
			}
			c, ok := r.minimapColor(uint16(gx), uint16(gy), z)
			if !ok {
				continue
			}
			img.SetRGBA(tx, ty, c)
		}
	}
	return img
}

func (r *Renderer) renderMinimap(floor int, zoom, x, y int) ([]byte, error) {
	gameTilesPerTile := GameTilesForZoom(zoom)

	startGameX := x * gameTilesPerTile
	startGameY := y * gameTilesPerTile

	img := image.NewRGBA(image.Rect(0, 0, TileSize, TileSize))
	hasData := false

	z := uint8(floor)

	if gameTilesPerTile <= TileSize {
		pixelsPerGameTile := TileSize / gameTilesPerTile

		for ty := 0; ty < gameTilesPerTile; ty++ {
			gy := startGameY + ty
			for tx := 0; tx < gameTilesPerTile; tx++ {
				gx := startGameX + tx

				c, ok := r.minimapColor(uint16(gx), uint16(gy), z)
				if !ok {
					continue
				}
				hasData = true
				px := tx * pixelsPerGameTile
				py := ty * pixelsPerGameTile
				for dy := 0; dy < pixelsPerGameTile; dy++ {
					for dx := 0; dx < pixelsPerGameTile; dx++ {
						img.SetRGBA(px+dx, py+dy, c)
					}
				}
			}
		}
	} else {
		for py := 0; py < TileSize; py++ {
			gy := startGameY + py*gameTilesPerTile/TileSize
			for px := 0; px < TileSize; px++ {
				gx := startGameX + px*gameTilesPerTile/TileSize

				c, ok := r.minimapColor(uint16(gx), uint16(gy), z)
				if !ok {
					continue
				}
				hasData = true
				img.SetRGBA(px, py, c)
			}
		}
	}

	if !hasData {
		return nil, nil
	}

	return EncodePNG(img)
}

func (r *Renderer) renderSprite(floor int, zoom, x, y int) ([]byte, error) {
	pixelsPerGameTile := 1 << uint(zoom)
	gameTilesPerTile := TileSize / pixelsPerGameTile
	if gameTilesPerTile < 1 {
		gameTilesPerTile = 1
	}

	startGameX := x * gameTilesPerTile
	startGameY := y * gameTilesPerTile

	img := image.NewRGBA(image.Rect(0, 0, TileSize, TileSize))
	hasData := false

	floorStack := BuildFloorStack(floor)

	const margin = 3

	shadowFloor := floor + 1

	for _, f := range floorStack {
		floorDelta := int(f) - floor
		floorOffsetPx := floorDelta * pixelsPerGameTile

		floorMargin := floorDelta
		if floorMargin < 0 {
			floorMargin = -floorMargin
		}
		totalMargin := margin + floorMargin

		for ty := -totalMargin; ty < gameTilesPerTile+totalMargin; ty++ {
			for tx := -totalMargin; tx < gameTilesPerTile+totalMargin; tx++ {
				rawGX := startGameX + tx
				rawGY := startGameY + ty
				if rawGX < 0 || rawGY < 0 || rawGX > 0xFFFF || rawGY > 0xFFFF {
					continue
				}
				gx := uint16(rawGX)
				gy := uint16(rawGY)

				key := otbm.PackPos(gx, gy, f)
				if _, ok := r.gameMap.Tiles[key]; ok {
					r.drawTileSprites(img, gx, gy, f, tx, ty, pixelsPerGameTile, floorOffsetPx, &hasData, startGameX, startGameY, gameTilesPerTile)
				}
			}
		}

		if int(f) == shadowFloor && hasData {
			if r.gpu != nil {
				if result, err := r.gpu.ApplyShadow(img.Pix, TileSize, TileSize, 0.70); err == nil {
					copy(img.Pix, result)
				} else {
					applyShadowOverlay(img, 0.70)
				}
			} else {
				applyShadowOverlay(img, 0.70)
			}
		}
	}

	if !hasData {
		return nil, nil
	}

	return EncodePNG(img)
}

// BuildFloorStack returns the floor rendering order for sprite compositing.
func BuildFloorStack(floor int) []uint8 {
	var stack []uint8
	if floor <= 7 {
		for f := 7; f >= floor; f-- {
			stack = append(stack, uint8(f))
		}
	} else {
		maxDepth := floor + 2
		if maxDepth > 15 {
			maxDepth = 15
		}
		for f := maxDepth; f >= floor; f-- {
			stack = append(stack, uint8(f))
		}
	}
	return stack
}

func (r *Renderer) drawTileSprites(img *image.RGBA, gx, gy uint16, z uint8, tx, ty, pixelsPerGameTile, floorOffsetPx int, hasData *bool, startGameX, startGameY, gameTilesPerTile int) {
	key := otbm.PackPos(gx, gy, z)
	tile, ok := r.gameMap.Tiles[key]
	if !ok {
		return
	}

	destX := tx*pixelsPerGameTile + floorOffsetPx
	destY := ty*pixelsPerGameTile + floorOffsetPx

	var elevation int

	for _, serverID := range tile.Items {
		clientID, ok := r.otb.ServerToClient[serverID]
		if !ok {
			continue
		}
		item, ok := r.dat.Items[clientID]
		if !ok || len(item.SpriteIDs) == 0 {
			continue
		}

		if _, isOnTop := item.Flags[0x03]; isOnTop {
			elevation = 0
		}

		w := int(item.Width)
		h := int(item.Height)

		sprW := w * SpritePixels
		sprH := h * SpritePixels
		composite := image.NewRGBA(image.Rect(0, 0, sprW, sprH))
		anyPixels := false

		xPat := int(gx) % int(item.XDiv)
		yPat := int(gy) % int(item.YDiv)

		for sy := 0; sy < h; sy++ {
			for sx := 0; sx < w; sx++ {
				sprIdx := ((((0*int(item.ZDiv)+0)*int(item.YDiv)+yPat)*int(item.XDiv)+xPat)*int(item.ColorLayers)+0)*h*w + sy*w + sx
				if sprIdx >= len(item.SpriteIDs) {
					sprIdx = sy*w + sx
					if sprIdx >= len(item.SpriteIDs) {
						continue
					}
				}
				sprID := item.SpriteIDs[sprIdx]
				if sprID == 0 {
					continue
				}
				sprImg, err := r.spr.GetRGBA(sprID)
				if err != nil {
					continue
				}
				anyPixels = true

				px := (w - sx - 1) * SpritePixels
				py := (h - sy - 1) * SpritePixels
				draw.Over.Draw(composite, image.Rect(px, py, px+SpritePixels, py+SpritePixels), sprImg, image.Point{})
			}
		}

		if !anyPixels {
			continue
		}

		var dispX, dispY int
		if flagData, has := item.Flags[0x18]; has && len(flagData) >= 4 {
			dispX = int(binary.LittleEndian.Uint16(flagData[0:2]))
			dispY = int(binary.LittleEndian.Uint16(flagData[2:4]))
		}

		var itemElevation int
		if flagData, has := item.Flags[0x19]; has && len(flagData) >= 2 {
			itemElevation = int(binary.LittleEndian.Uint16(flagData))
		}

		src := composite

		outputW := w * pixelsPerGameTile
		outputH := h * pixelsPerGameTile

		scaledDispX := dispX * pixelsPerGameTile / SpritePixels
		scaledDispY := dispY * pixelsPerGameTile / SpritePixels
		scaledElev := elevation * pixelsPerGameTile / SpritePixels

		drawX := destX - (w-1)*pixelsPerGameTile - scaledDispX - scaledElev
		drawY := destY - (h-1)*pixelsPerGameTile - scaledDispY - scaledElev

		if DrawSpriteScaled(img, src, drawX, drawY, outputW, outputH) {
			*hasData = true
		}

		if itemElevation > 0 {
			elevation += itemElevation
			if elevation > 24 {
				elevation = 24
			}
		}
	}
}

// DrawSpriteScaled draws a sprite onto the output image with nearest-neighbor scaling.
func DrawSpriteScaled(dst *image.RGBA, src *image.RGBA, dstX, dstY, outW, outH int) bool {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	dstBounds := dst.Bounds()
	drawn := false

	if outW == srcW && outH == srcH {
		for sy := 0; sy < srcH; sy++ {
			dy := dstY + sy
			if dy < dstBounds.Min.Y || dy >= dstBounds.Max.Y {
				continue
			}
			for sx := 0; sx < srcW; sx++ {
				dx := dstX + sx
				if dx < dstBounds.Min.X || dx >= dstBounds.Max.X {
					continue
				}
				c := src.RGBAAt(srcBounds.Min.X+sx, srcBounds.Min.Y+sy)
				if c.A == 0 {
					continue
				}
				drawn = true
				if c.A == 255 {
					dst.SetRGBA(dx, dy, c)
				} else {
					bg := dst.RGBAAt(dx, dy)
					dst.SetRGBA(dx, dy, BlendOver(c, bg))
				}
			}
		}
		return drawn
	}

	for dy := 0; dy < outH; dy++ {
		iy := dstY + dy
		if iy < dstBounds.Min.Y || iy >= dstBounds.Max.Y {
			continue
		}
		srcY := dy * srcH / outH
		if srcY >= srcH {
			srcY = srcH - 1
		}
		for dx := 0; dx < outW; dx++ {
			ix := dstX + dx
			if ix < dstBounds.Min.X || ix >= dstBounds.Max.X {
				continue
			}
			srcX := dx * srcW / outW
			if srcX >= srcW {
				srcX = srcW - 1
			}
			c := src.RGBAAt(srcBounds.Min.X+srcX, srcBounds.Min.Y+srcY)
			if c.A == 0 {
				continue
			}
			drawn = true
			if c.A == 255 {
				dst.SetRGBA(ix, iy, c)
			} else {
				bg := dst.RGBAAt(ix, iy)
				dst.SetRGBA(ix, iy, BlendOver(c, bg))
			}
		}
	}
	return drawn
}

func applyShadowOverlay(img *image.RGBA, intensity float64) {
	mult := int(256 * (1.0 - intensity))
	if mult < 0 {
		mult = 0
	}
	if mult > 256 {
		mult = 256
	}

	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		rowOff := (y - b.Min.Y) * img.Stride
		for x := b.Min.X; x < b.Max.X; x++ {
			off := rowOff + (x-b.Min.X)*4
			if img.Pix[off+3] == 0 {
				continue
			}
			img.Pix[off+0] = uint8(int(img.Pix[off+0]) * mult >> 8)
			img.Pix[off+1] = uint8(int(img.Pix[off+1]) * mult >> 8)
			img.Pix[off+2] = uint8(int(img.Pix[off+2]) * mult >> 8)
		}
	}
}

// BlendOver performs alpha-over compositing of src over dst.
func BlendOver(src, dst color.RGBA) color.RGBA {
	sa := uint32(src.A)
	da := uint32(dst.A)
	outA := sa + da*(255-sa)/255
	if outA == 0 {
		return color.RGBA{}
	}
	return color.RGBA{
		R: uint8((uint32(src.R)*sa + uint32(dst.R)*da*(255-sa)/255) / outA),
		G: uint8((uint32(src.G)*sa + uint32(dst.G)*da*(255-sa)/255) / outA),
		B: uint8((uint32(src.B)*sa + uint32(dst.B)*da*(255-sa)/255) / outA),
		A: uint8(outA),
	}
}

func (r *Renderer) minimapColor(x, y uint16, z uint8) (c color.RGBA, ok bool) {
	key := otbm.PackPos(x, y, z)
	tile, exists := r.gameMap.Tiles[key]
	if !exists {
		return c, false
	}

	for _, serverID := range tile.Items {
		clientID, mapped := r.otb.ServerToClient[serverID]
		if !mapped {
			continue
		}
		item, found := r.dat.Items[clientID]
		if !found {
			continue
		}
		flagData, has := item.Flags[0x1C]
		if !has || len(flagData) < 2 {
			continue
		}
		colorIdx := binary.LittleEndian.Uint16(flagData)
		c = From8Bit(uint8(colorIdx))
		ok = true
	}
	return c, ok
}

// EncodePNG encodes an RGBA image to PNG bytes using the shared buffer pool.
func EncodePNG(img *image.RGBA) ([]byte, error) {
	buf := gpu.GetBuf()
	defer gpu.PutBuf(buf)
	buf.Grow(TileSize * TileSize)
	if err := pngEncoder.Encode(buf, img); err != nil {
		return nil, err
	}
	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}
