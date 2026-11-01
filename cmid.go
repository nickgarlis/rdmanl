//go:build linux

package rdmanl

import (
	"encoding/binary"
	"fmt"
	"net/netip"

	"golang.org/x/sys/unix"
)

// A CMID is an RDMA CM ID.
type CMID struct {
	// Port, or 0 until the ID is bound to one
	Port uint32
	ID   uint32
	// Port space, enum rdma_ucm_port_space, and state, enum rdma_cm_state
	PortSpace uint32
	State     uint8
	// Its QP, if it has one: number and type, enum ib_qp_type
	QPN    *uint32
	QPType *uint8
	// Source and destination addresses, for IPv4 and IPv6
	Src netip.AddrPort
	Dst netip.AddrPort
	// Process of a user ID, or the module of a kernel ID
	PID      uint32
	KernName string
	Driver   []Driver
}

// A CMIDReq asks for the RDMA CM IDs of a device, or for one RDMA CM ID by its number:
// CMIDs uses DriverDetails, CMID uses CMIDN.
type CMIDReq struct {
	DevIndex uint32
	// The port: a filter for CMIDs, if not 0; required for CMID
	PortIndex uint32
	CMIDN     uint32
	// Whether to include RDMA CM IDs only drivers use
	DriverDetails bool
}

func (r CMIDReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	if r.PortIndex != 0 {
		ae.Uint32(attrPortIndex, r.PortIndex)
	}
	ae.Uint32(attrResCMIDN, r.CMIDN)
	if r.DriverDetails {
		ae.Uint8(attrDriverDetails, 1)
	}
	return ae.Encode()
}

// CMIDs returns the RDMA CM IDs of a device, like "rdma resource show cm_id".
func (c *Conn) CMIDs(req CMIDReq) ([]CMID, error) {
	var l cmidList
	if err := c.dump(cmdResCMIDGet, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: list RDMA CM IDs of device %d: %w", req.DevIndex, err)
	}

	return l, nil
}

// CMID returns a RDMA CM ID of a device.
func (c *Conn) CMID(req CMIDReq) (CMID, error) {
	var cm CMID
	if err := c.get(cmdResCMIDGet, req, &cm); err != nil {
		return cm, fmt.Errorf("rdmanl: get RDMA CM ID %d of device %d: %w", req.CMIDN, req.DevIndex, err)
	}

	return cm, nil
}

// A cmidList is the CMIDs of a dump.
type cmidList []CMID

func (l *cmidList) unmarshal(data []byte) error {
	return unmarshalTable(data, attrResCMID, attrResCMIDEntry, (*[]CMID)(l))
}

func (id *CMID) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrPortIndex:
			id.Port = ad.Uint32()
		case attrResCMIDN:
			id.ID = ad.Uint32()
		case attrResPS:
			id.PortSpace = ad.Uint32()
		case attrResState:
			id.State = ad.Uint8()
		case attrResLQPN:
			id.QPN = new(ad.Uint32())
		case attrResType:
			id.QPType = new(ad.Uint8())
		case attrResSrcAddr:
			id.Src = parseSockaddr(ad.Bytes())
		case attrResDstAddr:
			id.Dst = parseSockaddr(ad.Bytes())
		case attrResPID:
			id.PID = ad.Uint32()
		case attrResKernName:
			id.KernName = ad.String()
		case attrDriver:
			if err := (*drivers)(&id.Driver).unmarshal(ad.Bytes()); err != nil {
				return err
			}
		}
	}
	return ad.Err()
}

// parseSockaddr parses an IPv4 or IPv6 struct sockaddr_storage. Other
// families, such as AF_IB, give the zero AddrPort.
func parseSockaddr(b []byte) netip.AddrPort {
	if len(b) < 2 {
		return netip.AddrPort{}
	}

	switch binary.NativeEndian.Uint16(b) {
	case unix.AF_INET:
		if len(b) >= unix.SizeofSockaddrInet4 {
			return netip.AddrPortFrom(netip.AddrFrom4([4]byte(b[4:8])), binary.BigEndian.Uint16(b[2:]))
		}
	case unix.AF_INET6:
		if len(b) >= unix.SizeofSockaddrInet6 {
			return netip.AddrPortFrom(netip.AddrFrom16([16]byte(b[8:24])), binary.BigEndian.Uint16(b[2:]))
		}
	}

	return netip.AddrPort{}
}
