package gpu

type Renderer interface {
	Composite(dstW, dstH int, ops []BlitOp) ([]byte, error)
	TintOutfit(p OutfitTintParams) ([]byte, error)
	ResizeNN(src []byte, srcW, srcH, dstW, dstH int) ([]byte, error)
	ResizeCatmullRom(src []byte, srcW, srcH, dstW, dstH int) ([]byte, error)
	RenderHeatmap(p HeatmapParams) ([]byte, error)
	ApplyShadow(pixels []byte, w, h int, intensity float64) ([]byte, error)
	RenderMinimap(w, h int, tiles []MinimapTile) ([]byte, error)
	Close()
}

func New() Renderer {
	r, err := newMetal()
	if err == nil {
		return r
	}
	return NewCPU()
}
