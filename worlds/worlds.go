package worlds

import (
	"encoding/json"
	"fmt"
	"os"
)

type World struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    int    `json:"port"`
	Region  string `json:"region,omitempty"`
	PvPType string `json:"pvp_type,omitempty"`
}

type WorldList struct {
	worlds []World
	byID   map[int]World
}

func Load(path string) (*WorldList, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read worlds config: %w", err)
	}
	var worlds []World
	if err := json.Unmarshal(data, &worlds); err != nil {
		return nil, fmt.Errorf("parse worlds config: %w", err)
	}
	if len(worlds) == 0 {
		return nil, fmt.Errorf("worlds config: no worlds defined")
	}
	byID := make(map[int]World, len(worlds))
	for _, w := range worlds {
		if _, dup := byID[w.ID]; dup {
			return nil, fmt.Errorf("worlds config: duplicate world id %d", w.ID)
		}
		byID[w.ID] = w
	}
	return &WorldList{worlds: worlds, byID: byID}, nil
}

func NewFromSlice(ww []World) *WorldList {
	byID := make(map[int]World, len(ww))
	for _, w := range ww {
		byID[w.ID] = w
	}
	return &WorldList{worlds: ww, byID: byID}
}

func (wl *WorldList) All() []World              { return wl.worlds }
func (wl *WorldList) ByID(id int) (World, bool) { w, ok := wl.byID[id]; return w, ok }
func (wl *WorldList) Default() World            { return wl.worlds[0] }

func ToWorldsMap(wl []World) map[int]string {
	m := make(map[int]string, len(wl))
	for _, w := range wl {
		m[w.ID] = w.Name
	}
	return m
}
