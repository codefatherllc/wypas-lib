package taxonomy

type SemanticGroup struct {
	Index        int      `json:"index"`
	Name         string   `json:"name"`
	MinimapColor uint16   `json:"minimap_color"`
	Items        []uint16 `json:"items"`
	Role         Role     `json:"-"`
}

type AdjacencyRule struct {
	FromGroup   int      `json:"from_group"`
	ToGroup     int      `json:"to_group"`
	BorderItems []uint16 `json:"border_items"`
	Count       int      `json:"count"`
}

type WallPattern struct {
	WallGroup int    `json:"wall_group"`
	Neighbors string `json:"neighbors"`
	ItemID    uint16 `json:"item_id"`
	Count     int    `json:"count"`
}

type MonsterAffinity struct {
	Family        string `json:"family"`
	GroundGroups  []int  `json:"ground_groups"`
	AvgExperience int    `json:"avg_experience"`
	AvgHealth     int    `json:"avg_health"`
	FloorType     string `json:"floor_type"`
}

type GroundPairData struct {
	FromID    int `json:"from_id"`
	ToID      int `json:"to_id"`
	Direction int `json:"direction"`
	Count     int `json:"count"`
}

type BorderSeqData struct {
	FromGroup int   `json:"from_group"`
	ToGroup   int   `json:"to_group"`
	BorderIDs []int `json:"border_ids"`
}

type WallConnData struct {
	FromGroup int `json:"from_wall_group"`
	ToGroup   int `json:"to_wall_group"`
	Direction int `json:"direction"`
}

type WFCAdjacencyData struct {
	GroundPairs []GroundPairData `json:"ground_pairs"`
	BorderSeqs  []BorderSeqData  `json:"border_seqs"`
	WallConns   []WallConnData   `json:"wall_conns"`
}
