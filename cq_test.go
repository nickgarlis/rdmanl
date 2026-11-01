//go:build linux

package rdmanl

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netlink"
)

func TestCQ(t *testing.T) {
	// A reply for one CQ has its attributes next to the device's.
	c := fakeConn(t, Config{}, func(cmd uint16, _ netlink.HeaderFlags, attrs map[uint16][]byte) [][]byte {
		if cmd != cmdResCQGet || u32(t, attrs, attrResCQN) != 7 {
			t.Fatalf("request %d %v, want CMD_RES_CQ_GET of CQ 7", cmd, attrs)
		}

		return [][]byte{encode(t, func(ae *netlink.AttributeEncoder) {
			ae.Uint32(attrDevIndex, 3)
			ae.String(attrDevName, "rxe0")
			ae.Uint32(attrResCQE, 255)
			ae.Uint64(attrResUseCnt, 2)
			ae.Uint32(attrResCQN, 7)
			ae.Uint32(attrResCtxN, 1)
		})}
	})

	cq, err := c.CQ(CQReq{DevIndex: 3, CQN: 7})
	if err != nil {
		t.Fatalf("CQ: %v", err)
	}

	if diff := cmp.Diff(CQ{ID: 7, CQE: 255, Users: 2, CtxN: new(uint32(1))}, cq); diff != "" {
		t.Errorf("CQ (-want +got):\n%s", diff)
	}
}
