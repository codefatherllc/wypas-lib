//go:build darwin

package gpu

import (
	_ "embed"
	"errors"
	"unsafe"
)

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Metal -framework CoreGraphics -framework Foundation
#include "bridge/bridge.h"
#include <stdlib.h>
*/
import "C"

//go:embed bridge/shaders.metal
var metalShaderSource string

type metalRenderer struct {
	gpu *C.MetalGPU
}

func NewMetal() (Renderer, error) {
	return newMetal()
}

func newMetal() (Renderer, error) {
	cs := C.CString(metalShaderSource)
	defer C.free(unsafe.Pointer(cs))

	gpu := C.MetalGPU_Init(cs, C.int(len(metalShaderSource)))
	if gpu == nil {
		return nil, errors.New("gpu: failed to initialize Metal device")
	}

	return &metalRenderer{gpu: gpu}, nil
}

func (m *metalRenderer) Close() {
	if m.gpu != nil {
		C.MetalGPU_Destroy(m.gpu)
		m.gpu = nil
	}
}

func (m *metalRenderer) Composite(dstW, dstH int, ops []BlitOp) ([]byte, error) {
	n := dstW * dstH * 4
	dst := GetPix(n)
	for i := range dst {
		dst[i] = 0
	}

	if len(ops) == 0 {
		return dst, nil
	}

	totalSrcBytes := 0
	for _, op := range ops {
		totalSrcBytes += len(op.SrcPixels)
	}
	srcData := make([]byte, 0, totalSrcBytes)
	descriptors := make([]C.int, 0, len(ops)*5)

	offset := 0
	for _, op := range ops {
		srcData = append(srcData, op.SrcPixels...)
		descriptors = append(descriptors,
			C.int(offset), C.int(op.SrcW), C.int(op.SrcH),
			C.int(op.DstX), C.int(op.DstY))
		offset += len(op.SrcPixels)
	}

	rc := C.MetalGPU_Composite(m.gpu,
		(*C.uchar)(unsafe.Pointer(&dst[0])), C.int(dstW), C.int(dstH),
		(*C.uchar)(unsafe.Pointer(&srcData[0])),
		(*C.int)(unsafe.Pointer(&descriptors[0])), C.int(len(ops)))
	if rc != 0 {
		return nil, errors.New("gpu: Metal composite failed")
	}

	return dst, nil
}

func (m *metalRenderer) TintOutfit(p OutfitTintParams) ([]byte, error) {
	const size = 32
	n := size * size * 4
	out := GetPix(n)

	var overlayPtr *C.uchar
	if p.Overlay != nil {
		overlayPtr = (*C.uchar)(unsafe.Pointer(&p.Overlay[0]))
	}

	isAddon := C.int(0)
	if p.IsAddon {
		isAddon = 1
	}

	head := [3]C.uchar{C.uchar(p.Head[0]), C.uchar(p.Head[1]), C.uchar(p.Head[2])}
	body := [3]C.uchar{C.uchar(p.Body[0]), C.uchar(p.Body[1]), C.uchar(p.Body[2])}
	legs := [3]C.uchar{C.uchar(p.Legs[0]), C.uchar(p.Legs[1]), C.uchar(p.Legs[2])}
	feet := [3]C.uchar{C.uchar(p.Feet[0]), C.uchar(p.Feet[1]), C.uchar(p.Feet[2])}

	rc := C.MetalGPU_TintOutfit(m.gpu,
		(*C.uchar)(unsafe.Pointer(&p.Base[0])), overlayPtr,
		C.int(size), C.int(size),
		&head[0], &body[0], &legs[0], &feet[0],
		isAddon, (*C.uchar)(unsafe.Pointer(&out[0])))
	if rc != 0 {
		return nil, errors.New("gpu: Metal outfit tint failed")
	}

	return out, nil
}

func (m *metalRenderer) ResizeNN(src []byte, srcW, srcH, dstW, dstH int) ([]byte, error) {
	n := dstW * dstH * 4
	out := GetPix(n)

	rc := C.MetalGPU_ResizeNN(m.gpu,
		(*C.uchar)(unsafe.Pointer(&src[0])), C.int(srcW), C.int(srcH),
		C.int(dstW), C.int(dstH), (*C.uchar)(unsafe.Pointer(&out[0])))
	if rc != 0 {
		return nil, errors.New("gpu: Metal resize NN failed")
	}

	return out, nil
}

func (m *metalRenderer) ResizeCatmullRom(src []byte, srcW, srcH, dstW, dstH int) ([]byte, error) {
	n := dstW * dstH * 4
	out := GetPix(n)

	rc := C.MetalGPU_ResizeCatmullRom(m.gpu,
		(*C.uchar)(unsafe.Pointer(&src[0])), C.int(srcW), C.int(srcH),
		C.int(dstW), C.int(dstH), (*C.uchar)(unsafe.Pointer(&out[0])))
	if rc != 0 {
		return nil, errors.New("gpu: Metal resize CatmullRom failed")
	}

	return out, nil
}

func (m *metalRenderer) RenderHeatmap(p HeatmapParams) ([]byte, error) {
	n := p.ImgW * p.ImgH * 4
	out := GetPix(n)

	if len(p.Freq) == 0 {
		return out, nil
	}

	freq := make([]C.float, p.ImgW*p.ImgH)
	for coord, count := range p.Freq {
		idx := coord[1]*p.ImgW + coord[0]
		if idx >= 0 && idx < len(freq) {
			freq[idx] = C.float(count)
		}
	}

	rc := C.MetalGPU_RenderHeatmap(m.gpu,
		(*C.float)(unsafe.Pointer(&freq[0])),
		C.int(p.ImgW), C.int(p.ImgH), C.int(p.Radius),
		(*C.uchar)(unsafe.Pointer(&out[0])))
	if rc != 0 {
		return nil, errors.New("gpu: Metal heatmap failed")
	}

	return out, nil
}

func (m *metalRenderer) ApplyShadow(pixels []byte, w, h int, intensity float64) ([]byte, error) {
	n := w * h * 4
	out := GetPix(n)
	copy(out, pixels)

	rc := C.MetalGPU_ApplyShadow(m.gpu,
		(*C.uchar)(unsafe.Pointer(&out[0])), C.int(w), C.int(h), C.float(intensity))
	if rc != 0 {
		return nil, errors.New("gpu: Metal shadow failed")
	}

	return out, nil
}

func (m *metalRenderer) RenderMinimap(w, h int, tiles []MinimapTile) ([]byte, error) {
	n := w * h * 4
	out := GetPix(n)
	for i := range out {
		out[i] = 0
	}

	if len(tiles) == 0 {
		return out, nil
	}

	tileData := make([]C.int, len(tiles)*3)
	for i, t := range tiles {
		tileData[i*3+0] = C.int(t.X)
		tileData[i*3+1] = C.int(t.Y)
		tileData[i*3+2] = C.int(t.ColorIndex)
	}

	rc := C.MetalGPU_RenderMinimap(m.gpu, C.int(w), C.int(h),
		(*C.int)(unsafe.Pointer(&tileData[0])), C.int(len(tiles)),
		(*C.uchar)(unsafe.Pointer(&out[0])))
	if rc != 0 {
		return nil, errors.New("gpu: Metal minimap failed")
	}

	return out, nil
}

var _ Renderer = (*metalRenderer)(nil)
