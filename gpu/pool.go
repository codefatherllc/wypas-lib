package gpu

import (
	"bytes"
	"sync"
)

var (
	BufPool = sync.Pool{New: func() any { return &bytes.Buffer{} }}
	PixPool = sync.Pool{New: func() any { return make([]byte, 0, 256*256*4) }}
)

func GetBuf() *bytes.Buffer {
	b := BufPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

func PutBuf(b *bytes.Buffer) {
	b.Reset()
	BufPool.Put(b)
}

func GetPix(n int) []byte {
	p := PixPool.Get().([]byte)
	if cap(p) < n {
		return make([]byte, n)
	}
	return p[:n]
}

func PutPix(p []byte) {
	PixPool.Put(p[:0])
}
