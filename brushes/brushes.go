package brushes

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

type Item struct {
	ID     int `xml:"id,attr" json:"id"`
	Chance int `xml:"chance,attr" json:"chance,omitempty"`
}

type BorderRef struct {
	Align string `xml:"align,attr" json:"align"`
	ID    int    `xml:"id,attr" json:"id,omitempty"`
	To    string `xml:"to,attr" json:"to,omitempty"`
}

type FriendRef struct {
	Name string `xml:"name,attr" json:"name"`
}

type CompositeTile struct {
	X     int    `xml:"x,attr" json:"x"`
	Y     int    `xml:"y,attr" json:"y"`
	Items []Item `xml:"item" json:"items"`
}

type Composite struct {
	Chance int             `xml:"chance,attr" json:"chance"`
	Tiles  []CompositeTile `xml:"tile" json:"tiles"`
}

type CarpetPiece struct {
	Align string `xml:"align,attr" json:"align"`
	ID    int    `xml:"id,attr" json:"id"`
}

type TableAlign struct {
	XMLName xml.Name `xml:"table"`
	Align   string   `xml:"align,attr" json:"align"`
	Items   []Item   `xml:"item" json:"items"`
}

type Door struct {
	ID   int    `xml:"id,attr" json:"id"`
	Type string `xml:"type,attr" json:"type"`
	Open string `xml:"open,attr" json:"open,omitempty"`
}

type WallSection struct {
	XMLName xml.Name `xml:"wall"`
	Type    string   `xml:"type,attr" json:"type"`
	Items   []Item   `xml:"item" json:"items"`
	Doors   []Door   `xml:"door" json:"doors,omitempty"`
}

type Alternate struct {
	Items []Item `xml:"item" json:"items"`
}

type Brush struct {
	Name         string        `xml:"name,attr" json:"name"`
	Type         string        `xml:"type,attr" json:"type"`
	ServerLookID int           `xml:"server_lookid,attr" json:"serverLookId,omitempty"`
	ZOrder       int           `xml:"z-order,attr" json:"zOrder,omitempty"`
	Draggable    string        `xml:"draggable,attr" json:"draggable,omitempty"`
	OnBlocking   string        `xml:"on_blocking,attr" json:"onBlocking,omitempty"`
	Thickness    string        `xml:"thickness,attr" json:"thickness,omitempty"`
	Items        []Item        `xml:"item" json:"items,omitempty"`
	Borders      []BorderRef   `xml:"border" json:"borders,omitempty"`
	Friends      []FriendRef   `xml:"friend" json:"friends,omitempty"`
	Composites   []Composite   `xml:"composite" json:"composites,omitempty"`
	Carpets      []CarpetPiece `xml:"carpet" json:"carpets,omitempty"`
	Tables       []TableAlign  `xml:"table" json:"tables,omitempty"`
	Walls        []WallSection `xml:"wall" json:"walls,omitempty"`
	Alternates   *Alternate    `xml:"alternate" json:"alternate,omitempty"`
}

type BorderItem struct {
	Edge string `xml:"edge,attr" json:"edge"`
	Item int    `xml:"item,attr" json:"item"`
}

type Border struct {
	ID    int          `xml:"id,attr" json:"id"`
	Group int          `xml:"group,attr" json:"group,omitempty"`
	Items []BorderItem `xml:"borderitem" json:"items"`
}

type TilesetBrushRef struct {
	Name string `xml:"name,attr" json:"name"`
}

type TilesetItem struct {
	ID     int `xml:"id,attr" json:"id,omitempty"`
	FromID int `xml:"fromid,attr" json:"fromid,omitempty"`
	ToID   int `xml:"toid,attr" json:"toid,omitempty"`
}

type TilesetSection struct {
	Brushes []TilesetBrushRef `xml:"brush" json:"brushes,omitempty"`
	Items   []TilesetItem     `xml:"item" json:"items,omitempty"`
}

type Tileset struct {
	Name           string          `xml:"name,attr" json:"name"`
	Terrain        *TilesetSection `xml:"terrain" json:"terrain,omitempty"`
	Doodad         *TilesetSection `xml:"doodad" json:"doodad,omitempty"`
	Raw            *TilesetSection `xml:"raw" json:"raw,omitempty"`
	WallDeco       *TilesetSection `xml:"wall_deco" json:"wall_deco,omitempty"`
	OptionalBorder *TilesetSection `xml:"optional_border" json:"optional_border,omitempty"`
	Creature       *TilesetSection `xml:"creature" json:"creature,omitempty"`
}

type xmlMaterials struct {
	XMLName  xml.Name  `xml:"materials"`
	Brushes  []Brush   `xml:"brush"`
	Borders  []Border  `xml:"border"`
	Tilesets []Tileset `xml:"tileset"`
}

type Registry struct {
	Grounds   []Brush   `json:"grounds"`
	Borders   []Border  `json:"borders"`
	Walls     []Brush   `json:"walls"`
	WallDecos []Brush   `json:"wallDecos"`
	Doodads   []Brush   `json:"doodads"`
	Carpets   []Brush   `json:"carpets"`
	Tables    []Brush   `json:"tables"`
	Tilesets  []Tileset `json:"tilesets"`
}

func LoadFromDir(dir string) (*Registry, error) {
	reg := &Registry{}

	type fileTarget struct {
		file   string
		parse  func(data []byte) error
	}

	targets := []fileTarget{
		{"brushes/grounds.xml", func(data []byte) error {
			var m xmlMaterials
			if err := xml.Unmarshal(data, &m); err != nil {
				return err
			}
			reg.Grounds = filterByType(m.Brushes, "ground")
			return nil
		}},
		{"brushes/borders.xml", func(data []byte) error {
			var m xmlMaterials
			if err := xml.Unmarshal(data, &m); err != nil {
				return err
			}
			reg.Borders = m.Borders
			return nil
		}},
		{"brushes/walls.xml", func(data []byte) error {
			var m xmlMaterials
			if err := xml.Unmarshal(data, &m); err != nil {
				return err
			}
			reg.Walls = filterByType(m.Brushes, "wall")
			reg.WallDecos = filterByType(m.Brushes, "wall decoration")
			return nil
		}},
		{"brushes/doodads.xml", func(data []byte) error {
			var m xmlMaterials
			if err := xml.Unmarshal(data, &m); err != nil {
				return err
			}
			reg.Doodads = filterByType(m.Brushes, "doodad")
			return nil
		}},
		{"brushes/carpets.xml", func(data []byte) error {
			var m xmlMaterials
			if err := xml.Unmarshal(data, &m); err != nil {
				return err
			}
			reg.Carpets = filterByType(m.Brushes, "carpet")
			return nil
		}},
		{"brushes/tables.xml", func(data []byte) error {
			var m xmlMaterials
			if err := xml.Unmarshal(data, &m); err != nil {
				return err
			}
			reg.Tables = filterByType(m.Brushes, "table")
			return nil
		}},
		{"materials/tilesets.xml", func(data []byte) error {
			var m xmlMaterials
			if err := xml.Unmarshal(data, &m); err != nil {
				return err
			}
			reg.Tilesets = m.Tilesets
			return nil
		}},
	}

	for _, t := range targets {
		path := filepath.Join(dir, t.file)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", t.file, err)
		}
		if err := t.parse(data); err != nil {
			return nil, fmt.Errorf("parse %s: %w", t.file, err)
		}
	}

	return reg, nil
}

func (r *Registry) BrushesByType(typ string) []Brush {
	switch typ {
	case "ground":
		return r.Grounds
	case "border":
		return nil
	case "wall":
		return r.Walls
	case "wall_deco":
		return r.WallDecos
	case "doodad":
		return r.Doodads
	case "carpet":
		return r.Carpets
	case "table":
		return r.Tables
	case "optional_border":
		return r.Grounds
	default:
		return nil
	}
}

type xmlExtension struct {
	XMLName     xml.Name  `xml:"materialsextension"`
	Name        string    `xml:"name,attr"`
	Author      string    `xml:"author,attr"`
	Description string    `xml:"description,attr"`
	Brushes     []Brush   `xml:"brush"`
	Borders     []Border  `xml:"border"`
	Tilesets    []Tileset `xml:"tileset"`
}

type ExtensionMeta struct {
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Brushes     int    `json:"brushes"`
	Tilesets    int    `json:"tilesets"`
}

func (r *Registry) LoadExtension(data []byte) (*ExtensionMeta, error) {
	var ext xmlExtension
	if err := xml.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("parse extension xml: %w", err)
	}

	for _, b := range ext.Brushes {
		switch b.Type {
		case "ground":
			r.Grounds = appendOrReplace(r.Grounds, b)
		case "wall":
			r.Walls = appendOrReplace(r.Walls, b)
		case "wall decoration":
			r.WallDecos = appendOrReplace(r.WallDecos, b)
		case "doodad":
			r.Doodads = appendOrReplace(r.Doodads, b)
		case "carpet":
			r.Carpets = appendOrReplace(r.Carpets, b)
		case "table":
			r.Tables = appendOrReplace(r.Tables, b)
		}
	}
	r.Borders = append(r.Borders, ext.Borders...)

	for _, ts := range ext.Tilesets {
		merged := false
		for i, existing := range r.Tilesets {
			if existing.Name == ts.Name {
				r.Tilesets[i] = mergeTileset(existing, ts)
				merged = true
				break
			}
		}
		if !merged {
			r.Tilesets = append(r.Tilesets, ts)
		}
	}

	return &ExtensionMeta{
		Name:        ext.Name,
		Author:      ext.Author,
		Description: ext.Description,
		Brushes:     len(ext.Brushes),
		Tilesets:    len(ext.Tilesets),
	}, nil
}

func appendOrReplace(list []Brush, b Brush) []Brush {
	for i, existing := range list {
		if existing.Name == b.Name {
			list[i] = b
			return list
		}
	}
	return append(list, b)
}

func mergeTileset(base, ext Tileset) Tileset {
	if ext.Terrain != nil {
		if base.Terrain == nil {
			base.Terrain = ext.Terrain
		} else {
			base.Terrain.Brushes = append(base.Terrain.Brushes, ext.Terrain.Brushes...)
			base.Terrain.Items = append(base.Terrain.Items, ext.Terrain.Items...)
		}
	}
	if ext.Doodad != nil {
		if base.Doodad == nil {
			base.Doodad = ext.Doodad
		} else {
			base.Doodad.Brushes = append(base.Doodad.Brushes, ext.Doodad.Brushes...)
			base.Doodad.Items = append(base.Doodad.Items, ext.Doodad.Items...)
		}
	}
	if ext.Raw != nil {
		if base.Raw == nil {
			base.Raw = ext.Raw
		} else {
			base.Raw.Brushes = append(base.Raw.Brushes, ext.Raw.Brushes...)
			base.Raw.Items = append(base.Raw.Items, ext.Raw.Items...)
		}
	}
	return base
}

func filterByType(brushes []Brush, typ string) []Brush {
	out := make([]Brush, 0, len(brushes))
	for _, b := range brushes {
		if b.Type == typ || b.Type == "" {
			out = append(out, b)
		}
	}
	return out
}
