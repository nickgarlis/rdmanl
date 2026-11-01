//go:build linux

package rdmanl

import "fmt"

// A Port is a port of a device.
type Port struct {
	DevIndex uint32
	DevName  string
	// Port number, from 1
	Index uint32
	// Logical and physical state, enum ib_port_state and
	// ib_port_phys_state
	State     uint8
	PhysState uint8
	// Netdev of a RoCE or iWARP port, if it has one
	NdevIndex uint32
	NdevName  string
	// InfiniBand ports only: capabilities (CapabilityMask, then
	// CapabilityMask2 above bit 32), subnet prefix in host byte order, LIDs
	// and LID mask control
	CapFlags     uint64
	SubnetPrefix uint64
	LID          uint32
	SMLID        uint32
	LMC          uint8
}

// A PortReq asks for the ports of a device, or for one of them: Ports
// lists them all, Port uses PortIndex.
type PortReq struct {
	DevIndex  uint32
	PortIndex uint32
}

func (r PortReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	if r.PortIndex != 0 {
		ae.Uint32(attrPortIndex, r.PortIndex)
	}
	return ae.Encode()
}

// Ports returns the ports of a device.
func (c *Conn) Ports(req PortReq) ([]Port, error) {
	var ports portList
	if err := c.dump(cmdPortGet, req, &ports); err != nil {
		return nil, fmt.Errorf("rdmanl: list ports of device %d: %w", req.DevIndex, err)
	}

	return ports, nil
}

// Port returns a port of a device.
func (c *Conn) Port(req PortReq) (Port, error) {
	var p Port
	if err := c.get(cmdPortGet, req, &p); err != nil {
		return p, fmt.Errorf("rdmanl: get port %d of device %d: %w", req.PortIndex, req.DevIndex, err)
	}

	return p, nil
}

// A portList is the ports of a dump, one per reply.
type portList []Port

func (l *portList) unmarshal(data []byte) error {
	var p Port
	if err := p.unmarshal(data); err != nil {
		return err
	}

	*l = append(*l, p)
	return nil
}

func (p *Port) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrDevIndex:
			p.DevIndex = ad.Uint32()
		case attrDevName:
			p.DevName = ad.String()
		case attrPortIndex:
			p.Index = ad.Uint32()
		case attrPortState:
			p.State = ad.Uint8()
		case attrPortPhysState:
			p.PhysState = ad.Uint8()
		case attrNdevIndex:
			p.NdevIndex = ad.Uint32()
		case attrNdevName:
			p.NdevName = ad.String()
		case attrCapFlags:
			p.CapFlags = ad.Uint64()
		case attrSubnetPrefix:
			p.SubnetPrefix = ad.Uint64()
		case attrLID:
			p.LID = ad.Uint32()
		case attrSMLID:
			p.SMLID = ad.Uint32()
		case attrLMC:
			p.LMC = ad.Uint8()
		}
	}
	return ad.Err()
}
