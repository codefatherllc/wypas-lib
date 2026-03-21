package otbm

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	nodeStart = 0xFE
	nodeEnd   = 0xFF
	escape    = 0xFD
)

type Node struct {
	Type     uint8
	Data     []byte
	Children []*Node
	pos      int
}

func ParseFile(path string) (*Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	if len(data) < 4 {
		return nil, fmt.Errorf("file too small: %d bytes", len(data))
	}
	buf := data[4:]
	if len(buf) == 0 || buf[0] != nodeStart {
		return nil, fmt.Errorf("expected node start (0xFE), got 0x%02X", buf[0])
	}
	node, _, err := parseNode(buf, 1)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func parseNode(buf []byte, pos int) (*Node, int, error) {
	if pos >= len(buf) {
		return nil, pos, fmt.Errorf("unexpected end of data at pos %d", pos)
	}
	node := &Node{Type: buf[pos]}
	pos++

	for pos < len(buf) {
		b := buf[pos]
		switch b {
		case nodeStart:
			pos++
			child, newPos, err := parseNode(buf, pos)
			if err != nil {
				return nil, newPos, err
			}
			node.Children = append(node.Children, child)
			pos = newPos
		case nodeEnd:
			return node, pos + 1, nil
		case escape:
			pos++
			if pos >= len(buf) {
				return nil, pos, fmt.Errorf("escape at end of data")
			}
			node.Data = append(node.Data, buf[pos])
			pos++
		default:
			node.Data = append(node.Data, b)
			pos++
		}
	}
	return nil, pos, fmt.Errorf("unexpected end of data, missing node end")
}

func (n *Node) GetU8() (uint8, error) {
	if n.pos >= len(n.Data) {
		return 0, fmt.Errorf("read u8: past end of data (pos=%d, len=%d)", n.pos, len(n.Data))
	}
	v := n.Data[n.pos]
	n.pos++
	return v, nil
}

func (n *Node) GetU16() (uint16, error) {
	if n.pos+2 > len(n.Data) {
		return 0, fmt.Errorf("read u16: need 2 bytes at pos %d, have %d", n.pos, len(n.Data))
	}
	v := binary.LittleEndian.Uint16(n.Data[n.pos:])
	n.pos += 2
	return v, nil
}

func (n *Node) GetU32() (uint32, error) {
	if n.pos+4 > len(n.Data) {
		return 0, fmt.Errorf("read u32: need 4 bytes at pos %d, have %d", n.pos, len(n.Data))
	}
	v := binary.LittleEndian.Uint32(n.Data[n.pos:])
	n.pos += 4
	return v, nil
}

func (n *Node) GetString() (string, error) {
	length, err := n.GetU16()
	if err != nil {
		return "", fmt.Errorf("read string length: %w", err)
	}
	if n.pos+int(length) > len(n.Data) {
		return "", fmt.Errorf("read string: need %d bytes at pos %d, have %d", length, n.pos, len(n.Data))
	}
	s := string(n.Data[n.pos : n.pos+int(length)])
	n.pos += int(length)
	return s, nil
}

func (n *Node) Skip(count int) error {
	if n.pos+count > len(n.Data) {
		return fmt.Errorf("skip %d: past end of data (pos=%d, len=%d)", count, n.pos, len(n.Data))
	}
	n.pos += count
	return nil
}

func (n *Node) Remaining() int {
	return len(n.Data) - n.pos
}

func (n *Node) ResetPos() {
	n.pos = 0
}
