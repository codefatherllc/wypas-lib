package taxonomy

import "github.com/codefatherllc/wypas-lib/gamedata"

type Role int

const (
	GROUND Role = iota
	GROUND_BORDER
	WALL
	BLOCKING
	DOOR
	STAIRCASE
	CONTAINER
	TELEPORT
	MAGICFIELD
	DECORATION
	SKIP
)

var roleNames = [...]string{
	"GROUND", "GROUND_BORDER", "WALL", "BLOCKING", "DOOR",
	"STAIRCASE", "CONTAINER", "TELEPORT", "MAGICFIELD",
	"DECORATION", "SKIP",
}

func (r Role) String() string {
	if int(r) < len(roleNames) {
		return roleNames[r]
	}
	return "UNKNOWN"
}

func ClassifyItem(it *gamedata.ItemType) Role {
	if it.ItemGroup > 0 {
		switch it.ItemGroup {
		case 1:
			return GROUND
		case 13:
			return DOOR
		case 2:
			return CONTAINER
		case 7:
			return TELEPORT
		case 8:
			return MAGICFIELD
		case 3, 4, 5, 6, 14:
			return SKIP
		}
	}

	if it.Floorchange > 0 {
		return STAIRCASE
	}
	if it.BlockSolid && (it.IsVertical || it.IsHorizontal) && !it.Pickupable {
		return WALL
	}
	if it.BlockSolid && !it.IsVertical && !it.IsHorizontal && !it.Pickupable {
		return BLOCKING
	}
	if it.TopOrder == 1 && !it.BlockSolid && !it.Pickupable {
		return GROUND_BORDER
	}
	return DECORATION
}
