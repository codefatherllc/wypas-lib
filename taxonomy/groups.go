package taxonomy

import (
	"fmt"

	"github.com/codefatherllc/wypas-lib/v2/gamedata"
)

var KnownGroundNames = map[uint16]string{
	129: "stone_floor",
	24:  "grass",
	114: "cave_wall",
	121: "earth",
	210: "passage",
	186: "wall",
	40:  "water",
	192: "lava",
	215: "snow",
	179: "ice",
	30:  "swamp",
	207: "structure",
}

func GroupByMinimapColor(items []gamedata.ItemType, roleMap map[uint16]Role) map[uint16]*SemanticGroup {
	groups := make(map[uint16]*SemanticGroup)
	idx := 0

	for i := range items {
		it := &items[i]
		role, ok := roleMap[it.ServerID]
		if !ok || role == SKIP {
			continue
		}
		a := attrs(it)
		if a.MinimapColor == 0 {
			continue
		}

		g, exists := groups[a.MinimapColor]
		if !exists {
			name := KnownGroundNames[a.MinimapColor]
			if name == "" {
				name = fmt.Sprintf("unknown_%d", a.MinimapColor)
			}
			g = &SemanticGroup{
				Index:        idx,
				Name:         name,
				MinimapColor: a.MinimapColor,
				Role:         role,
			}
			groups[a.MinimapColor] = g
			idx++
		}
		g.Items = append(g.Items, it.ServerID)
	}

	return groups
}

func attrs(it *gamedata.ItemType) gamedata.ItemTypeAttributes {
	a, _ := it.Attributes.GetAttributes()
	if a == nil {
		return gamedata.ItemTypeAttributes{}
	}
	return *a
}
