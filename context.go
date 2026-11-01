//go:build linux

package rdmanl

import "fmt"

// A Context is the user context of a process that opened a device.
type Context struct {
	ID       uint32
	PID      uint32
	KernName string
}

// A ContextReq asks for the user contexts of a device, or for one user context by its number:
// Contexts uses DriverDetails, Context uses CtxN.
type ContextReq struct {
	DevIndex uint32
	CtxN     uint32
	// Whether to include user contexts only drivers use
	DriverDetails bool
}

func (r ContextReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	ae.Uint32(attrResCtxN, r.CtxN)
	if r.DriverDetails {
		ae.Uint8(attrDriverDetails, 1)
	}
	return ae.Encode()
}

// Contexts returns the user contexts of a device, like "rdma resource show ctx".
func (c *Conn) Contexts(req ContextReq) ([]Context, error) {
	var l contextList
	if err := c.dump(cmdResCtxGet, req, &l); err != nil {
		return nil, fmt.Errorf("rdmanl: list user contexts of device %d: %w", req.DevIndex, err)
	}

	return l, nil
}

// Context returns a user context of a device.
func (c *Conn) Context(req ContextReq) (Context, error) {
	var ctx Context
	if err := c.get(cmdResCtxGet, req, &ctx); err != nil {
		return ctx, fmt.Errorf("rdmanl: get user context %d of device %d: %w", req.CtxN, req.DevIndex, err)
	}

	return ctx, nil
}

// A contextList is the user contexts of a dump.
type contextList []Context

func (l *contextList) unmarshal(data []byte) error {
	return unmarshalTable(data, attrResCtx, attrResCtxEntry, (*[]Context)(l))
}

func (ctx *Context) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrResCtxN:
			ctx.ID = ad.Uint32()
		case attrResPID:
			ctx.PID = ad.Uint32()
		case attrResKernName:
			ctx.KernName = ad.String()
		}
	}
	return ad.Err()
}
