package taxonomy

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/codefatherllc/wypas-lib/gamedata"
)

type BuildTaxonomy struct {
	GroundGroups map[uint16]*SemanticGroup
	DecoItems    map[uint16]int
	GroundIDs    map[uint16]int
	BorderIDs    map[uint16]int
	RoleMap      map[uint16]Role
	AllGroups    []*SemanticGroup
}

func BuildFromItems(items []gamedata.ItemType) *BuildTaxonomy {
	roleMap := make(map[uint16]Role, len(items))
	for i := range items {
		roleMap[items[i].ServerID] = ClassifyItem(&items[i])
	}

	allGroups := GroupByMinimapColor(items, roleMap)

	groundGroups := make(map[uint16]*SemanticGroup)

	for color, g := range allGroups {
		if g.Role == GROUND {
			groundGroups[color] = g
		}
	}

	decoItems := make(map[uint16]int)
	groundIDs := make(map[uint16]int)
	borderIDs := make(map[uint16]int)
	decoIdx, groundIdx, borderIdx := 0, 1, 1
	for i := range items {
		sid := items[i].ServerID
		switch roleMap[sid] {
		case DECORATION:
			decoItems[sid] = decoIdx + 2
			decoIdx++
		case GROUND:
			groundIDs[sid] = groundIdx
			groundIdx++
		case GROUND_BORDER:
			borderIDs[sid] = borderIdx
			borderIdx++
		}
	}

	grouped := make([]*SemanticGroup, 0, len(allGroups))
	for _, g := range allGroups {
		grouped = append(grouped, g)
	}

	return &BuildTaxonomy{
		GroundGroups: groundGroups,
		DecoItems:    decoItems,
		GroundIDs:    groundIDs,
		BorderIDs:    borderIDs,
		RoleMap:      roleMap,
		AllGroups:    grouped,
	}
}

type Taxonomy struct {
	GroundGroups    map[string]*SemanticGroup `json:"ground_groups"`
	WallGroups      map[string]*SemanticGroup `json:"wall_groups,omitempty"`
	DecoVocab       map[string]int            `json:"deco_vocab"`
	RoleMap         map[string]string         `json:"role_map,omitempty"`
	AdjacencyRules  []AdjacencyRule           `json:"adjacency_rules"`
	WallPatterns    []WallPattern             `json:"wall_patterns"`
	MonsterAffinity []MonsterAffinity         `json:"monster_affinity"`
	WFCAdjacency    *WFCAdjacencyData         `json:"wfc_adjacency,omitempty"`

	NumGroundGroups int `json:"num_ground_groups"`
	NumWallGroups   int `json:"num_wall_groups,omitempty"`
	NumBorderItems  int `json:"num_border_items"`
	NumDecoItems    int `json:"num_deco_items"`
	NumGroundIDs    int `json:"num_ground_ids"`

	groundByIndex   map[int]*SemanticGroup
	wallByIndex     map[int]*SemanticGroup
	itemToGroup     map[uint16]int
	itemToRole      map[uint16]string
	decoByIndex     map[int]uint16
	affinityByGroup map[int][]MonsterAffinity
}

func LoadTaxonomy(dir string) (*Taxonomy, error) {
	path := filepath.Join(dir, "taxonomy.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read taxonomy: %w", err)
	}
	var tax Taxonomy
	if err := json.Unmarshal(data, &tax); err != nil {
		return nil, fmt.Errorf("parse taxonomy: %w", err)
	}
	tax.init()
	return &tax, nil
}

func (t *Taxonomy) init() {
	t.groundByIndex = make(map[int]*SemanticGroup)
	t.wallByIndex = make(map[int]*SemanticGroup)
	t.itemToGroup = make(map[uint16]int)
	t.itemToRole = make(map[uint16]string)
	t.decoByIndex = make(map[int]uint16)

	groundIDs := make(map[uint16]struct{})
	for _, g := range t.GroundGroups {
		g.Role = GROUND
		t.groundByIndex[g.Index] = g
		for _, id := range g.Items {
			t.itemToGroup[id] = g.Index
			t.itemToRole[id] = "ground"
			groundIDs[id] = struct{}{}
		}
	}
	if t.NumGroundGroups == 0 {
		t.NumGroundGroups = len(t.GroundGroups)
	}
	t.NumGroundIDs = len(groundIDs)

	for _, g := range t.WallGroups {
		g.Role = WALL
		t.wallByIndex[g.Index] = g
		for _, id := range g.Items {
			t.itemToGroup[id] = g.Index
			t.itemToRole[id] = "wall"
		}
	}
	if t.NumWallGroups == 0 {
		t.NumWallGroups = len(t.WallGroups)
	}

	if t.NumBorderItems == 0 {
		t.NumBorderItems = 0
	}

	if t.NumDecoItems == 0 {
		t.NumDecoItems = len(t.DecoVocab)
	}
	for key, idx := range t.DecoVocab {
		var sid uint16
		fmt.Sscanf(key, "%d", &sid)
		t.decoByIndex[idx] = sid
	}

	// Load role_map from JSON (written by scrapper) — overrides inferred roles
	if len(t.RoleMap) > 0 {
		for key, role := range t.RoleMap {
			var sid uint16
			fmt.Sscanf(key, "%d", &sid)
			t.itemToRole[sid] = role
		}
	}

	t.affinityByGroup = make(map[int][]MonsterAffinity)
	for _, ma := range t.MonsterAffinity {
		for _, g := range ma.GroundGroups {
			t.affinityByGroup[g] = append(t.affinityByGroup[g], ma)
		}
	}
}

func (t *Taxonomy) GroupForItem(serverID uint16) (int, bool) {
	idx, ok := t.itemToGroup[serverID]
	return idx, ok
}

func (t *Taxonomy) GroupByIndex(index int) *SemanticGroup {
	if g, ok := t.groundByIndex[index]; ok {
		return g
	}
	return t.wallByIndex[index]
}

func (t *Taxonomy) GroupName(index int) string {
	g := t.GroupByIndex(index)
	if g == nil {
		return ""
	}
	return g.Name
}

func (t *Taxonomy) AdjacencyBorders(fromGroup, toGroup int) []uint16 {
	for _, r := range t.AdjacencyRules {
		if r.FromGroup == fromGroup && r.ToGroup == toGroup {
			return r.BorderItems
		}
	}
	return nil
}

func (t *Taxonomy) WallItem(wallGroup int, neighborPattern string) uint16 {
	for _, p := range t.WallPatterns {
		if p.WallGroup == wallGroup && p.Neighbors == neighborPattern {
			return p.ItemID
		}
	}
	return 0
}

func (t *Taxonomy) DecoIndex(serverID uint16) int {
	key := fmt.Sprintf("%d", serverID)
	if idx, ok := t.DecoVocab[key]; ok {
		return idx
	}
	return 0
}

func (t *Taxonomy) DecoItemID(vocabIndex int) uint16 {
	return t.decoByIndex[vocabIndex]
}

func (t *Taxonomy) MinimapColorForGroup(index int) uint16 {
	g := t.GroupByIndex(index)
	if g == nil {
		return 0
	}
	return g.MinimapColor
}

func (t *Taxonomy) IsGround(serverID uint16) bool {
	return t.itemToRole[serverID] == "ground"
}

func (t *Taxonomy) IsWall(serverID uint16) bool {
	return t.itemToRole[serverID] == "wall"
}

func (t *Taxonomy) IsBorder(serverID uint16) bool {
	return t.itemToRole[serverID] == "border"
}

func (t *Taxonomy) IsBlocking(serverID uint16) bool {
	return t.itemToRole[serverID] == "blocking"
}

func (t *Taxonomy) IsDecoration(serverID uint16) bool {
	return t.itemToRole[serverID] == "decoration"
}

func (t *Taxonomy) RoleOf(serverID uint16) string {
	return t.itemToRole[serverID]
}

func (t *Taxonomy) AffinityForGroundGroup(groupIndex int) []MonsterAffinity {
	return t.affinityByGroup[groupIndex]
}

func (t *Taxonomy) AllAffinities() []MonsterAffinity {
	return t.MonsterAffinity
}

func (t *Taxonomy) RandomGroundForZone(zoneIndex int) uint16 {
	g, ok := t.groundByIndex[zoneIndex]
	if !ok || len(g.Items) == 0 {
		return 0
	}
	return g.Items[rand.Intn(len(g.Items))]
}

func (t *Taxonomy) GroupItems(groupIndex int) []uint16 {
	if g, ok := t.groundByIndex[groupIndex]; ok {
		return g.Items
	}
	if g, ok := t.wallByIndex[groupIndex]; ok {
		return g.Items
	}
	return nil
}
