//go:build linux

package rdmanl

import "fmt"

// Sys is the settings of the RDMA subsystem.
type Sys struct {
	// Whether devices are shared by all network namespaces, rather than
	// each belonging to one
	NetNSShared bool
	// Whether only privileged processes may use privileged Q_Keys
	PrivilegedQKey bool
	// Whether the kernel sends monitoring events (see Monitor)
	Monitor bool
	// Whether pinned memory is copied on fork, rather than shared
	CopyOnFork bool
}

// Sys returns the settings of the RDMA subsystem, like "rdma system show".
func (c *Conn) Sys() (Sys, error) {
	var s Sys
	if err := c.get(cmdSysGet, nil, &s); err != nil {
		return s, fmt.Errorf("rdmanl: get system settings: %w", err)
	}

	return s, nil
}

// A SetNetNSModeReq sets whether devices are shared by all network
// namespaces, rather than each belonging to one.
type SetNetNSModeReq struct {
	NetNSShared bool
}

func (r SetNetNSModeReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	putBool(ae, attrSysNetNSMode, r.NetNSShared)
	return ae.Encode()
}

// SetNetNSMode sets the network namespace mode, like
// "rdma system set netns".
func (c *Conn) SetNetNSMode(req SetNetNSModeReq) error {
	if err := c.do(cmdSysSet, req); err != nil {
		return fmt.Errorf("rdmanl: set network namespace mode: %w", err)
	}

	return nil
}

// A SetPrivilegedQKeyModeReq sets whether only privileged processes may
// use privileged Q_Keys.
type SetPrivilegedQKeyModeReq struct {
	PrivilegedQKey bool
}

func (r SetPrivilegedQKeyModeReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	putBool(ae, attrSysPrivilegedQKeyMode, r.PrivilegedQKey)
	return ae.Encode()
}

// SetPrivilegedQKeyMode sets the privileged Q_Key mode, like
// "rdma system set privileged-qkey".
func (c *Conn) SetPrivilegedQKeyMode(req SetPrivilegedQKeyModeReq) error {
	if err := c.do(cmdSysSet, req); err != nil {
		return fmt.Errorf("rdmanl: set privileged Q_Key mode: %w", err)
	}

	return nil
}

func (s *Sys) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrSysNetNSMode:
			s.NetNSShared = ad.Uint8() != 0
		case attrSysPrivilegedQKeyMode:
			s.PrivilegedQKey = ad.Uint8() != 0
		case attrSysMonitorMode:
			s.Monitor = ad.Uint8() != 0
		case attrSysCopyOnFork:
			s.CopyOnFork = ad.Uint8() != 0
		}
	}
	return ad.Err()
}
