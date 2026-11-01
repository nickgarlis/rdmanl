//go:build linux

package rdmanl

import (
	"fmt"

	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

// The RDMA netlink client of the device interface, part of every message
// type: RDMA_NL_GET_TYPE(RDMA_NL_NLDEV, cmd).
//
// https://github.com/torvalds/linux/blob/v7.3/include/uapi/rdma/rdma_netlink.h#L11
const (
	clientNLDev = 0x05 // RDMA_NL_NLDEV
	clientShift = 10
)

// Commands.
//
// https://github.com/torvalds/linux/blob/v7.3/include/uapi/rdma/rdma_netlink.h#L254
const (
	cmdGet           = 0x01 // RDMA_NLDEV_CMD_GET
	cmdSet           = 0x02 // RDMA_NLDEV_CMD_SET
	cmdNewLink       = 0x03 // RDMA_NLDEV_CMD_NEWLINK
	cmdDelLink       = 0x04 // RDMA_NLDEV_CMD_DELLINK
	cmdPortGet       = 0x05 // RDMA_NLDEV_CMD_PORT_GET
	cmdSysGet        = 0x06 // RDMA_NLDEV_CMD_SYS_GET
	cmdSysSet        = 0x07 // RDMA_NLDEV_CMD_SYS_SET
	cmdResGet        = 0x09 // RDMA_NLDEV_CMD_RES_GET
	cmdResQPGet      = 0x0a // RDMA_NLDEV_CMD_RES_QP_GET
	cmdResCMIDGet    = 0x0b // RDMA_NLDEV_CMD_RES_CM_ID_GET
	cmdResCQGet      = 0x0c // RDMA_NLDEV_CMD_RES_CQ_GET
	cmdResMRGet      = 0x0d // RDMA_NLDEV_CMD_RES_MR_GET
	cmdResPDGet      = 0x0e // RDMA_NLDEV_CMD_RES_PD_GET
	cmdGetChardev    = 0x0f // RDMA_NLDEV_CMD_GET_CHARDEV
	cmdStatSet       = 0x10 // RDMA_NLDEV_CMD_STAT_SET
	cmdStatGet       = 0x11 // RDMA_NLDEV_CMD_STAT_GET
	cmdStatDel       = 0x12 // RDMA_NLDEV_CMD_STAT_DEL
	cmdResQPGetRaw   = 0x13 // RDMA_NLDEV_CMD_RES_QP_GET_RAW
	cmdResCQGetRaw   = 0x14 // RDMA_NLDEV_CMD_RES_CQ_GET_RAW
	cmdResMRGetRaw   = 0x15 // RDMA_NLDEV_CMD_RES_MR_GET_RAW
	cmdResCtxGet     = 0x16 // RDMA_NLDEV_CMD_RES_CTX_GET
	cmdResSRQGet     = 0x17 // RDMA_NLDEV_CMD_RES_SRQ_GET
	cmdStatGetStatus = 0x18 // RDMA_NLDEV_CMD_STAT_GET_STATUS
	cmdResSRQGetRaw  = 0x19 // RDMA_NLDEV_CMD_RES_SRQ_GET_RAW
	cmdNewDev        = 0x1a // RDMA_NLDEV_CMD_NEWDEV
	cmdDelDev        = 0x1b // RDMA_NLDEV_CMD_DELDEV
	cmdMonitor       = 0x1c // RDMA_NLDEV_CMD_MONITOR
	cmdFRMRPoolsGet  = 0x1d // RDMA_NLDEV_CMD_FRMR_POOLS_GET
	cmdFRMRPoolsSet  = 0x1e // RDMA_NLDEV_CMD_FRMR_POOLS_SET
)

// RDMA netlink multicast groups.
//
// https://github.com/torvalds/linux/blob/v7.3/include/uapi/rdma/rdma_netlink.h#L15
const groupNotify = 0x04 // RDMA_NL_GROUP_NOTIFY

type Config struct {
	// NetNS is the network namespace to operate in. If 0, the current
	// network namespace is used.
	NetNS int
}

// A Conn is a connection to RDMA netlink.
type Conn struct {
	nlconn *netlink.Conn
	config Config
}

// Open opens a connection to RDMA netlink. config may be nil.
func Open(config *Config) (*Conn, error) {
	if config == nil {
		config = &Config{}
	}

	nlconn, err := dial(config)
	if err != nil {
		return nil, err
	}

	return &Conn{nlconn: nlconn, config: *config}, nil
}

func dial(config *Config) (*netlink.Conn, error) {
	return netlink.Dial(unix.NETLINK_RDMA, &netlink.Config{NetNS: config.NetNS})
}

func (c *Conn) Close() error {
	return c.nlconn.Close()
}

// execute sends the command cmd with the attributes of req, which may be
// nil, and returns the replies, without acknowledgements. The kernel
// reports a failed request as a *netlink.OpError.
func (c *Conn) execute(cmd uint16, flags netlink.HeaderFlags, req request) ([]netlink.Message, error) {
	msg := netlink.Message{
		Header: netlink.Header{
			Type:  netlink.HeaderType(clientNLDev<<clientShift | cmd),
			Flags: netlink.Request | flags,
		},
	}

	if req != nil {
		b, err := req.marshal()
		if err != nil {
			return nil, err
		}

		msg.Data = b
	}

	replies, err := c.nlconn.Execute(msg)
	if err != nil {
		return nil, err
	}

	msgs := replies[:0]
	for _, r := range replies {
		// TODO: check the reply type once all kernels have
		// https://lore.kernel.org/linux-rdma/20261101141521.2426198-1-nickgarlis@gmail.com/T/#t
		if r.Header.Type >= unix.NLMSG_MIN_TYPE {
			msgs = append(msgs, r)
		}
	}

	return msgs, nil
}

// get sends req and decodes the one reply into resp.
func (c *Conn) get(cmd uint16, req request, resp response) error {
	msgs, err := c.execute(cmd, 0, req)
	if err != nil {
		return err
	}

	if len(msgs) != 1 {
		return fmt.Errorf("%d replies, want 1", len(msgs))
	}

	return resp.unmarshal(msgs[0].Data)
}

// dump sends req as a dump and decodes every reply into resp.
func (c *Conn) dump(cmd uint16, req request, resp response) error {
	msgs, err := c.execute(cmd, netlink.Dump, req)
	if err != nil {
		return err
	}

	for _, m := range msgs {
		if err := resp.unmarshal(m.Data); err != nil {
			return err
		}
	}

	return nil
}

// do sends req, which changes something, and waits for the kernel to
// acknowledge it.
func (c *Conn) do(cmd uint16, req request) error {
	_, err := c.execute(cmd, netlink.Acknowledge, req)
	return err
}
