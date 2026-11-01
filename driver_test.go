//go:build linux

package rdmanl

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netlink"
)

func TestDriverAttrs(t *testing.T) {
	data := encode(t, func(ae *netlink.AttributeEncoder) {
		ae.String(attrDriverString, "qpn")
		ae.Uint8(attrDriverPrintType, uint8(PrintTypeHex))
		ae.Uint32(attrDriverU32, 0x1a)
		ae.String(attrDriverString, "state")
		ae.String(attrDriverString, "ready")
		ae.String(attrDriverString, "delta")
		ae.Uint64(attrDriverS64, uint64(1<<64-5)) // -5
	})

	var got drivers
	if err := got.unmarshal(data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	want := []Driver{
		{Name: "qpn", PrintType: new(PrintTypeHex), U32: new(uint32(0x1a))},
		{Name: "state", String: new("ready")},
		{Name: "delta", S64: new(int64(-5))},
	}
	if diff := cmp.Diff(want, []Driver(got)); diff != "" {
		t.Errorf("driver attributes (-want +got):\n%s", diff)
	}
}
