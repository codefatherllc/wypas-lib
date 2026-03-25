package maptile

import "image/color"

// From8Bit converts an 8-bit Tibia minimap color index to RGBA.
// The color space uses a 6x6x6 color cube (216 colors).
func From8Bit(c uint8) color.RGBA {
	r := uint8((int(c) / 36 % 6) * 51)
	g := uint8((int(c) / 6 % 6) * 51)
	b := uint8((int(c) % 6) * 51)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}
