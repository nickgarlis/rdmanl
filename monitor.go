//go:build linux

package rdmanl

import (
	"fmt"
	"time"

	"github.com/mdlayher/netlink"
)

// EventType is the kind of a monitoring event.
//
// https://github.com/torvalds/linux/blob/v7.3/include/uapi/rdma/rdma_netlink.h#L668
type EventType uint8

const (
	EventRegister     EventType = 0 // RDMA_REGISTER_EVENT: a device was added
	EventUnregister   EventType = 1 // RDMA_UNREGISTER_EVENT: a device was removed
	EventNetdevAttach EventType = 2 // RDMA_NETDEV_ATTACH_EVENT: a netdev was attached to a port
	EventNetdevDetach EventType = 3 // RDMA_NETDEV_DETACH_EVENT: a netdev was detached from a port
	EventRename       EventType = 4 // RDMA_RENAME_EVENT: a device was renamed
	EventNetdevRename EventType = 5 // RDMA_NETDEV_RENAME_EVENT: the netdev of a port was renamed
)

// An Event is a change to the devices of a network namespace.
type Event struct {
	Type EventType
	// The device: for every event but EventNetdevRename
	DevIndex uint32
	DevName  string
	// The port and its netdev: for netdev events
	Port      uint32
	NdevIndex uint32
	NdevName  string
}

// A Monitor receives the events of the devices of a network namespace,
// like "rdma monitor".
type Monitor struct {
	nlconn  *netlink.Conn
	pending []Event
}

// Monitor starts receiving events, on a socket of its own. No event after
// Monitor returns is missed.
func (c *Conn) Monitor() (*Monitor, error) {
	nlconn, err := dial(&c.config)
	if err != nil {
		return nil, fmt.Errorf("rdmanl: monitor: %w", err)
	}

	if err := nlconn.JoinGroup(groupNotify); err != nil {
		_ = nlconn.Close()
		return nil, fmt.Errorf("rdmanl: monitor: %w", err)
	}

	return &Monitor{nlconn: nlconn}, nil
}

// Receive waits for the next event, until the deadline, if one is set, or
// until Close. It returns one event per call.
func (m *Monitor) Receive() (Event, error) {
	for len(m.pending) == 0 {
		msgs, err := m.nlconn.Receive()
		if err != nil {
			return Event{}, fmt.Errorf("rdmanl: monitor: %w", err)
		}

		for _, msg := range msgs {
			if uint16(msg.Header.Type) != clientNLDev<<clientShift|cmdMonitor {
				continue
			}

			var e Event
			if err := e.unmarshal(msg.Data); err != nil {
				return Event{}, fmt.Errorf("rdmanl: monitor: %w", err)
			}

			m.pending = append(m.pending, e)
		}
	}

	e := m.pending[0]
	m.pending = m.pending[1:]
	return e, nil
}

// SetDeadline sets when a waiting Receive gives up. The zero time means
// never.
func (m *Monitor) SetDeadline(t time.Time) error {
	return m.nlconn.SetReadDeadline(t)
}

// Close stops receiving events. A waiting Receive returns an error.
func (m *Monitor) Close() error {
	return m.nlconn.Close()
}

func (e *Event) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrEventType:
			e.Type = EventType(ad.Uint8())
		case attrDevIndex:
			e.DevIndex = ad.Uint32()
		case attrDevName:
			e.DevName = ad.String()
		case attrPortIndex:
			e.Port = ad.Uint32()
		case attrNdevIndex:
			e.NdevIndex = ad.Uint32()
		case attrNdevName:
			e.NdevName = ad.String()
		}
	}
	return ad.Err()
}
