package taxonomy

import (
	"fmt"

	"github.com/codefatherllc/wypas-lib/gamedata"
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
		if it.MinimapColor == 0 {
			continue
		}

		g, exists := groups[it.MinimapColor]
		if !exists {
			name := KnownGroundNames[it.MinimapColor]
			if name == "" {
				name = fmt.Sprintf("unknown_%d", it.MinimapColor)
			}
			g = &SemanticGroup{
				Index:        idx,
				Name:         name,
				MinimapColor: it.MinimapColor,
				Role:         role,
			}
			groups[it.MinimapColor] = g
			idx++
		}
		g.Items = append(g.Items, it.ServerID)
	}

	return groups
}
