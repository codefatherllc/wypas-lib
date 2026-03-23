package gpu

type BlitOp struct {
	SrcPixels  []byte
	SrcW, SrcH int
	DstX, DstY int
}

type OutfitTintParams struct {
	Base, Overlay          []byte
	Head, Body, Legs, Feet [3]uint8
	IsAddon                bool
}

type HeatmapParams struct {
	Freq       map[[2]int]int
	ImgW, ImgH int
	Radius     int
}

type MinimapTile struct {
	X, Y       int
	ColorIndex uint8
}
