//go:build linux

package rdmanl

import (
	"fmt"

	"github.com/mdlayher/netlink"
)

// CounterMode is how QPs are bound to counters.
//
// https://github.com/torvalds/linux/blob/v7.3/include/uapi/rdma/rdma_netlink.h#L624
type CounterMode uint32

const (
	CounterModeNone   CounterMode = 0 // RDMA_COUNTER_MODE_NONE
	CounterModeAuto   CounterMode = 1 // RDMA_COUNTER_MODE_AUTO
	CounterModeManual CounterMode = 2 // RDMA_COUNTER_MODE_MANUAL
)

// CounterMask is the criteria by which CounterModeAuto groups QPs.
//
// https://github.com/torvalds/linux/blob/v7.3/include/uapi/rdma/rdma_netlink.h#L649
type CounterMask uint32

const (
	CounterMaskQPType CounterMask = 1 << 0 // RDMA_COUNTER_MASK_QP_TYPE
	CounterMaskPID    CounterMask = 1 << 1 // RDMA_COUNTER_MASK_PID
)

// A Counter is a hardware counter.
type Counter struct {
	Name  string
	Value uint64
}

// A CounterStatus tells whether a counter of a port is optional, and
// enabled.
type CounterStatus struct {
	Name string
	// Index, for EnableCounters
	Index    uint32
	Optional bool
	Enabled  bool
}

// A CounterConfig is how the QPs of a port are bound to counters.
type CounterConfig struct {
	Mode CounterMode
	// With CounterModeAuto: the criteria, and whether per-opcode counters
	// are on
	Mask       CounterMask
	Opcounters bool
}

// A QPCounter is a counter QPs are bound to.
type QPCounter struct {
	Port uint32
	ID   uint32
	Mode CounterMode
	// The QP type and process it counts, with CounterModeAuto
	QPType   *uint8
	PID      uint32
	KernName string
	// The QPs bound to it
	QPs      []uint32
	Counters []Counter
}

// MRCounters are the counters of an MR, for drivers that count per MR.
type MRCounters struct {
	ID       uint32
	Counters []Counter
}

// A CountersReq asks for the counters of a port.
type CountersReq struct {
	DevIndex  uint32
	PortIndex uint32
}

func (r CountersReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrPortIndex, r.PortIndex)
	return ae.Encode()
}

// Counters returns the counters of a port, like "rdma statistic show".
func (c *Conn) Counters(req CountersReq) ([]Counter, error) {
	var l counterList
	if err := c.get(cmdStatGet, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: get counters of port %d of device %d: %w", req.PortIndex, req.DevIndex, err)
	}

	return l, nil
}

// A CounterStatusReq asks which counters of a port are optional, and
// enabled.
type CounterStatusReq struct {
	DevIndex  uint32
	PortIndex uint32
}

func (r CounterStatusReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrPortIndex, r.PortIndex)
	return ae.Encode()
}

// CounterStatus returns the status of the counters of a port, like
// "rdma statistic mode supported".
func (c *Conn) CounterStatus(req CounterStatusReq) ([]CounterStatus, error) {
	var l counterStatusList
	if err := c.get(cmdStatGetStatus, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: get counter status of port %d of device %d: %w", req.PortIndex, req.DevIndex, err)
	}

	return l, nil
}

// An EnableCountersReq enables the optional counters of a port with the
// indices Indices, and disables the others. The kernel takes a list of
// indices, not of counters; an empty list disables them all.
type EnableCountersReq struct {
	DevIndex  uint32
	PortIndex uint32
	Indices   []uint32
}

func (r EnableCountersReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrPortIndex, r.PortIndex)
	ae.Nested(attrStatHWCounters, func(ae *netlink.AttributeEncoder) error {
		for _, i := range r.Indices {
			ae.Uint32(attrStatHWCounterIndex, i)
		}
		return nil
	})
	return ae.Encode()
}

// EnableCounters enables optional counters, like "rdma statistic set".
func (c *Conn) EnableCounters(req EnableCountersReq) error {
	if err := c.do(cmdStatSet, req); err != nil {
		return fmt.Errorf("rdmanl: enable counters of port %d of device %d: %w", req.PortIndex, req.DevIndex, err)
	}

	return nil
}

// A CounterModeReq asks how the QPs of a port are bound to counters.
type CounterModeReq struct {
	DevIndex  uint32
	PortIndex uint32
}

func (r CounterModeReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrPortIndex, r.PortIndex)
	ae.Uint32(attrStatRes, attrResQP)
	// The kernel wants a mode in the request; it doesn't use its value.
	ae.Uint32(attrStatMode, uint32(CounterModeNone))
	return ae.Encode()
}

// CounterMode returns how the QPs of a port are bound to counters, like
// "rdma statistic qp mode".
func (c *Conn) CounterMode(req CounterModeReq) (CounterConfig, error) {
	var cfg CounterConfig
	if err := c.get(cmdStatGet, req, &cfg); err != nil {
		return cfg, fmt.Errorf("rdmanl: get counter mode of port %d of device %d: %w", req.PortIndex, req.DevIndex, err)
	}

	return cfg, nil
}

// A SetAutoCounterModeReq binds new QPs of a port to counters
// automatically: a counter for each group of QPs with the same criteria of
// Mask. A Mask of 0 stops it. Opcounters turns per-opcode counters on.
type SetAutoCounterModeReq struct {
	DevIndex   uint32
	PortIndex  uint32
	Mask       CounterMask
	Opcounters bool
}

func (r SetAutoCounterModeReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrPortIndex, r.PortIndex)
	ae.Uint32(attrStatRes, attrResQP)
	ae.Uint32(attrStatMode, uint32(CounterModeAuto))
	ae.Uint32(attrStatAutoModeMask, uint32(r.Mask))
	putBool(ae, attrStatOpcounterEnabled, r.Opcounters)
	return ae.Encode()
}

// SetAutoCounterMode sets the automatic counter mode, like
// "rdma statistic qp set link auto".
func (c *Conn) SetAutoCounterMode(req SetAutoCounterModeReq) error {
	if err := c.do(cmdStatSet, req); err != nil {
		return fmt.Errorf("rdmanl: set auto counter mode of port %d of device %d: %w", req.PortIndex, req.DevIndex, err)
	}

	return nil
}

// A BindQPReq binds the QP LQPN of a port to the counter CounterID, or
// to a new counter if CounterID is nil.
type BindQPReq struct {
	DevIndex  uint32
	PortIndex uint32
	LQPN      uint32
	CounterID *uint32
}

func (r BindQPReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrPortIndex, r.PortIndex)
	ae.Uint32(attrStatRes, attrResQP)
	ae.Uint32(attrStatMode, uint32(CounterModeManual))
	ae.Uint32(attrResLQPN, r.LQPN)
	if r.CounterID != nil {
		ae.Uint32(attrStatCounterID, *r.CounterID)
	}
	return ae.Encode()
}

// BindQP binds a QP to a counter and returns the counter's ID, like
// "rdma statistic qp bind".
func (c *Conn) BindQP(req BindQPReq) (uint32, error) {
	var b boundQP
	if err := c.get(cmdStatSet, req, &b); err != nil {
		return 0, fmt.Errorf("rdmanl: bind QP %d to a counter: %w", req.LQPN, err)
	}

	if b.counterID == nil {
		return 0, fmt.Errorf("rdmanl: bind QP %d to a counter: no counter in the reply", req.LQPN)
	}

	return *b.counterID, nil
}

// An UnbindQPReq unbinds the QP LQPN of a port from the counter
// CounterID.
type UnbindQPReq struct {
	DevIndex  uint32
	PortIndex uint32
	CounterID uint32
	LQPN      uint32
}

func (r UnbindQPReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrPortIndex, r.PortIndex)
	ae.Uint32(attrStatRes, attrResQP)
	ae.Uint32(attrStatCounterID, r.CounterID)
	ae.Uint32(attrResLQPN, r.LQPN)
	return ae.Encode()
}

// UnbindQP unbinds a QP from a counter, like "rdma statistic qp unbind".
func (c *Conn) UnbindQP(req UnbindQPReq) error {
	if err := c.do(cmdStatDel, req); err != nil {
		return fmt.Errorf("rdmanl: unbind QP %d from counter %d: %w", req.LQPN, req.CounterID, err)
	}

	return nil
}

// A QPCounterReq asks for the counters QPs of a device are bound to, or
// for one of them: QPCounters uses DriverDetails, QPCounter uses CounterID.
type QPCounterReq struct {
	DevIndex uint32
	// The port: a filter for QPCounters, if not 0; required for QPCounter
	PortIndex uint32
	CounterID uint32
	// Whether to include counters only drivers use
	DriverDetails bool
}

func (r QPCounterReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	if r.PortIndex != 0 {
		ae.Uint32(attrPortIndex, r.PortIndex)
	}
	ae.Uint32(attrStatRes, attrResQP)
	ae.Uint32(attrStatCounterID, r.CounterID)
	if r.DriverDetails {
		ae.Uint8(attrDriverDetails, 1)
	}
	return ae.Encode()
}

// QPCounters returns the counters QPs of a device are bound to, like
// "rdma statistic qp show".
func (c *Conn) QPCounters(req QPCounterReq) ([]QPCounter, error) {
	var l qpCounterList
	if err := c.dump(cmdStatGet, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: list QP counters of device %d: %w", req.DevIndex, err)
	}

	return l, nil
}

// QPCounter returns a counter QPs are bound to.
func (c *Conn) QPCounter(req QPCounterReq) (QPCounter, error) {
	var ctr QPCounter
	if err := c.get(cmdStatGet, req, &ctr); err != nil {
		return ctr, fmt.Errorf("rdmanl: get QP counter %d of device %d: %w", req.CounterID, req.DevIndex, err)
	}

	return ctr, nil
}

// An MRCounterReq asks for the counters of the MRs of a device, or of one
// MR: MRCounters uses DriverDetails, MRCounter uses MRN.
type MRCounterReq struct {
	DevIndex uint32
	MRN      uint32
	// Whether to include MRs only drivers use
	DriverDetails bool
}

func (r MRCounterReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrStatRes, attrResMR)
	ae.Uint32(attrResMRN, r.MRN)
	if r.DriverDetails {
		ae.Uint8(attrDriverDetails, 1)
	}
	return ae.Encode()
}

// MRCounters returns the counters of the MRs of a device, like
// "rdma statistic mr show".
func (c *Conn) MRCounters(req MRCounterReq) ([]MRCounters, error) {
	var l mrCountersList
	if err := c.dump(cmdStatGet, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: list MR counters of device %d: %w", req.DevIndex, err)
	}

	return l, nil
}

// MRCounter returns the counters of an MR.
func (c *Conn) MRCounter(req MRCounterReq) (MRCounters, error) {
	var mr MRCounters
	if err := c.get(cmdStatGet, req, &mr); err != nil {
		return mr, fmt.Errorf("rdmanl: get counters of MR %d of device %d: %w", req.MRN, req.DevIndex, err)
	}

	return mr, nil
}

// A boundQP is the reply to binding a QP to a new counter.
type boundQP struct {
	counterID *uint32
}

func (b *boundQP) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		if ad.Type() == attrStatCounterID {
			b.counterID = new(ad.Uint32())
		}
	}
	return ad.Err()
}

// A counterList is the counters of a port.
type counterList []Counter

func (l *counterList) unmarshal(data []byte) error {
	return unmarshalTable(data, attrStatHWCounters, attrStatHWCounterEntry, (*[]Counter)(l))
}

// A counterStatusList is the status of the counters of a port.
type counterStatusList []CounterStatus

func (l *counterStatusList) unmarshal(data []byte) error {
	return unmarshalTable(data, attrStatHWCounters, attrStatHWCounterEntry, (*[]CounterStatus)(l))
}

// A qpCounterList is the counters of a dump that QPs are bound to.
type qpCounterList []QPCounter

func (l *qpCounterList) unmarshal(data []byte) error {
	return unmarshalTable(data, attrStatCounter, attrStatCounterEntry, (*[]QPCounter)(l))
}

// An mrCountersList is the counters of the MRs of a dump.
type mrCountersList []MRCounters

func (l *mrCountersList) unmarshal(data []byte) error {
	return unmarshalTable(data, attrResMR, attrResMREntry, (*[]MRCounters)(l))
}

func (ctr *Counter) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrStatHWCounterEntryName:
			ctr.Name = ad.String()
		case attrStatHWCounterEntryValue:
			ctr.Value = ad.Uint64()
		}
	}
	return ad.Err()
}

func (s *CounterStatus) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrStatHWCounterEntryName:
			s.Name = ad.String()
		case attrStatHWCounterIndex:
			s.Index = ad.Uint32()
		case attrStatHWCounterDynamic:
			// Only optional counters say whether they are enabled.
			s.Optional = true
			s.Enabled = ad.Uint8() != 0
		}
	}
	return ad.Err()
}

func (cfg *CounterConfig) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrStatMode:
			cfg.Mode = CounterMode(ad.Uint32())
		case attrStatAutoModeMask:
			cfg.Mask = CounterMask(ad.Uint32())
		case attrStatOpcounterEnabled:
			cfg.Opcounters = ad.Uint8() != 0
		}
	}
	return ad.Err()
}

func (ctr *QPCounter) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrPortIndex:
			ctr.Port = ad.Uint32()
		case attrStatCounterID:
			ctr.ID = ad.Uint32()
		case attrStatMode:
			ctr.Mode = CounterMode(ad.Uint32())
		case attrResType:
			ctr.QPType = new(ad.Uint8())
		case attrResPID:
			ctr.PID = ad.Uint32()
		case attrResKernName:
			ctr.KernName = ad.String()
		case attrResQP:
			var qps []QPRange
			if err := unmarshalEntries(ad.Bytes(), attrResQPEntry, &qps); err != nil {
				return err
			}

			for _, r := range qps {
				ctr.QPs = append(ctr.QPs, r.Min)
			}
		case attrStatHWCounters:
			if err := unmarshalEntries(ad.Bytes(), attrStatHWCounterEntry, &ctr.Counters); err != nil {
				return err
			}
		}
	}
	return ad.Err()
}

func (mr *MRCounters) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrResMRN:
			mr.ID = ad.Uint32()
		case attrStatHWCounters:
			if err := unmarshalEntries(ad.Bytes(), attrStatHWCounterEntry, &mr.Counters); err != nil {
				return err
			}
		}
	}
	return ad.Err()
}
