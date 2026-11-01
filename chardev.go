//go:build linux

package rdmanl

import "fmt"

// A Chardev is a character device of a device, such as its uverbs device.
type Chardev struct {
	// Name under /dev/infiniband, such as "uverbs0"
	Name string
	// ABI revision
	ABI uint64
	// Device number (dev_t)
	Dev uint64
	// Driver of the device, enum rdma_driver_id; only for "uverbs"
	DriverID uint32
}

// A ChardevReq asks for the character device of the type ChardevType,
// such as "uverbs", "umad" or "rdma_cm": of the device DevIndex, of its
// port PortIndex, or of no device if DevIndex is nil.
type ChardevReq struct {
	DevIndex *uint32
	// The port, from 1, or 0 for the device itself
	PortIndex   uint32
	ChardevType string
}

func (r ChardevReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	if r.DevIndex != nil {
		ae.Uint32(attrDevIndex, *r.DevIndex)
	}
	if r.PortIndex != 0 {
		ae.Uint32(attrPortIndex, r.PortIndex)
	}
	ae.String(attrChardevType, r.ChardevType)
	return ae.Encode()
}

// Chardev returns a character device.
func (c *Conn) Chardev(req ChardevReq) (Chardev, error) {
	var cd Chardev
	if err := c.get(cmdGetChardev, req, &cd); err != nil {
		return cd, fmt.Errorf("rdmanl: get %s device: %w", req.ChardevType, err)
	}

	return cd, nil
}

func (cd *Chardev) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrChardevName:
			cd.Name = ad.String()
		case attrChardevABI:
			cd.ABI = ad.Uint64()
		case attrChardev:
			cd.Dev = ad.Uint64()
		case attrUverbsDriverID:
			cd.DriverID = ad.Uint32()
		}
	}
	return ad.Err()
}
