//go:build !darwin

package gpu

import "errors"

func NewMetal() (Renderer, error) {
	return nil, errors.New("gpu: Metal not available on this platform")
}

func newMetal() (Renderer, error) {
	return NewMetal()
}
