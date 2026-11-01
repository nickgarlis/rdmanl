//go:build linux

package rdmanl

import "fmt"

// A PD is a protection domain.
type PD struct {
	ID uint32
	// Number of objects in it
	Users uint64
	// Keys, for privileged connections (CAP_NET_ADMIN) only
	LocalDMALKey     *uint32
	UnsafeGlobalRKey *uint32
	// User context of a user PD
	CtxN *uint32
	// Process of a user PD, or the module of a kernel PD
	PID      uint32
	KernName string
	Driver   []Driver
}

// A PDReq asks for the PDs of a device, or for one PD by its number:
// PDs uses DriverDetails, PD uses PDN.
type PDReq struct {
	DevIndex uint32
	PDN      uint32
	// Whether to include PDs only drivers use
	DriverDetails bool
}

func (r PDReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrResPDN, r.PDN)
	if r.DriverDetails {
		ae.Uint8(attrDriverDetails, 1)
	}
	return ae.Encode()
}

// PDs returns the PDs of a device, like "rdma resource show pd".
func (c *Conn) PDs(req PDReq) ([]PD, error) {
	var l pdList
	if err := c.dump(cmdResPDGet, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: list PDs of device %d: %w", req.DevIndex, err)
	}

	return l, nil
}

// PD returns a PD of a device.
func (c *Conn) PD(req PDReq) (PD, error) {
	var pd PD
	if err := c.get(cmdResPDGet, req, &pd); err != nil {
		return pd, fmt.Errorf("rdmanl: get PD %d of device %d: %w", req.PDN, req.DevIndex, err)
	}

	return pd, nil
}

// A pdList is the PDs of a dump.
type pdList []PD

func (l *pdList) unmarshal(data []byte) error {
	return unmarshalTable(data, attrResPD, attrResPDEntry, (*[]PD)(l))
}

func (pd *PD) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrResPDN:
			pd.ID = ad.Uint32()
		case attrResUseCnt:
			pd.Users = ad.Uint64()
		case attrResLocalDMALKey:
			pd.LocalDMALKey = new(ad.Uint32())
		case attrResUnsafeGlobalRKey:
			pd.UnsafeGlobalRKey = new(ad.Uint32())
		case attrResCtxN:
			pd.CtxN = new(ad.Uint32())
		case attrResPID:
			pd.PID = ad.Uint32()
		case attrResKernName:
			pd.KernName = ad.String()
		case attrDriver:
			if err := (*drivers)(&pd.Driver).unmarshal(ad.Bytes()); err != nil {
				return err
			}
		}
	}
	return ad.Err()
}
