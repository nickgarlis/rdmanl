//go:build linux

package rdmanl

import (
	"testing"
)

func TestSockaddr(t *testing.T) {
	v4 := make([]byte, 128) // struct sockaddr_storage
	v4[0], v4[1] = 2, 0     // AF_INET, host byte order (little-endian here)
	v4[2], v4[3] = 0x12, 0xb7
	copy(v4[4:], []byte{10, 0, 0, 1})

	v6 := make([]byte, 128)
	v6[0] = 10 // AF_INET6
	v6[2], v6[3] = 0x12, 0xb7
	v6[8], v6[23] = 0xfe, 1

	for _, c := range []struct {
		b    []byte
		want string
	}{{v4, "10.0.0.1:4791"}, {v6, "[fe00::1]:4791"}, {make([]byte, 128), "invalid AddrPort"}} {
		if got := parseSockaddr(c.b); got.String() != c.want {
			t.Errorf("parseSockaddr = %s, want %s", got, c.want)
		}
	}

}
