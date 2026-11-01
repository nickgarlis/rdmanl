//go:build linux

package rdmanl

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netlink"
)

func TestDevices(t *testing.T) {
	c := fakeConn(t, Config{}, func(cmd uint16, flags netlink.HeaderFlags, attrs map[uint16][]byte) [][]byte {
		if cmd != cmdGet || flags&netlink.Dump == 0 {
			t.Fatalf("request %d with flags %v, want a dump of CMD_GET", cmd, flags)
		}

		return [][]byte{
			encode(t, func(ae *netlink.AttributeEncoder) {
				ae.Uint32(attrDevIndex, 3)
				ae.String(attrDevName, "rxe0")
				ae.Uint32(attrPortIndex, 1)
				ae.Bytes(0, make([]byte, 4)) // RDMA_NLDEV_ATTR_PAD, ahead of a u64
				ae.Uint64(attrNodeGUID, 0x1122334455667788)
				ae.Uint8(attrDevDim, 1)
				ae.String(attrDevProtocol, "roce")
			}),
			encode(t, func(ae *netlink.AttributeEncoder) {
				ae.Uint32(attrDevIndex, 4)
				ae.String(attrDevName, "mlx5_0")
			}),
		}
	})

	devs, err := c.Devices()
	if err != nil {
		t.Fatalf("Devices: %v", err)
	}

	want := []Device{
		{Index: 3, Name: "rxe0", Ports: 1, NodeGUID: 0x1122334455667788, Dim: true, Protocol: "roce"},
		{Index: 4, Name: "mlx5_0"},
	}
	if diff := cmp.Diff(want, devs); diff != "" {
		t.Errorf("devices (-want +got):\n%s", diff)
	}
}
