package otbm

import (
	"fmt"
)

type OTB struct {
	ServerToClient map[uint16]uint16
}

func ParseOTB(path string) (*OTB, error) {
	root, err := ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("parse otb tree: %w", err)
	}

	root.ResetPos()
	if _, err := root.GetU32(); err != nil {
		return nil, fmt.Errorf("read root signature: %w", err)
	}
	if root.Remaining() > 0 {
		attr, _ := root.GetU8()
		if attr == 0x01 {
			size, err := root.GetU16()
			if err != nil {
				return nil, fmt.Errorf("read version info size: %w", err)
			}
			if err := root.Skip(int(size)); err != nil {
				return nil, fmt.Errorf("skip version info: %w", err)
			}
		}
	}

	otb := &OTB{
		ServerToClient: make(map[uint16]uint16, len(root.Children)),
	}

	for _, child := range root.Children {
		child.ResetPos()

		if _, err := child.GetU32(); err != nil {
			continue
		}

		var serverID, clientID uint16
		var hasServer, hasClient bool

		for child.Remaining() > 0 {
			attrType, err := child.GetU8()
			if err != nil || attrType == 0 || attrType == 0xFF {
				break
			}
			attrLen, err := child.GetU16()
			if err != nil {
				break
			}

			switch attrType {
			case 0x10: // ServerId
				if attrLen >= 2 {
					serverID, _ = child.GetU16()
					hasServer = true
					if attrLen > 2 {
						child.Skip(int(attrLen) - 2)
					}
				} else {
					child.Skip(int(attrLen))
				}
			case 0x11: // ClientId
				if attrLen >= 2 {
					clientID, _ = child.GetU16()
					hasClient = true
					if attrLen > 2 {
						child.Skip(int(attrLen) - 2)
					}
				} else {
					child.Skip(int(attrLen))
				}
			default:
				child.Skip(int(attrLen))
			}
		}

		if hasServer && hasClient {
			otb.ServerToClient[serverID] = clientID
		}
	}

	return otb, nil
}
