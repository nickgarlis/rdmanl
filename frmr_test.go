//go:build linux

package rdmanl

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netlink"
)

func TestSetFRMRPinnedHandles(t *testing.T) {
	c := fakeConn(t, Config{}, func(cmd uint16, _ netlink.HeaderFlags, attrs map[uint16][]byte) [][]byte {
		if cmd != cmdFRMRPoolsSet || u32(t, attrs, attrFRMRPoolPinnedHandles) != 16 {
			t.Fatalf("request %d %v, want 16 pinned handles", cmd, attrs)
		}

		var key FRMRKey
		if err := key.unmarshal(attrs[attrFRMRPoolKey]); err != nil {
			t.Fatalf("key: %v", err)
		}

		// The kernel refuses a kernel vendor key from user space.
		if diff := cmp.Diff(FRMRKey{ATS: true, AccessFlags: 7, NumDMABlocks: 32}, key); diff != "" {
			t.Errorf("key (-want +got):\n%s", diff)
		}

		return nil
	})

	if err := c.SetFRMRPinnedHandles(SetFRMRPinnedHandlesReq{
		DevIndex:      3,
		Key:           FRMRKey{ATS: true, AccessFlags: 7, NumDMABlocks: 32, KernelVendorKey: 9},
		PinnedHandles: 16,
	}); err != nil {
		t.Fatalf("SetFRMRPinnedHandles: %v", err)
	}
}
