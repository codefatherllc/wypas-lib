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
	LOOT
	INTERACTABLE
	ENVIRONMENT
)

var roleNames = [...]string{
	"GROUND", "GROUND_BORDER", "WALL", "BLOCKING", "DOOR",
	"STAIRCASE", "CONTAINER", "TELEPORT", "MAGICFIELD",
	"DECORATION", "SKIP", "LOOT", "INTERACTABLE", "ENVIRONMENT",
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

	if it.HasFlag(gamedata.FlagPickupable) {
		return LOOT
	}
	if it.HasFlag(gamedata.FlagMovable) {
		return INTERACTABLE
	}
	if it.FloorChange > 0 {
		return STAIRCASE
	}
	if it.TopOrder == 1 {
		return GROUND_BORDER
	}
	if it.TopOrder == 2 {
		return DECORATION
	}
	if !it.HasFlag(gamedata.FlagBlockSolid) {
		return ENVIRONMENT
	}
	if it.HasFlag(gamedata.FlagVertical) || it.HasFlag(gamedata.FlagHorizontal) {
		return WALL
	}
	return BLOCKING
}
