package autobrush

import (
	"math/rand"
	"sort"
	"strings"

	"github.com/codefatherllc/wypas-lib/brushes"
)

type Pos [3]int

type TileData struct {
	GroundID int      `json:"groundId"`
	Flags    int      `json:"flags"`
	HouseID  int      `json:"houseId"`
	Items    []uint16 `json:"items"`
}

type TileChange struct {
	X        int      `json:"x"`
	Y        int      `json:"y"`
	Z        int      `json:"z"`
	GroundID int      `json:"groundId"`
	Flags    int      `json:"flags"`
	HouseID  int      `json:"houseId"`
	Items    []uint16 `json:"items"`
}

type resolvedBrush struct {
	name    string
	zOrder  int
	borders []brushes.BorderRef
	friends []string
}

var neighborOffsets = [8][2]int{
	{-1, -1}, // NW = 0
	{0, -1},  // N  = 1
	{1, -1},  // NE = 2
	{-1, 0},  // W  = 3
	{1, 0},   // E  = 4
	{-1, 1},  // SW = 5
	{0, 1},   // S  = 6
	{1, 1},   // SE = 7
}

func findGroundBrush(reg *brushes.Registry, groundID int) *brushes.Brush {
	for i := range reg.Grounds {
		for _, item := range reg.Grounds[i].Items {
			if item.ID == groundID {
				return &reg.Grounds[i]
			}
		}
	}
	return nil
}

func findGroundBrushByName(reg *brushes.Registry, name string) *brushes.Brush {
	lower := strings.ToLower(name)
	for i := range reg.Grounds {
		if strings.ToLower(reg.Grounds[i].Name) == lower {
			return &reg.Grounds[i]
		}
	}
	return nil
}

func findBorderDef(reg *brushes.Registry, borderID int) *brushes.Border {
	for i := range reg.Borders {
		if reg.Borders[i].ID == borderID {
			return &reg.Borders[i]
		}
	}
	return nil
}

func isFriend(a, b *brushes.Brush) bool {
	for _, f := range a.Friends {
		if strings.EqualFold(f.Name, b.Name) {
			return true
		}
	}
	return false
}

func PickRandomItem(items []brushes.Item) int {
	if len(items) == 0 {
		return 0
	}
	totalChance := 0
	for _, it := range items {
		totalChance += it.Chance
	}
	if totalChance <= 0 {
		return items[0].ID
	}
	r := rand.Intn(totalChance) + 1
	cumulative := 0
	for _, it := range items {
		cumulative += it.Chance
		if r <= cumulative {
			return it.ID
		}
	}
	return items[0].ID
}

func edgeNameToIdx(edge string) int {
	switch edge {
	case "n":
		return NorthHorizontal
	case "e":
		return EastHorizontal
	case "s":
		return SouthHorizontal
	case "w":
		return WestHorizontal
	case "cnw":
		return NorthwestCorner
	case "cne":
		return NortheastCorner
	case "csw":
		return SouthwestCorner
	case "cse":
		return SoutheastCorner
	case "dnw":
		return NorthwestDiagonal
	case "dne":
		return NortheastDiagonal
	case "dse":
		return SoutheastDiagonal
	case "dsw":
		return SouthwestDiagonal
	}
	return BorderNone
}

func borderToEdgeMap(border *brushes.Border) map[int]int {
	m := make(map[int]int)
	for _, bi := range border.Items {
		idx := edgeNameToIdx(bi.Edge)
		if idx != BorderNone {
			m[idx] = bi.Item
		}
	}
	return m
}

type borderCluster struct {
	alignment uint32
	z         int
	edgeMap   map[int]int
}

func getOuterBorderRef(brush *brushes.Brush) *brushes.BorderRef {
	for i := range brush.Borders {
		if brush.Borders[i].Align == "outer" {
			return &brush.Borders[i]
		}
	}
	return nil
}

func getInnerBorderRef(brush *brushes.Brush, toName string) *brushes.BorderRef {
	for i := range brush.Borders {
		if brush.Borders[i].Align == "inner" {
			if toName == "" || brush.Borders[i].To == "" || brush.Borders[i].To == "none" || strings.EqualFold(brush.Borders[i].To, toName) {
				return &brush.Borders[i]
			}
		}
	}
	return nil
}

func ApplyGroundBrush(tiles map[Pos]TileData, center Pos, brushName string, reg *brushes.Registry) {
	brush := findGroundBrushByName(reg, brushName)
	if brush == nil || len(brush.Items) == 0 {
		return
	}

	groundID := PickRandomItem(brush.Items)
	if groundID == 0 {
		return
	}

	centerTile := tiles[center]
	centerTile.GroundID = groundID
	tiles[center] = centerTile
}

// RecomputeBordersAround recalculates borders for a tile and all its neighbors.
// Call this AFTER all ground placements in a batch for correct results.
func RecomputeBordersAround(tiles map[Pos]TileData, center Pos, reg *brushes.Registry) []TileChange {
	affected := make(map[Pos]bool)
	affected[center] = true
	for _, off := range neighborOffsets {
		p := Pos{center[0] + off[0], center[1] + off[1], center[2]}
		affected[p] = true
	}

	var changes []TileChange
	for pos := range affected {
		borderItems := ComputeBordersForTile(pos, tiles, reg)
		td := tiles[pos]
		td.Items = borderItems
		tiles[pos] = td
		changes = append(changes, TileChangeFromData(pos, td))
	}
	return changes
}

func ComputeBordersForTile(pos Pos, tiles map[Pos]TileData, reg *brushes.Registry) []uint16 {
	td := tiles[pos]
	centerBrush := findGroundBrush(reg, td.GroundID)

	neighbors := [8]*brushes.Brush{}
	for i, off := range neighborOffsets {
		np := Pos{pos[0] + off[0], pos[1] + off[1], pos[2]}
		if ntd, ok := tiles[np]; ok {
			neighbors[i] = findGroundBrush(reg, ntd.GroundID)
		}
	}

	type visitPair struct {
		visited bool
		brush   *brushes.Brush
	}
	neigh := [8]visitPair{}
	for i := 0; i < 8; i++ {
		neigh[i] = visitPair{false, neighbors[i]}
	}

	var clusters []borderCluster

	for i := 0; i < 8; i++ {
		if neigh[i].visited {
			continue
		}

		other := neigh[i].brush

		if centerBrush != nil && other != nil {
			if other.Name == centerBrush.Name {
				continue
			}

			if isFriend(centerBrush, other) || isFriend(other, centerBrush) {
				neigh[i].visited = true
				continue
			}

			tiledata := uint32(0)
			for j := i; j < 8; j++ {
				if !neigh[j].visited && neigh[j].brush != nil && neigh[j].brush.Name == other.Name {
					neigh[j].visited = true
					tiledata |= 1 << uint(j)
				}
			}

			if tiledata != 0 {
				border := resolveBorder(centerBrush, other, reg)
				if border != nil {
					em := borderToEdgeMap(border)
					found := false
					for k := range clusters {
						if clusters[k].z == other.ZOrder {
							clusters[k].alignment |= tiledata
							found = true
							break
						}
					}
					if !found {
						clusters = append(clusters, borderCluster{
							alignment: tiledata,
							z:         other.ZOrder,
							edgeMap:   em,
						})
					}
				}
			}
		} else if centerBrush != nil && other == nil {
			innerRef := getInnerBorderRef(centerBrush, "none")
			if innerRef != nil && innerRef.ID > 0 {
				tiledata := uint32(0)
				for j := i; j < 8; j++ {
					if !neigh[j].visited && neigh[j].brush == nil {
						neigh[j].visited = true
						tiledata |= 1 << uint(j)
					}
				}
				if tiledata != 0 {
					border := findBorderDef(reg, innerRef.ID)
					if border != nil {
						clusters = append(clusters, borderCluster{
							alignment: tiledata,
							z:         5000,
							edgeMap:   borderToEdgeMap(border),
						})
					}
				}
			}
			neigh[i].visited = true
		} else if centerBrush == nil && other != nil {
			outerRef := getOuterBorderRef(other)
			if outerRef != nil && outerRef.ID > 0 {
				tiledata := uint32(0)
				for j := i; j < 8; j++ {
					if !neigh[j].visited && neigh[j].brush != nil && neigh[j].brush.Name == other.Name {
						neigh[j].visited = true
						tiledata |= 1 << uint(j)
					}
				}
				if tiledata != 0 {
					border := findBorderDef(reg, outerRef.ID)
					if border != nil {
						clusters = append(clusters, borderCluster{
							alignment: tiledata,
							z:         other.ZOrder,
							edgeMap:   borderToEdgeMap(border),
						})
					}
				}
			}
			neigh[i].visited = true
		} else {
			neigh[i].visited = true
		}
	}

	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].z < clusters[j].z
	})

	var items []uint16
	for idx := len(clusters) - 1; idx >= 0; idx-- {
		cl := clusters[idx]
		if cl.alignment == 0 {
			continue
		}

		dirs := unpackDirections(groundBorderTypes[cl.alignment])
		for _, dir := range dirs {
			if dir == BorderNone {
				break
			}
			if itemID, ok := cl.edgeMap[dir]; ok && itemID > 0 {
				items = append(items, uint16(itemID))
			} else {
				// Fallback for diagonal: split into two edges
				switch dir {
				case NorthwestDiagonal:
					addEdge(&items, cl.edgeMap, WestHorizontal)
					addEdge(&items, cl.edgeMap, NorthHorizontal)
				case NortheastDiagonal:
					addEdge(&items, cl.edgeMap, EastHorizontal)
					addEdge(&items, cl.edgeMap, NorthHorizontal)
				case SouthwestDiagonal:
					addEdge(&items, cl.edgeMap, SouthHorizontal)
					addEdge(&items, cl.edgeMap, WestHorizontal)
				case SoutheastDiagonal:
					addEdge(&items, cl.edgeMap, SouthHorizontal)
					addEdge(&items, cl.edgeMap, EastHorizontal)
				}
			}
		}
	}

	return items
}

func addEdge(items *[]uint16, edgeMap map[int]int, dir int) {
	if id, ok := edgeMap[dir]; ok && id > 0 {
		*items = append(*items, uint16(id))
	}
}

func resolveBorder(center, other *brushes.Brush, reg *brushes.Registry) *brushes.Border {
	if center.ZOrder < other.ZOrder {
		outerRef := getOuterBorderRef(other)
		if outerRef != nil && outerRef.ID > 0 {
			return findBorderDef(reg, outerRef.ID)
		}
		innerRef := getInnerBorderRef(center, "")
		if innerRef != nil && innerRef.ID > 0 {
			return findBorderDef(reg, innerRef.ID)
		}
	} else {
		innerRef := getInnerBorderRef(center, "")
		if innerRef != nil && innerRef.ID > 0 {
			return findBorderDef(reg, innerRef.ID)
		}
		outerRef := getOuterBorderRef(other)
		if outerRef != nil && outerRef.ID > 0 {
			return findBorderDef(reg, outerRef.ID)
		}
	}
	return nil
}

func TileChangeFromData(pos Pos, td TileData) TileChange {
	items := td.Items
	if items == nil {
		items = []uint16{}
	}
	return TileChange{
		X:        pos[0],
		Y:        pos[1],
		Z:        pos[2],
		GroundID: td.GroundID,
		Flags:    td.Flags,
		HouseID:  td.HouseID,
		Items:    items,
	}
}

func ApplyWallBrush(tiles map[Pos]TileData, center Pos, brushName string, reg *brushes.Registry) []TileChange {
	wallBrush := findWallBrushByName(reg, brushName)
	if wallBrush == nil {
		return nil
	}

	wallItemID := pickWallItem(wallBrush, WallHorizontal)
	if wallItemID == 0 {
		for i := 0; i < 17; i++ {
			wallItemID = pickWallItem(wallBrush, i)
			if wallItemID > 0 {
				break
			}
		}
	}
	if wallItemID == 0 {
		return nil
	}

	td := tiles[center]
	td.Items = appendIfNotPresent(td.Items, uint16(wallItemID))
	tiles[center] = td

	affected := []Pos{center}
	cardinalOffsets := [4][2]int{{0, -1}, {-1, 0}, {1, 0}, {0, 1}}
	for _, off := range cardinalOffsets {
		np := Pos{center[0] + off[0], center[1] + off[1], center[2]}
		if _, ok := tiles[np]; ok {
			affected = append(affected, np)
		}
	}

	var changes []TileChange
	for _, pos := range affected {
		doWallBorders(pos, tiles, wallBrush, reg)
		changes = append(changes, TileChangeFromData(pos, tiles[pos]))
	}
	return changes
}

func doWallBorders(pos Pos, tiles map[Pos]TileData, wallBrush *brushes.Brush, reg *brushes.Registry) {
	td := tiles[pos]

	var wallItemIdx int = -1
	for i, itemID := range td.Items {
		if isWallItem(wallBrush, int(itemID)) {
			wallItemIdx = i
			break
		}
	}
	if wallItemIdx < 0 {
		return
	}

	cardinalOffsets := [4][2]int{{0, -1}, {-1, 0}, {1, 0}, {0, 1}} // N, W, E, S
	var tiledata uint32
	for i, off := range cardinalOffsets {
		np := Pos{pos[0] + off[0], pos[1] + off[1], pos[2]}
		if ntd, ok := tiles[np]; ok {
			for _, itemID := range ntd.Items {
				if isWallItem(wallBrush, int(itemID)) {
					tiledata |= 1 << uint(i)
					break
				}
			}
		}
	}

	bt := int(wallFullBorderTypes[tiledata])
	newID := pickWallItem(wallBrush, bt)
	if newID == 0 {
		bt = int(wallHalfBorderTypes[tiledata])
		newID = pickWallItem(wallBrush, bt)
	}
	if newID > 0 {
		td.Items[wallItemIdx] = uint16(newID)
		tiles[pos] = td
	}
}

func findWallBrushByName(reg *brushes.Registry, name string) *brushes.Brush {
	lower := strings.ToLower(name)
	for i := range reg.Walls {
		if strings.ToLower(reg.Walls[i].Name) == lower {
			return &reg.Walls[i]
		}
	}
	return nil
}

func wallTypeToString(bt int) string {
	switch bt {
	case WallHorizontal:
		return "horizontal"
	case WallVertical:
		return "vertical"
	case WallNorthwestDiagonal:
		return "corner"
	case WallPole:
		return "pole"
	case WallSouthEnd:
		return "south end"
	case WallEastEnd:
		return "east end"
	case WallNorthEnd:
		return "north end"
	case WallWestEnd:
		return "west end"
	case WallSouthT:
		return "south T"
	case WallEastT:
		return "east T"
	case WallWestT:
		return "west T"
	case WallNorthT:
		return "north T"
	case WallNortheastDiagonal:
		return "northeast diagonal"
	case WallSouthwestDiagonal:
		return "southwest diagonal"
	case WallSoutheastDiagonal:
		return "southeast diagonal"
	case WallIntersection:
		return "intersection"
	}
	return ""
}

func pickWallItem(brush *brushes.Brush, bt int) int {
	typeName := wallTypeToString(bt)
	if typeName == "" {
		return 0
	}
	for _, ws := range brush.Walls {
		if strings.EqualFold(ws.Type, typeName) && len(ws.Items) > 0 {
			return PickRandomItem(ws.Items)
		}
	}
	return 0
}

func isWallItem(brush *brushes.Brush, itemID int) bool {
	for _, ws := range brush.Walls {
		for _, it := range ws.Items {
			if it.ID == itemID {
				return true
			}
		}
	}
	return false
}

func ApplyTableBrush(tiles map[Pos]TileData, center Pos, brushName string, reg *brushes.Registry) []TileChange {
	brush := findTableBrushByName(reg, brushName)
	if brush == nil {
		return nil
	}

	itemID := pickTableItem(brush, "alone")
	if itemID == 0 {
		itemID = pickTableItem(brush, "horizontal")
	}
	if itemID == 0 {
		return nil
	}

	td := tiles[center]
	td.Items = appendIfNotPresent(td.Items, uint16(itemID))
	tiles[center] = td

	affected := []Pos{center}
	cardinalOffsets := [4][2]int{{0, -1}, {-1, 0}, {1, 0}, {0, 1}}
	for _, off := range cardinalOffsets {
		np := Pos{center[0] + off[0], center[1] + off[1], center[2]}
		if _, ok := tiles[np]; ok {
			affected = append(affected, np)
		}
	}

	var changes []TileChange
	for _, pos := range affected {
		doTableBorders(pos, tiles, brush)
		changes = append(changes, TileChangeFromData(pos, tiles[pos]))
	}
	return changes
}

func doTableBorders(pos Pos, tiles map[Pos]TileData, brush *brushes.Brush) {
	td := tiles[pos]
	tableIdx := -1
	for i, itemID := range td.Items {
		if isTableItem(brush, int(itemID)) {
			tableIdx = i
			break
		}
	}
	if tableIdx < 0 {
		return
	}

	hasN := hasTableNeighbor(tiles, brush, pos, 0, -1)
	hasS := hasTableNeighbor(tiles, brush, pos, 0, 1)
	hasW := hasTableNeighbor(tiles, brush, pos, -1, 0)
	hasE := hasTableNeighbor(tiles, brush, pos, 1, 0)

	var align string
	switch {
	case !hasN && !hasS && !hasW && !hasE:
		align = "alone"
	case hasW && hasE && !hasN && !hasS:
		align = "horizontal"
	case hasN && hasS && !hasW && !hasE:
		align = "vertical"
	case hasE && !hasW && !hasN && !hasS:
		align = "west"
	case hasW && !hasE && !hasN && !hasS:
		align = "east"
	case hasS && !hasN && !hasW && !hasE:
		align = "north"
	case hasN && !hasS && !hasW && !hasE:
		align = "south"
	default:
		if hasN || hasS {
			align = "vertical"
		} else {
			align = "horizontal"
		}
	}

	newID := pickTableItem(brush, align)
	if newID > 0 {
		td.Items[tableIdx] = uint16(newID)
		tiles[pos] = td
	}
}

func hasTableNeighbor(tiles map[Pos]TileData, brush *brushes.Brush, pos Pos, dx, dy int) bool {
	np := Pos{pos[0] + dx, pos[1] + dy, pos[2]}
	if ntd, ok := tiles[np]; ok {
		for _, itemID := range ntd.Items {
			if isTableItem(brush, int(itemID)) {
				return true
			}
		}
	}
	return false
}

func findTableBrushByName(reg *brushes.Registry, name string) *brushes.Brush {
	lower := strings.ToLower(name)
	for i := range reg.Tables {
		if strings.ToLower(reg.Tables[i].Name) == lower {
			return &reg.Tables[i]
		}
	}
	return nil
}

func pickTableItem(brush *brushes.Brush, align string) int {
	lower := strings.ToLower(align)
	for _, ta := range brush.Tables {
		if strings.ToLower(ta.Align) == lower && len(ta.Items) > 0 {
			return PickRandomItem(ta.Items)
		}
	}
	return 0
}

func isTableItem(brush *brushes.Brush, itemID int) bool {
	for _, ta := range brush.Tables {
		for _, it := range ta.Items {
			if it.ID == itemID {
				return true
			}
		}
	}
	return false
}

func ApplyCarpetBrush(tiles map[Pos]TileData, center Pos, brushName string, reg *brushes.Registry) []TileChange {
	brush := findCarpetBrushByName(reg, brushName)
	if brush == nil {
		return nil
	}

	centerID := carpetPieceID(brush, "center")
	if centerID == 0 {
		return nil
	}

	td := tiles[center]
	td.Items = appendIfNotPresent(td.Items, uint16(centerID))
	tiles[center] = td

	affected := make(map[Pos]bool)
	affected[center] = true
	for _, off := range neighborOffsets {
		p := Pos{center[0] + off[0], center[1] + off[1], center[2]}
		affected[p] = true
	}

	var changes []TileChange
	for pos := range affected {
		doCarpetBorders(pos, tiles, brush)
		changes = append(changes, TileChangeFromData(pos, tiles[pos]))
	}
	return changes
}

func doCarpetBorders(pos Pos, tiles map[Pos]TileData, brush *brushes.Brush) {
	td := tiles[pos]
	carpetIdx := -1
	for i, itemID := range td.Items {
		if isCarpetItem(brush, int(itemID)) {
			carpetIdx = i
			break
		}
	}
	if carpetIdx < 0 {
		return
	}

	// Check 8 neighbors
	hasNeighbor := [8]bool{}
	for i, off := range neighborOffsets {
		np := Pos{pos[0] + off[0], pos[1] + off[1], pos[2]}
		if ntd, ok := tiles[np]; ok {
			for _, itemID := range ntd.Items {
				if isCarpetItem(brush, int(itemID)) {
					hasNeighbor[i] = true
					break
				}
			}
		}
	}

	hasNW, hasN, hasNE := hasNeighbor[0], hasNeighbor[1], hasNeighbor[2]
	hasW, hasE := hasNeighbor[3], hasNeighbor[4]
	hasSW, hasS, hasSE := hasNeighbor[5], hasNeighbor[6], hasNeighbor[7]

	var align string
	switch {
	case hasN && hasS && hasW && hasE:
		align = "center"
	case !hasN && hasS && hasW && hasE:
		align = "n"
	case hasN && !hasS && hasW && hasE:
		align = "s"
	case hasN && hasS && !hasW && hasE:
		align = "w"
	case hasN && hasS && hasW && !hasE:
		align = "e"
	case !hasN && !hasW && hasS && hasE:
		align = "cnw"
	case !hasN && hasW && hasS && !hasE:
		align = "cne"
	case hasN && !hasS && !hasW && hasE:
		align = "csw"
	case hasN && !hasS && hasW && !hasE:
		align = "cse"
	case !hasN && !hasS && !hasW && !hasE:
		align = "center"
	default:
		// Diagonal checks
		if hasN && hasW && !hasNW {
			align = "dnw"
		} else if hasN && hasE && !hasNE {
			align = "dne"
		} else if hasS && hasW && !hasSW {
			align = "dsw"
		} else if hasS && hasE && !hasSE {
			align = "dse"
		} else {
			align = "center"
		}
	}

	newID := carpetPieceID(brush, align)
	if newID > 0 {
		td.Items[carpetIdx] = uint16(newID)
		tiles[pos] = td
	}
}

func findCarpetBrushByName(reg *brushes.Registry, name string) *brushes.Brush {
	lower := strings.ToLower(name)
	for i := range reg.Carpets {
		if strings.ToLower(reg.Carpets[i].Name) == lower {
			return &reg.Carpets[i]
		}
	}
	return nil
}

func carpetPieceID(brush *brushes.Brush, align string) int {
	lower := strings.ToLower(align)
	for _, cp := range brush.Carpets {
		if strings.ToLower(cp.Align) == lower && cp.ID > 0 {
			return cp.ID
		}
	}
	return 0
}

func isCarpetItem(brush *brushes.Brush, itemID int) bool {
	for _, cp := range brush.Carpets {
		if cp.ID == itemID {
			return true
		}
	}
	return false
}

func ApplyDoodadBrush(tiles map[Pos]TileData, center Pos, brushName string, reg *brushes.Registry) []TileChange {
	return ApplyDoodadBrushVariation(tiles, center, brushName, reg, -1)
}

func ApplyDoodadBrushVariation(tiles map[Pos]TileData, center Pos, brushName string, reg *brushes.Registry, variation int) []TileChange {
	brush := findDoodadBrushByName(reg, brushName)
	if brush == nil {
		return nil
	}

	totalVariations := len(brush.Items) + len(brush.Composites)
	if totalVariations == 0 {
		return nil
	}

	if variation >= 0 && variation < totalVariations {
		if variation < len(brush.Items) {
			return applyDoodadSingleItem(tiles, center, brush, brush.Items[variation].ID)
		}
		compIdx := variation - len(brush.Items)
		return applyDoodadCompositeByIndex(tiles, center, brush, compIdx)
	}

	singleChance := 0
	for _, it := range brush.Items {
		singleChance += it.Chance
	}

	compositeChance := 0
	for _, comp := range brush.Composites {
		compositeChance += comp.Chance
	}

	totalChance := singleChance + compositeChance
	if totalChance <= 0 {
		return nil
	}

	roll := rand.Intn(totalChance) + 1

	if roll <= compositeChance {
		return applyDoodadComposite(tiles, center, brush, roll)
	}
	return applyDoodadSingle(tiles, center, brush)
}

func applyDoodadComposite(tiles map[Pos]TileData, center Pos, brush *brushes.Brush, roll int) []TileChange {
	var selected *brushes.Composite
	cumulative := 0
	for i := range brush.Composites {
		cumulative += brush.Composites[i].Chance
		if roll <= cumulative {
			selected = &brush.Composites[i]
			break
		}
	}
	if selected == nil {
		return nil
	}

	var changes []TileChange
	for _, ct := range selected.Tiles {
		pos := Pos{center[0] + ct.X, center[1] + ct.Y, center[2]}
		td := tiles[pos]
		for _, it := range ct.Items {
			if it.ID > 0 {
				td.Items = appendIfNotPresent(td.Items, uint16(it.ID))
			}
		}
		tiles[pos] = td
		changes = append(changes, TileChangeFromData(pos, td))
	}
	return changes
}

func applyDoodadSingleItem(tiles map[Pos]TileData, center Pos, brush *brushes.Brush, itemID int) []TileChange {
	if itemID == 0 {
		return nil
	}
	_ = brush
	td := tiles[center]
	td.Items = appendIfNotPresent(td.Items, uint16(itemID))
	tiles[center] = td
	return []TileChange{TileChangeFromData(center, td)}
}

func applyDoodadCompositeByIndex(tiles map[Pos]TileData, center Pos, brush *brushes.Brush, compIdx int) []TileChange {
	if compIdx < 0 || compIdx >= len(brush.Composites) {
		return nil
	}
	selected := &brush.Composites[compIdx]
	var changes []TileChange
	for _, ct := range selected.Tiles {
		pos := Pos{center[0] + ct.X, center[1] + ct.Y, center[2]}
		td := tiles[pos]
		for _, it := range ct.Items {
			if it.ID > 0 {
				td.Items = appendIfNotPresent(td.Items, uint16(it.ID))
			}
		}
		tiles[pos] = td
		changes = append(changes, TileChangeFromData(pos, td))
	}
	return changes
}

func applyDoodadSingle(tiles map[Pos]TileData, center Pos, brush *brushes.Brush) []TileChange {
	itemID := PickRandomItem(brush.Items)
	if itemID == 0 {
		return nil
	}

	td := tiles[center]
	td.Items = appendIfNotPresent(td.Items, uint16(itemID))
	tiles[center] = td

	return []TileChange{TileChangeFromData(center, td)}
}

func findDoodadBrushByName(reg *brushes.Registry, name string) *brushes.Brush {
	lower := strings.ToLower(name)
	for i := range reg.Doodads {
		if strings.ToLower(reg.Doodads[i].Name) == lower {
			return &reg.Doodads[i]
		}
	}
	return nil
}

func appendIfNotPresent(items []uint16, id uint16) []uint16 {
	for _, existing := range items {
		if existing == id {
			return items
		}
	}
	return append(items, id)
}

func ApplyDoorBrush(tiles map[Pos]TileData, center Pos, wallBrushName, doorType string, reg *brushes.Registry) []TileChange {
	wallBrush := findWallBrushByName(reg, wallBrushName)
	if wallBrush == nil {
		return nil
	}

	if doorType == "" {
		doorType = "normal"
	}

	td := tiles[center]

	wallIdx := -1
	var wallAlignment int
	for i, itemID := range td.Items {
		if isWallItem(wallBrush, int(itemID)) {
			wallIdx = i
			wallAlignment = getWallAlignment(wallBrush, int(itemID))
			break
		}
	}

	if wallIdx < 0 {
		wallItemID := pickWallItem(wallBrush, WallHorizontal)
		if wallItemID == 0 {
			return nil
		}
		td.Items = appendIfNotPresent(td.Items, uint16(wallItemID))
		wallIdx = len(td.Items) - 1
		wallAlignment = WallHorizontal
	}

	doorID := pickDoorItem(wallBrush, wallAlignment, doorType)
	if doorID == 0 {
		return nil
	}

	td.Items[wallIdx] = uint16(doorID)
	tiles[center] = td

	return []TileChange{TileChangeFromData(center, td)}
}

func getWallAlignment(brush *brushes.Brush, itemID int) int {
	for bt := 0; bt <= WallUntouchable; bt++ {
		typeName := wallTypeToString(bt)
		if typeName == "" {
			continue
		}
		for _, ws := range brush.Walls {
			if strings.EqualFold(ws.Type, typeName) {
				for _, it := range ws.Items {
					if it.ID == itemID {
						return bt
					}
				}
			}
		}
	}
	return WallHorizontal
}

func pickDoorItem(brush *brushes.Brush, wallAlignment int, doorType string) int {
	typeName := wallTypeToString(wallAlignment)
	if typeName == "" {
		return 0
	}
	for _, ws := range brush.Walls {
		if !strings.EqualFold(ws.Type, typeName) {
			continue
		}
		for _, d := range ws.Doors {
			if strings.EqualFold(d.Type, doorType) {
				return d.ID
			}
		}
	}
	for _, ws := range brush.Walls {
		if !strings.EqualFold(ws.Type, typeName) {
			continue
		}
		if len(ws.Doors) > 0 {
			return ws.Doors[0].ID
		}
	}
	return 0
}

func ApplyFlagBrush(tiles map[Pos]TileData, center Pos, flagValue int) []TileChange {
	if flagValue == 0 {
		return nil
	}

	td := tiles[center]
	if td.Flags&flagValue != 0 {
		return nil
	}

	td.Flags |= flagValue
	tiles[center] = td

	return []TileChange{TileChangeFromData(center, td)}
}

func FindGroundBrushByID(reg *brushes.Registry, groundID int) *brushes.Brush {
	return findGroundBrush(reg, groundID)
}

func findWallDecoBrushByName(reg *brushes.Registry, name string) *brushes.Brush {
	lower := strings.ToLower(name)
	for i := range reg.WallDecos {
		if strings.ToLower(reg.WallDecos[i].Name) == lower {
			return &reg.WallDecos[i]
		}
	}
	return nil
}

func ApplyWallDecoBrush(tiles map[Pos]TileData, center Pos, brushName string, reg *brushes.Registry) []TileChange {
	decoBrush := findWallDecoBrushByName(reg, brushName)
	if decoBrush == nil {
		return nil
	}

	td := tiles[center]

	var wallAlignment int = -1
	for _, itemID := range td.Items {
		for i := range reg.Walls {
			if a := getWallAlignmentFromBrush(&reg.Walls[i], int(itemID)); a >= 0 {
				wallAlignment = a
				break
			}
		}
		if wallAlignment >= 0 {
			break
		}
	}

	if wallAlignment < 0 {
		wallAlignment = WallHorizontal
	}

	decoID := pickWallItem(decoBrush, wallAlignment)
	if decoID == 0 {
		for bt := 0; bt <= WallUntouchable; bt++ {
			decoID = pickWallItem(decoBrush, bt)
			if decoID > 0 {
				break
			}
		}
	}
	if decoID == 0 {
		return nil
	}

	td.Items = removeWallDecoItems(td.Items, decoBrush)
	td.Items = append(td.Items, uint16(decoID))
	tiles[center] = td

	return []TileChange{TileChangeFromData(center, td)}
}

func getWallAlignmentFromBrush(brush *brushes.Brush, itemID int) int {
	for bt := 0; bt <= WallUntouchable; bt++ {
		typeName := wallTypeToString(bt)
		if typeName == "" {
			continue
		}
		for _, ws := range brush.Walls {
			if strings.EqualFold(ws.Type, typeName) {
				for _, it := range ws.Items {
					if it.ID == itemID {
						return bt
					}
				}
			}
		}
	}
	return -1
}

func removeWallDecoItems(items []uint16, decoBrush *brushes.Brush) []uint16 {
	out := items[:0]
	for _, id := range items {
		if !isWallItem(decoBrush, int(id)) {
			out = append(out, id)
		}
	}
	return out
}

const OptionalBorderFlag = 1 << 16

func ApplyOptionalBorderBrush(tiles map[Pos]TileData, center Pos, reg *brushes.Registry) []TileChange {
	td := tiles[center]
	td.Flags |= OptionalBorderFlag
	tiles[center] = td

	affected := make(map[Pos]bool)
	affected[center] = true
	for _, off := range neighborOffsets {
		p := Pos{center[0] + off[0], center[1] + off[1], center[2]}
		affected[p] = true
	}

	var changes []TileChange
	for pos := range affected {
		ptd := tiles[pos]
		borderItems := ComputeBordersForTile(pos, tiles, reg)
		optItems := computeOptionalBordersForTile(pos, tiles, reg)
		ptd.Items = append(borderItems, optItems...)
		tiles[pos] = ptd
		changes = append(changes, TileChangeFromData(pos, ptd))
	}
	return changes
}

func computeOptionalBordersForTile(pos Pos, tiles map[Pos]TileData, reg *brushes.Registry) []uint16 {
	td := tiles[pos]
	if td.Flags&OptionalBorderFlag == 0 {
		return nil
	}

	centerBrush := findGroundBrush(reg, td.GroundID)
	if centerBrush == nil {
		return nil
	}

	for _, br := range centerBrush.Borders {
		if br.Align == "outer" && br.To == "optional" && br.ID > 0 {
			border := findBorderDef(reg, br.ID)
			if border == nil {
				continue
			}
			em := borderToEdgeMap(border)

			var tiledata uint32
			for i, off := range neighborOffsets {
				np := Pos{pos[0] + off[0], pos[1] + off[1], pos[2]}
				if ntd, ok := tiles[np]; ok {
					if ntd.Flags&OptionalBorderFlag == 0 {
						nb := findGroundBrush(reg, ntd.GroundID)
						if nb == nil || nb.Name != centerBrush.Name {
							tiledata |= 1 << uint(i)
						}
					}
				} else {
					tiledata |= 1 << uint(i)
				}
			}

			if tiledata == 0 {
				return nil
			}

			dirs := unpackDirections(groundBorderTypes[tiledata])
			var items []uint16
			for _, dir := range dirs {
				if dir == BorderNone {
					break
				}
				if itemID, ok := em[dir]; ok && itemID > 0 {
					items = append(items, uint16(itemID))
				}
			}
			return items
		}
	}

	return nil
}
