//go:build linux

package rdmanl

import (
	"testing"

	"github.com/mdlayher/netlink"
	"github.com/mdlayher/netlink/nltest"
)

// fakeConn returns a Conn whose kernel is fn: it gets the request's
// command, flags and attributes, and returns the attributes of the replies.
func fakeConn(t *testing.T, config Config, fn func(cmd uint16, flags netlink.HeaderFlags, attrs map[uint16][]byte) [][]byte) *Conn {
	t.Helper()

	nlconn := nltest.Dial(func(reqs []netlink.Message) ([]netlink.Message, error) {
		req := reqs[0]
		if client := uint16(req.Header.Type) >> clientShift; client != clientNLDev {
			t.Fatalf("request of RDMA netlink client %d", client)
		}

		attrs := map[uint16][]byte{}
		ad, err := newAttributeDecoder(req.Data)
		if err != nil {
			t.Fatalf("request attributes: %v", err)
		}
		for ad.Next() {
			attrs[ad.Type()] = ad.Bytes()
		}

		var replies []netlink.Message
		for _, data := range fn(uint16(req.Header.Type)&(1<<clientShift-1), req.Header.Flags, attrs) {
			replies = append(replies, netlink.Message{Header: req.Header, Data: data})
		}

		// Like the kernel: dumps end with NLMSG_DONE, and requests that
		// ask for one end with an acknowledgement.
		switch {
		case req.Header.Flags&netlink.Dump != 0:
			for i := range replies {
				replies[i].Header.Flags |= netlink.Multi
			}

			done := netlink.Message{Header: netlink.Header{Type: netlink.Done, Flags: netlink.Multi, Sequence: req.Header.Sequence}}
			return append(replies, done), nil
		case req.Header.Flags&netlink.Acknowledge != 0:
			ack, _ := nltest.Error(0, reqs)
			return append(replies, ack...), nil
		}

		return replies, nil
	})

	c := &Conn{nlconn: nlconn, config: config}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// encode builds attributes the way the kernel does.
func encode(t *testing.T, fn func(ae *netlink.AttributeEncoder)) []byte {
	t.Helper()

	ae := newAttributeEncoder()
	fn(ae)
	b, err := ae.Encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	return b
}

func u32(t *testing.T, attrs map[uint16][]byte, typ uint16) uint32 {
	t.Helper()

	b, ok := attrs[typ]
	if !ok || len(b) != 4 {
		t.Fatalf("request has no u32 attribute %#x: %v", typ, attrs)
	}

	ad, _ := newAttributeDecoder(encode(t, func(ae *netlink.AttributeEncoder) { ae.Bytes(typ, b) }))
	ad.Next()
	return ad.Uint32()
}
