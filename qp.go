//go:build linux

package rdmanl

import "fmt"

// A QP is a queue pair.
type QP struct {
	// Port, or 0 until the QP is bound to one
	Port uint32
	QPN  uint32
	// Type and state, enum ib_qp_type and ib_qp_state
	Type  uint8
	State uint8
	// RC and UC QPs only: the remote QP number and the receive queue PSN
	RemoteQPN *uint32
	RQPSN     *uint32
	SQPSN     uint32
	// RC, UC and XRC QPs only: path migration state, enum ib_mig_state
	PathMigState *uint8
	// Subtype of a driver QP
	Subtype string
	// PD of a user QP
	PDN *uint32
	// Process of a user QP, or the module of a kernel QP
	PID      uint32
	KernName string
	Driver   []Driver
}

// A QPReq asks for the QPs of a device, or for one QP by its number:
// QPs uses DriverDetails, QP uses LQPN.
type QPReq struct {
	DevIndex uint32
	// The port: a filter for QPs, if not 0; required for QP
	PortIndex uint32
	LQPN      uint32
	// Whether to include QPs only drivers use
	DriverDetails bool
}

func (r QPReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	if r.PortIndex != 0 {
		ae.Uint32(attrPortIndex, r.PortIndex)
	}
	ae.Uint32(attrResLQPN, r.LQPN)
	if r.DriverDetails {
		ae.Uint8(attrDriverDetails, 1)
	}
	return ae.Encode()
}

// QPs returns the QPs of a device, like "rdma resource show qp".
func (c *Conn) QPs(req QPReq) ([]QP, error) {
	var l qpList
	if err := c.dump(cmdResQPGet, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: list QPs of device %d: %w", req.DevIndex, err)
	}

	return l, nil
}

// QP returns a QP of a device.
func (c *Conn) QP(req QPReq) (QP, error) {
	var q QP
	if err := c.get(cmdResQPGet, req, &q); err != nil {
		return q, fmt.Errorf("rdmanl: get QP %d of device %d: %w", req.LQPN, req.DevIndex, err)
	}

	return q, nil
}

// RawQPs returns the QPs of a device in the driver's format, for
// drivers that have one, such as mlx5.
func (c *Conn) RawQPs(req QPReq) ([][]byte, error) {
	l := rawList{attr: attrResQP, entry: attrResQPEntry}
	if err := c.dump(cmdResQPGetRaw, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: list raw QPs of device %d: %w", req.DevIndex, err)
	}

	return l.bytes(), nil
}

// RawQP returns a QP of a device in the driver's format.
func (c *Conn) RawQP(req QPReq) ([]byte, error) {
	var r raw
	if err := c.get(cmdResQPGetRaw, req, &r); err != nil {
		return nil, fmt.Errorf("rdmanl: get raw QP %d of device %d: %w", req.LQPN, req.DevIndex, err)
	}

	return r, nil
}

// A qpList is the QPs of a dump.
type qpList []QP

func (l *qpList) unmarshal(data []byte) error {
	return unmarshalTable(data, attrResQP, attrResQPEntry, (*[]QP)(l))
}

func (q *QP) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrPortIndex:
			q.Port = ad.Uint32()
		case attrResLQPN:
			q.QPN = ad.Uint32()
		case attrResType:
			q.Type = ad.Uint8()
		case attrResState:
			q.State = ad.Uint8()
		case attrResRQPN:
			q.RemoteQPN = new(ad.Uint32())
		case attrResRQPSN:
			q.RQPSN = new(ad.Uint32())
		case attrResSQPSN:
			q.SQPSN = ad.Uint32()
		case attrResPathMigState:
			q.PathMigState = new(ad.Uint8())
		case attrResSubtype:
			q.Subtype = ad.String()
		case attrResPDN:
			q.PDN = new(ad.Uint32())
		case attrResPID:
			q.PID = ad.Uint32()
		case attrResKernName:
			q.KernName = ad.String()
		case attrDriver:
			if err := (*drivers)(&q.Driver).unmarshal(ad.Bytes()); err != nil {
				return err
			}
		}
	}
	return ad.Err()
}
