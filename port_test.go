//go:build linux

package rdmanl

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netlink"
)

func TestPort(t *testing.T) {
	c := fakeConn(t, Config{}, func(cmd uint16, _ netlink.HeaderFlags, attrs map[uint16][]byte) [][]byte {
		if cmd != cmdPortGet || u32(t, attrs, attrDevIndex) != 3 || u32(t, attrs, attrPortIndex) != 1 {
			t.Fatalf("request %d %v, want CMD_PORT_GET of port 1 of device 3", cmd, attrs)
		}

		return [][]byte{encode(t, func(ae *netlink.AttributeEncoder) {
			ae.Uint32(attrDevIndex, 3)
			ae.String(attrDevName, "rxe0")
			ae.Uint32(attrPortIndex, 1)
			ae.Uint8(attrPortState, 4)
			ae.Uint8(attrPortPhysState, 5)
			ae.Uint32(attrNdevIndex, 2)
			ae.String(attrNdevName, "eth0")
		})}
	})

	p, err := c.Port(PortReq{DevIndex: 3, PortIndex: 1})
	if err != nil {
		t.Fatalf("Port: %v", err)
	}

	want := Port{DevIndex: 3, DevName: "rxe0", Index: 1, State: 4, PhysState: 5, NdevIndex: 2, NdevName: "eth0"}
	if diff := cmp.Diff(want, p); diff != "" {
		t.Errorf("port (-want +got):\n%s", diff)
	}
}
