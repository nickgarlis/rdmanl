//go:build linux

package rdmanl

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netlink"
)

func TestQPs(t *testing.T) {
	c := fakeConn(t, Config{}, func(cmd uint16, flags netlink.HeaderFlags, attrs map[uint16][]byte) [][]byte {
		if cmd != cmdResQPGet || flags&netlink.Dump == 0 || u32(t, attrs, attrDevIndex) != 3 {
			t.Fatalf("request %d with flags %v, want a dump of CMD_RES_QP_GET of device 3", cmd, flags)
		}

		if _, ok := attrs[attrDriverDetails]; !ok {
			t.Error("DriverDetails did not ask for driver details")
		}

		// The kernel splits a long dump over messages.
		qp := func(ae *netlink.AttributeEncoder, fill func(ae *netlink.AttributeEncoder)) {
			ae.Uint32(attrDevIndex, 3)
			ae.String(attrDevName, "mlx5_0")
			ae.Nested(attrResQP, func(ae *netlink.AttributeEncoder) error {
				ae.Nested(attrResQPEntry, func(ae *netlink.AttributeEncoder) error {
					fill(ae)
					return nil
				})
				return nil
			})
		}

		return [][]byte{
			encode(t, func(ae *netlink.AttributeEncoder) {
				qp(ae, func(ae *netlink.AttributeEncoder) {
					ae.Uint32(attrPortIndex, 1)
					ae.Uint32(attrResLQPN, 0x42)
					ae.Uint32(attrResPDN, 0) // 0 is a valid PD number
					ae.Uint32(attrResRQPN, 0x43)
					ae.Uint8(attrResType, 2)
					ae.Uint8(attrResState, 3)
					ae.Uint32(attrResPID, 1234)
					ae.Nested(attrDriver, func(ae *netlink.AttributeEncoder) error {
						ae.String(attrDriverString, "sq_wqe_cnt")
						ae.Uint32(attrDriverU32, 64)
						return nil
					})
				})
			}),
			encode(t, func(ae *netlink.AttributeEncoder) {
				qp(ae, func(ae *netlink.AttributeEncoder) {
					ae.Uint32(attrResLQPN, 1)
					ae.String(attrResKernName, "ib_core")
				})
			}),
		}
	})

	qps, err := c.QPs(QPReq{DevIndex: 3, DriverDetails: true})
	if err != nil {
		t.Fatalf("QPs: %v", err)
	}

	want := []QP{
		{
			Port: 1, QPN: 0x42, PDN: new(uint32(0)), RemoteQPN: new(uint32(0x43)), Type: 2, State: 3, PID: 1234,
			Driver: []Driver{{Name: "sq_wqe_cnt", U32: new(uint32(64))}},
		},
		{QPN: 1, KernName: "ib_core"},
	}
	if diff := cmp.Diff(want, qps); diff != "" {
		t.Errorf("QPs (-want +got):\n%s", diff)
	}
}
