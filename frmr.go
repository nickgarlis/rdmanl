//go:build linux

package rdmanl

import (
	"fmt"

	"github.com/mdlayher/netlink"
)

// An FRMRKey identifies a pool of fast-registration MRs: the kind of MRs
// it holds.
//
// https://github.com/torvalds/linux/blob/v7.3/include/uapi/rdma/rdma_netlink.h#L596
type FRMRKey struct {
	ATS          bool
	AccessFlags  uint32
	VendorKey    uint64
	NumDMABlocks uint64
	// Set for pools the kernel uses internally; it can't be set in
	// SetFRMRPinnedHandles
	KernelVendorKey uint64
}

// An FRMRPool is a pool of fast-registration MRs the kernel keeps for a
// device.
type FRMRPool struct {
	Key FRMRKey
	// Handles in the pool
	QueueHandles uint32
	// The most handles in use at once, and those in use now
	MaxInUse uint64
	InUse    uint64
	// Handles kept from aging out
	PinnedHandles uint32
}

// An FRMRPoolsReq asks for the FRMR pools of a device.
type FRMRPoolsReq struct {
	DevIndex uint32
	// Whether to include the pools of keys only the kernel uses
	DriverDetails bool
}

func (r FRMRPoolsReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	if r.DriverDetails {
		ae.Uint8(attrDriverDetails, 1)
	}
	return ae.Encode()
}

// FRMRPools returns the FRMR pools of a device.
func (c *Conn) FRMRPools(req FRMRPoolsReq) ([]FRMRPool, error) {
	var l frmrPoolList
	if err := c.dump(cmdFRMRPoolsGet, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: list FRMR pools of device %d: %w", req.DevIndex, err)
	}

	return l, nil
}

// A SetFRMRAgingPeriodReq sets how many seconds an unused FRMR handle
// of a device stays in its pool.
type SetFRMRAgingPeriodReq struct {
	DevIndex    uint32
	AgingPeriod uint32
}

func (r SetFRMRAgingPeriodReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrFRMRPoolsAgingPeriod, r.AgingPeriod)
	return ae.Encode()
}

// SetFRMRAgingPeriod sets the aging period of a device's FRMR pools.
func (c *Conn) SetFRMRAgingPeriod(req SetFRMRAgingPeriodReq) error {
	if err := c.do(cmdFRMRPoolsSet, req); err != nil {
		return fmt.Errorf("rdmanl: set FRMR aging period of device %d: %w", req.DevIndex, err)
	}

	return nil
}

// A SetFRMRPinnedHandlesReq keeps PinnedHandles handles of the pool of
// Key from aging out.
type SetFRMRPinnedHandlesReq struct {
	DevIndex      uint32
	Key           FRMRKey
	PinnedHandles uint32
}

func (r SetFRMRPinnedHandlesReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrFRMRPoolPinnedHandles, r.PinnedHandles)
	ae.Nested(attrFRMRPoolKey, r.Key.marshal)
	return ae.Encode()
}

// SetFRMRPinnedHandles pins handles in an FRMR pool.
func (c *Conn) SetFRMRPinnedHandles(req SetFRMRPinnedHandlesReq) error {
	if err := c.do(cmdFRMRPoolsSet, req); err != nil {
		return fmt.Errorf("rdmanl: pin FRMR handles of device %d: %w", req.DevIndex, err)
	}

	return nil
}

// marshal encodes the key, without the kernel vendor key, which the kernel
// refuses from user space.
func (k *FRMRKey) marshal(ae *netlink.AttributeEncoder) error {
	putBool(ae, attrFRMRPoolKeyATS, k.ATS)
	ae.Uint32(attrFRMRPoolKeyAccessFlags, k.AccessFlags)
	ae.Uint64(attrFRMRPoolKeyVendorKey, k.VendorKey)
	ae.Uint64(attrFRMRPoolKeyNumDMABlocks, k.NumDMABlocks)
	return nil
}

// An frmrPoolList is the FRMR pools of a dump.
type frmrPoolList []FRMRPool

func (l *frmrPoolList) unmarshal(data []byte) error {
	return unmarshalTable(data, attrFRMRPools, attrFRMRPoolEntry, (*[]FRMRPool)(l))
}

func (p *FRMRPool) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrFRMRPoolKey:
			if err := p.Key.unmarshal(ad.Bytes()); err != nil {
				return err
			}
		case attrFRMRPoolQueueHandles:
			p.QueueHandles = ad.Uint32()
		case attrFRMRPoolMaxInUse:
			p.MaxInUse = ad.Uint64()
		case attrFRMRPoolInUse:
			p.InUse = ad.Uint64()
		case attrFRMRPoolPinnedHandles:
			p.PinnedHandles = ad.Uint32()
		}
	}
	return ad.Err()
}

func (k *FRMRKey) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrFRMRPoolKeyATS:
			k.ATS = ad.Uint8() != 0
		case attrFRMRPoolKeyAccessFlags:
			k.AccessFlags = ad.Uint32()
		case attrFRMRPoolKeyVendorKey:
			k.VendorKey = ad.Uint64()
		case attrFRMRPoolKeyNumDMABlocks:
			k.NumDMABlocks = ad.Uint64()
		case attrFRMRPoolKeyKernelVendorKey:
			k.KernelVendorKey = ad.Uint64()
		}
	}
	return ad.Err()
}
