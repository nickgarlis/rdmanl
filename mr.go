//go:build linux

package rdmanl

import "fmt"

// An MR is a memory region.
type MR struct {
	ID uint32
	// Length in bytes
	Length uint64
	// Keys, for privileged connections (CAP_NET_ADMIN) only
	RKey *uint32
	LKey *uint32
	// PD of a user MR
	PDN *uint32
	// Process of a user MR, or the module of a kernel MR
	PID      uint32
	KernName string
	Driver   []Driver
}

// An MRReq asks for the MRs of a device, or for one MR by its number:
// MRs uses DriverDetails, MR uses MRN.
type MRReq struct {
	DevIndex uint32
	MRN      uint32
	// Whether to include MRs only drivers use
	DriverDetails bool
}

func (r MRReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrResMRN, r.MRN)
	if r.DriverDetails {
		ae.Uint8(attrDriverDetails, 1)
	}
	return ae.Encode()
}

// MRs returns the MRs of a device, like "rdma resource show mr".
func (c *Conn) MRs(req MRReq) ([]MR, error) {
	var l mrList
	if err := c.dump(cmdResMRGet, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: list MRs of device %d: %w", req.DevIndex, err)
	}

	return l, nil
}

// MR returns an MR of a device.
func (c *Conn) MR(req MRReq) (MR, error) {
	var mr MR
	if err := c.get(cmdResMRGet, req, &mr); err != nil {
		return mr, fmt.Errorf("rdmanl: get MR %d of device %d: %w", req.MRN, req.DevIndex, err)
	}

	return mr, nil
}

// RawMRs returns the MRs of a device in the driver's format, for
// drivers that have one, such as mlx5.
func (c *Conn) RawMRs(req MRReq) ([][]byte, error) {
	l := rawList{attr: attrResMR, entry: attrResMREntry}
	if err := c.dump(cmdResMRGetRaw, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: list raw MRs of device %d: %w", req.DevIndex, err)
	}

	return l.bytes(), nil
}

// RawMR returns an MR of a device in the driver's format.
func (c *Conn) RawMR(req MRReq) ([]byte, error) {
	var r raw
	if err := c.get(cmdResMRGetRaw, req, &r); err != nil {
		return nil, fmt.Errorf("rdmanl: get raw MR %d of device %d: %w", req.MRN, req.DevIndex, err)
	}

	return r, nil
}

// A mrList is the MRs of a dump.
type mrList []MR

func (l *mrList) unmarshal(data []byte) error {
	return unmarshalTable(data, attrResMR, attrResMREntry, (*[]MR)(l))
}

func (mr *MR) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrResMRN:
			mr.ID = ad.Uint32()
		case attrResMRLen:
			mr.Length = ad.Uint64()
		case attrResRKey:
			mr.RKey = new(ad.Uint32())
		case attrResLKey:
			mr.LKey = new(ad.Uint32())
		case attrResPDN:
			mr.PDN = new(ad.Uint32())
		case attrResPID:
			mr.PID = ad.Uint32()
		case attrResKernName:
			mr.KernName = ad.String()
		case attrDriver:
			if err := (*drivers)(&mr.Driver).unmarshal(ad.Bytes()); err != nil {
				return err
			}
		}
	}
	return ad.Err()
}
