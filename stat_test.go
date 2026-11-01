//go:build linux

package rdmanl

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netlink"
)

func TestEnableCounters(t *testing.T) {
	c := fakeConn(t, Config{}, func(cmd uint16, flags netlink.HeaderFlags, attrs map[uint16][]byte) [][]byte {
		if cmd != cmdStatSet || flags&netlink.Acknowledge == 0 {
			t.Fatalf("request %d with flags %v, want CMD_STAT_SET with an ACK", cmd, flags)
		}

		// The counters to enable are a list of indices, not entries.
		want := encode(t, func(ae *netlink.AttributeEncoder) {
			ae.Uint32(attrStatHWCounterIndex, 4)
			ae.Uint32(attrStatHWCounterIndex, 9)
		})
		if diff := cmp.Diff(want, attrs[attrStatHWCounters]); diff != "" {
			t.Errorf("counter list (-want +got):\n%s", diff)
		}

		return nil
	})

	if err := c.EnableCounters(EnableCountersReq{DevIndex: 3, PortIndex: 1, Indices: []uint32{4, 9}}); err != nil {
		t.Fatalf("EnableCounters: %v", err)
	}
}

func TestBindQP(t *testing.T) {
	c := fakeConn(t, Config{}, func(cmd uint16, _ netlink.HeaderFlags, attrs map[uint16][]byte) [][]byte {
		if cmd != cmdStatSet || u32(t, attrs, attrStatMode) != uint32(CounterModeManual) || u32(t, attrs, attrResLQPN) != 0x42 {
			t.Fatalf("request %d %v, want a manual bind of QP 0x42", cmd, attrs)
		}

		if _, ok := attrs[attrStatCounterID]; ok {
			t.Error("a bind to a new counter names a counter")
		}

		return [][]byte{encode(t, func(ae *netlink.AttributeEncoder) {
			ae.Uint32(attrDevIndex, 3)
			ae.Uint32(attrPortIndex, 1)
			ae.Uint32(attrStatCounterID, 5)
			ae.Uint32(attrResLQPN, 0x42)
		})}
	})

	id, err := c.BindQP(BindQPReq{DevIndex: 3, PortIndex: 1, LQPN: 0x42})
	if err != nil || id != 5 {
		t.Fatalf("BindQP = %d, %v; want 5", id, err)
	}
}

func TestQPCounters(t *testing.T) {
	data := encode(t, func(ae *netlink.AttributeEncoder) {
		ae.Uint32(attrPortIndex, 1)
		ae.Uint32(attrStatCounterID, 2)
		ae.Uint32(attrStatMode, uint32(CounterModeManual))
		ae.Nested(attrResQP, func(ae *netlink.AttributeEncoder) error {
			for _, qpn := range []uint32{0x42, 0x43} {
				ae.Nested(attrResQPEntry, func(ae *netlink.AttributeEncoder) error {
					ae.Uint32(attrResLQPN, qpn)
					return nil
				})
			}
			return nil
		})
		ae.Nested(attrStatHWCounters, func(ae *netlink.AttributeEncoder) error {
			ae.Nested(attrStatHWCounterEntry, func(ae *netlink.AttributeEncoder) error {
				ae.String(attrStatHWCounterEntryName, "rx_write_requests")
				ae.Uint64(attrStatHWCounterEntryValue, 10)
				return nil
			})
			return nil
		})
	})

	var got QPCounter
	if err := got.unmarshal(data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	want := QPCounter{Port: 1, ID: 2, Mode: CounterModeManual, QPs: []uint32{0x42, 0x43}, Counters: []Counter{{"rx_write_requests", 10}}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("counter (-want +got):\n%s", diff)
	}
}
