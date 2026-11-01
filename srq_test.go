//go:build linux

package rdmanl

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netlink"
)

func TestSRQRanges(t *testing.T) {
	data := encode(t, func(ae *netlink.AttributeEncoder) {
		ae.Uint32(attrResSRQN, 1)
		ae.Nested(attrResQP, func(ae *netlink.AttributeEncoder) error {
			ae.Nested(attrResQPEntry, func(ae *netlink.AttributeEncoder) error {
				ae.Uint32(attrResLQPN, 5)
				return nil
			})
			ae.Nested(attrResQPEntry, func(ae *netlink.AttributeEncoder) error {
				ae.Uint32(attrMinRange, 8)
				ae.Uint32(attrMaxRange, 11)
				return nil
			})
			return nil
		})
	})

	var srq SRQ
	if err := srq.unmarshal(data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if diff := cmp.Diff([]QPRange{{5, 5}, {8, 11}}, srq.QPs); diff != "" {
		t.Errorf("QPs of the SRQ (-want +got):\n%s", diff)
	}
}
