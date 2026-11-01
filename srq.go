//go:build linux

package rdmanl

import "fmt"

// An SRQ is a shared receive queue.
type SRQ struct {
	ID uint32
	// Type, enum ib_srq_type
	Type uint8
	PDN  uint32
	// CQ of an XRC SRQ
	CQN *uint32
	// QPs using it, as ranges of QP numbers
	QPs []QPRange
	// Process of a user SRQ, or the module of a kernel SRQ
	PID      uint32
	KernName string
	Driver   []Driver
}

// A QPRange is the QP numbers from Min to Max.
type QPRange struct {
	Min uint32
	Max uint32
}

// An SRQReq asks for the SRQs of a device, or for one SRQ by its number:
// SRQs uses DriverDetails, SRQ uses SRQN.
type SRQReq struct {
	DevIndex uint32
	SRQN     uint32
	// Whether to include SRQs only drivers use
	DriverDetails bool
}

func (r SRQReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrResSRQN, r.SRQN)
	if r.DriverDetails {
		ae.Uint8(attrDriverDetails, 1)
	}
	return ae.Encode()
}

// SRQs returns the SRQs of a device, like "rdma resource show srq".
func (c *Conn) SRQs(req SRQReq) ([]SRQ, error) {
	var l srqList
	if err := c.dump(cmdResSRQGet, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: list SRQs of device %d: %w", req.DevIndex, err)
	}

	return l, nil
}

// SRQ returns an SRQ of a device.
func (c *Conn) SRQ(req SRQReq) (SRQ, error) {
	var srq SRQ
	if err := c.get(cmdResSRQGet, req, &srq); err != nil {
		return srq, fmt.Errorf("rdmanl: get SRQ %d of device %d: %w", req.SRQN, req.DevIndex, err)
	}

	return srq, nil
}

// RawSRQs returns the SRQs of a device in the driver's format, for
// drivers that have one, such as mlx5.
func (c *Conn) RawSRQs(req SRQReq) ([][]byte, error) {
	l := rawList{attr: attrResSRQ, entry: attrResSRQEntry}
	if err := c.dump(cmdResSRQGetRaw, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: list raw SRQs of device %d: %w", req.DevIndex, err)
	}

	return l.bytes(), nil
}

// RawSRQ returns an SRQ of a device in the driver's format.
func (c *Conn) RawSRQ(req SRQReq) ([]byte, error) {
	var r raw
	if err := c.get(cmdResSRQGetRaw, req, &r); err != nil {
		return nil, fmt.Errorf("rdmanl: get raw SRQ %d of device %d: %w", req.SRQN, req.DevIndex, err)
	}

	return r, nil
}

// A srqList is the SRQs of a dump.
type srqList []SRQ

func (l *srqList) unmarshal(data []byte) error {
	return unmarshalTable(data, attrResSRQ, attrResSRQEntry, (*[]SRQ)(l))
}

func (srq *SRQ) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrResSRQN:
			srq.ID = ad.Uint32()
		case attrResType:
			srq.Type = ad.Uint8()
		case attrResPDN:
			srq.PDN = ad.Uint32()
		case attrResCQN:
			srq.CQN = new(ad.Uint32())
		case attrResQP:
			if err := unmarshalEntries(ad.Bytes(), attrResQPEntry, &srq.QPs); err != nil {
				return err
			}
		case attrResPID:
			srq.PID = ad.Uint32()
		case attrResKernName:
			srq.KernName = ad.String()
		case attrDriver:
			if err := (*drivers)(&srq.Driver).unmarshal(ad.Bytes()); err != nil {
				return err
			}
		}
	}
	return ad.Err()
}

// parseQPRange parses a QP of an SRQ: one number, or a range.
func (r *QPRange) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrResLQPN:
			r.Min = ad.Uint32()
			r.Max = r.Min
		case attrMinRange:
			r.Min = ad.Uint32()
		case attrMaxRange:
			r.Max = ad.Uint32()
		}
	}
	return ad.Err()
}
