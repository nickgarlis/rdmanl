//go:build linux

package rdmanl

import "fmt"

// A CQ is a completion queue.
type CQ struct {
	ID uint32
	// Number of entries
	CQE uint32
	// Number of QPs and other objects using it
	Users uint64
	// Whether it uses adaptive moderation (DIM)
	Dim bool
	// How a kernel CQ is polled, enum ib_poll_context
	PollCtx *uint8
	// User context of a user CQ
	CtxN *uint32
	// Process of a user CQ, or the module of a kernel CQ
	PID      uint32
	KernName string
	Driver   []Driver
}

// A CQReq asks for the CQs of a device, or for one CQ by its number:
// CQs uses DriverDetails, CQ uses CQN.
type CQReq struct {
	DevIndex uint32
	CQN      uint32
	// Whether to include CQs only drivers use
	DriverDetails bool
}

func (r CQReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrResCQN, r.CQN)
	if r.DriverDetails {
		ae.Uint8(attrDriverDetails, 1)
	}
	return ae.Encode()
}

// CQs returns the CQs of a device, like "rdma resource show cq".
func (c *Conn) CQs(req CQReq) ([]CQ, error) {
	var l cqList
	if err := c.dump(cmdResCQGet, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: list CQs of device %d: %w", req.DevIndex, err)
	}

	return l, nil
}

// CQ returns a CQ of a device.
func (c *Conn) CQ(req CQReq) (CQ, error) {
	var cq CQ
	if err := c.get(cmdResCQGet, req, &cq); err != nil {
		return cq, fmt.Errorf("rdmanl: get CQ %d of device %d: %w", req.CQN, req.DevIndex, err)
	}

	return cq, nil
}

// RawCQs returns the CQs of a device in the driver's format, for
// drivers that have one, such as mlx5.
func (c *Conn) RawCQs(req CQReq) ([][]byte, error) {
	l := rawList{attr: attrResCQ, entry: attrResCQEntry}
	if err := c.dump(cmdResCQGetRaw, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: list raw CQs of device %d: %w", req.DevIndex, err)
	}

	return l.bytes(), nil
}

// RawCQ returns a CQ of a device in the driver's format.
func (c *Conn) RawCQ(req CQReq) ([]byte, error) {
	var r raw
	if err := c.get(cmdResCQGetRaw, req, &r); err != nil {
		return nil, fmt.Errorf("rdmanl: get raw CQ %d of device %d: %w", req.CQN, req.DevIndex, err)
	}

	return r, nil
}

// A cqList is the CQs of a dump.
type cqList []CQ

func (l *cqList) unmarshal(data []byte) error {
	return unmarshalTable(data, attrResCQ, attrResCQEntry, (*[]CQ)(l))
}

func (cq *CQ) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrResCQN:
			cq.ID = ad.Uint32()
		case attrResCQE:
			cq.CQE = ad.Uint32()
		case attrResUseCnt:
			cq.Users = ad.Uint64()
		case attrDevDim:
			cq.Dim = ad.Uint8() != 0
		case attrResPollCtx:
			cq.PollCtx = new(ad.Uint8())
		case attrResCtxN:
			cq.CtxN = new(ad.Uint32())
		case attrResPID:
			cq.PID = ad.Uint32()
		case attrResKernName:
			cq.KernName = ad.String()
		case attrDriver:
			if err := (*drivers)(&cq.Driver).unmarshal(ad.Bytes()); err != nil {
				return err
			}
		}
	}
	return ad.Err()
}
