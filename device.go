//go:build linux

package rdmanl

import "fmt"

// DevType is the type of a sub-device.
//
// https://github.com/torvalds/linux/blob/v7.3/include/uapi/rdma/rdma_netlink.h#L655
type DevType uint8

const (
	// A sub-device for subnet management (QP0) of its parent.
	DevTypeSMI DevType = 1 // RDMA_DEVICE_TYPE_SMI
)

// NameAssignType tells how a device got its name.
//
// https://github.com/torvalds/linux/blob/v7.3/include/uapi/rdma/rdma_netlink.h#L660
type NameAssignType uint8

const (
	NameAssignTypeUnknown NameAssignType = 0 // RDMA_NAME_ASSIGN_TYPE_UNKNOWN
	NameAssignTypeUser    NameAssignType = 1 // RDMA_NAME_ASSIGN_TYPE_USER
)

// A Device is an RDMA device.
type Device struct {
	Index uint32
	// Name, such as "mlx5_0" or "rxe0"
	Name string
	// Number of ports; they are numbered from 1
	Ports uint32
	// Capabilities, enum ib_device_cap_flags
	CapFlags uint64
	// Firmware version, if the device has firmware
	FWVersion string
	// Node and system image GUIDs, in host byte order
	NodeGUID     uint64
	SysImageGUID uint64
	// Node type, enum rdma_node_type
	NodeType uint8
	// Whether CQs use adaptive moderation (DIM)
	Dim bool
	// Protocol, such as "ib", "roce", "iw" or "opa"
	Protocol string
	// Type and parent of a sub-device; zero for other devices
	Type       DevType
	ParentName string
	// How the device got its name
	NameAssignType NameAssignType
}

// Devices returns the RDMA devices of the connection's network namespace.
func (c *Conn) Devices() ([]Device, error) {
	var devs deviceList
	if err := c.dump(cmdGet, nil, &devs); err != nil {
		return nil, fmt.Errorf("rdmanl: list devices: %w", err)
	}

	return devs, nil
}

// A DeviceReq asks for a device.
type DeviceReq struct {
	DevIndex uint32
}

func (r DeviceReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	return ae.Encode()
}

// Device returns a device.
func (c *Conn) Device(req DeviceReq) (Device, error) {
	var d Device
	if err := c.get(cmdGet, req, &d); err != nil {
		return d, fmt.Errorf("rdmanl: get device %d: %w", req.DevIndex, err)
	}

	return d, nil
}

// A NewLinkReq adds a software device named DevName of the link type
// LinkType, such as "rxe" or "siw", on the netdev NdevName.
type NewLinkReq struct {
	DevName  string
	LinkType string
	NdevName string
}

func (r NewLinkReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.String(attrDevName, r.DevName)
	ae.String(attrLinkType, r.LinkType)
	ae.String(attrNdevName, r.NdevName)
	return ae.Encode()
}

// NewLink adds a software device, like "rdma link add".
func (c *Conn) NewLink(req NewLinkReq) error {
	if err := c.do(cmdNewLink, req); err != nil {
		return fmt.Errorf("rdmanl: add %s link %q: %w", req.LinkType, req.DevName, err)
	}

	return nil
}

// A DelLinkReq deletes a software device.
type DelLinkReq struct {
	DevIndex uint32
}

func (r DelLinkReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	return ae.Encode()
}

// DelLink deletes a software device, like "rdma link delete".
func (c *Conn) DelLink(req DelLinkReq) error {
	if err := c.do(cmdDelLink, req); err != nil {
		return fmt.Errorf("rdmanl: delete link of device %d: %w", req.DevIndex, err)
	}

	return nil
}

// A NewDevReq adds a sub-device named DevName of the type DevType to
// the device DevIndex.
type NewDevReq struct {
	DevIndex uint32
	DevName  string
	DevType  DevType
}

func (r NewDevReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.String(attrDevName, r.DevName)
	ae.Uint8(attrDevType, uint8(r.DevType))
	return ae.Encode()
}

// NewDev adds a sub-device, like "rdma dev add".
func (c *Conn) NewDev(req NewDevReq) error {
	if err := c.do(cmdNewDev, req); err != nil {
		return fmt.Errorf("rdmanl: add sub-device %q of device %d: %w", req.DevName, req.DevIndex, err)
	}

	return nil
}

// A DelDevReq deletes a sub-device.
type DelDevReq struct {
	DevIndex uint32
}

func (r DelDevReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	return ae.Encode()
}

// DelDev deletes a sub-device, like "rdma dev delete".
func (c *Conn) DelDev(req DelDevReq) error {
	if err := c.do(cmdDelDev, req); err != nil {
		return fmt.Errorf("rdmanl: delete sub-device %d: %w", req.DevIndex, err)
	}

	return nil
}

// A RenameReq renames a device to DevName.
type RenameReq struct {
	DevIndex uint32
	DevName  string
}

func (r RenameReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.String(attrDevName, r.DevName)
	return ae.Encode()
}

// Rename renames a device, like "rdma dev set name".
func (c *Conn) Rename(req RenameReq) error {
	if err := c.do(cmdSet, req); err != nil {
		return fmt.Errorf("rdmanl: rename device %d to %q: %w", req.DevIndex, req.DevName, err)
	}

	return nil
}

// A SetNetNSReq moves a device to the network namespace of the file
// descriptor NetNSFD, and names it DevName there, unless DevName is empty.
type SetNetNSReq struct {
	DevIndex uint32
	NetNSFD  int
	DevName  string
}

func (r SetNetNSReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrNetNSFD, uint32(r.NetNSFD))
	if r.DevName != "" {
		ae.String(attrDevName, r.DevName)
	}
	return ae.Encode()
}

// SetNetNS moves a device to another network namespace, like
// "rdma dev set netns". It needs the exclusive network namespace mode.
func (c *Conn) SetNetNS(req SetNetNSReq) error {
	if err := c.do(cmdSet, req); err != nil {
		return fmt.Errorf("rdmanl: move device %d to another network namespace: %w", req.DevIndex, err)
	}

	return nil
}

// A SetDimReq turns adaptive moderation (DIM) of a device's CQs on or
// off.
type SetDimReq struct {
	DevIndex uint32
	DevDim   bool
}

func (r SetDimReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	putBool(ae, attrDevDim, r.DevDim)
	return ae.Encode()
}

// SetDim turns adaptive moderation on or off, like
// "rdma dev set adaptive-moderation".
func (c *Conn) SetDim(req SetDimReq) error {
	if err := c.do(cmdSet, req); err != nil {
		return fmt.Errorf("rdmanl: set adaptive moderation of device %d: %w", req.DevIndex, err)
	}

	return nil
}

// A deviceList is the devices of a dump, one per reply.
type deviceList []Device

func (l *deviceList) unmarshal(data []byte) error {
	var d Device
	if err := d.unmarshal(data); err != nil {
		return err
	}

	*l = append(*l, d)
	return nil
}

func (d *Device) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrDevIndex:
			d.Index = ad.Uint32()
		case attrDevName:
			d.Name = ad.String()
		case attrPortIndex:
			d.Ports = ad.Uint32()
		case attrCapFlags:
			d.CapFlags = ad.Uint64()
		case attrFWVersion:
			d.FWVersion = ad.String()
		case attrNodeGUID:
			d.NodeGUID = ad.Uint64()
		case attrSysImageGUID:
			d.SysImageGUID = ad.Uint64()
		case attrDevNodeType:
			d.NodeType = ad.Uint8()
		case attrDevDim:
			d.Dim = ad.Uint8() != 0
		case attrDevType:
			d.Type = DevType(ad.Uint8())
		case attrParentName:
			d.ParentName = ad.String()
		case attrNameAssignType:
			d.NameAssignType = NameAssignType(ad.Uint8())
		case attrDevProtocol:
			d.Protocol = ad.String()
		}
	}
	return ad.Err()
}
