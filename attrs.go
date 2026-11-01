//go:build linux

package rdmanl

import (
	"encoding/binary"

	"github.com/mdlayher/netlink"
)

// Attributes of RDMA_NL_NLDEV messages. Every command shares this one
// numbering, and replies carry them in host byte order.
//
// https://github.com/torvalds/linux/blob/v7.3/include/uapi/rdma/rdma_netlink.h#L323
const (
	attrDevIndex                   = 0x01 // RDMA_NLDEV_ATTR_DEV_INDEX
	attrDevName                    = 0x02 // RDMA_NLDEV_ATTR_DEV_NAME
	attrPortIndex                  = 0x03 // RDMA_NLDEV_ATTR_PORT_INDEX
	attrCapFlags                   = 0x04 // RDMA_NLDEV_ATTR_CAP_FLAGS
	attrFWVersion                  = 0x05 // RDMA_NLDEV_ATTR_FW_VERSION
	attrNodeGUID                   = 0x06 // RDMA_NLDEV_ATTR_NODE_GUID
	attrSysImageGUID               = 0x07 // RDMA_NLDEV_ATTR_SYS_IMAGE_GUID
	attrSubnetPrefix               = 0x08 // RDMA_NLDEV_ATTR_SUBNET_PREFIX
	attrLID                        = 0x09 // RDMA_NLDEV_ATTR_LID
	attrSMLID                      = 0x0a // RDMA_NLDEV_ATTR_SM_LID
	attrLMC                        = 0x0b // RDMA_NLDEV_ATTR_LMC
	attrPortState                  = 0x0c // RDMA_NLDEV_ATTR_PORT_STATE
	attrPortPhysState              = 0x0d // RDMA_NLDEV_ATTR_PORT_PHYS_STATE
	attrDevNodeType                = 0x0e // RDMA_NLDEV_ATTR_DEV_NODE_TYPE
	attrResSummary                 = 0x0f // RDMA_NLDEV_ATTR_RES_SUMMARY
	attrResSummaryEntry            = 0x10 // RDMA_NLDEV_ATTR_RES_SUMMARY_ENTRY
	attrResSummaryEntryName        = 0x11 // RDMA_NLDEV_ATTR_RES_SUMMARY_ENTRY_NAME
	attrResSummaryEntryCurr        = 0x12 // RDMA_NLDEV_ATTR_RES_SUMMARY_ENTRY_CURR
	attrResQP                      = 0x13 // RDMA_NLDEV_ATTR_RES_QP
	attrResQPEntry                 = 0x14 // RDMA_NLDEV_ATTR_RES_QP_ENTRY
	attrResLQPN                    = 0x15 // RDMA_NLDEV_ATTR_RES_LQPN
	attrResRQPN                    = 0x16 // RDMA_NLDEV_ATTR_RES_RQPN
	attrResRQPSN                   = 0x17 // RDMA_NLDEV_ATTR_RES_RQ_PSN
	attrResSQPSN                   = 0x18 // RDMA_NLDEV_ATTR_RES_SQ_PSN
	attrResPathMigState            = 0x19 // RDMA_NLDEV_ATTR_RES_PATH_MIG_STATE
	attrResType                    = 0x1a // RDMA_NLDEV_ATTR_RES_TYPE
	attrResState                   = 0x1b // RDMA_NLDEV_ATTR_RES_STATE
	attrResPID                     = 0x1c // RDMA_NLDEV_ATTR_RES_PID
	attrResKernName                = 0x1d // RDMA_NLDEV_ATTR_RES_KERN_NAME
	attrResCMID                    = 0x1e // RDMA_NLDEV_ATTR_RES_CM_ID
	attrResCMIDEntry               = 0x1f // RDMA_NLDEV_ATTR_RES_CM_ID_ENTRY
	attrResPS                      = 0x20 // RDMA_NLDEV_ATTR_RES_PS
	attrResSrcAddr                 = 0x21 // RDMA_NLDEV_ATTR_RES_SRC_ADDR
	attrResDstAddr                 = 0x22 // RDMA_NLDEV_ATTR_RES_DST_ADDR
	attrResCQ                      = 0x23 // RDMA_NLDEV_ATTR_RES_CQ
	attrResCQEntry                 = 0x24 // RDMA_NLDEV_ATTR_RES_CQ_ENTRY
	attrResCQE                     = 0x25 // RDMA_NLDEV_ATTR_RES_CQE
	attrResUseCnt                  = 0x26 // RDMA_NLDEV_ATTR_RES_USECNT
	attrResPollCtx                 = 0x27 // RDMA_NLDEV_ATTR_RES_POLL_CTX
	attrResMR                      = 0x28 // RDMA_NLDEV_ATTR_RES_MR
	attrResMREntry                 = 0x29 // RDMA_NLDEV_ATTR_RES_MR_ENTRY
	attrResRKey                    = 0x2a // RDMA_NLDEV_ATTR_RES_RKEY
	attrResLKey                    = 0x2b // RDMA_NLDEV_ATTR_RES_LKEY
	attrResMRLen                   = 0x2d // RDMA_NLDEV_ATTR_RES_MRLEN
	attrResPD                      = 0x2e // RDMA_NLDEV_ATTR_RES_PD
	attrResPDEntry                 = 0x2f // RDMA_NLDEV_ATTR_RES_PD_ENTRY
	attrResLocalDMALKey            = 0x30 // RDMA_NLDEV_ATTR_RES_LOCAL_DMA_LKEY
	attrResUnsafeGlobalRKey        = 0x31 // RDMA_NLDEV_ATTR_RES_UNSAFE_GLOBAL_RKEY
	attrNdevIndex                  = 0x32 // RDMA_NLDEV_ATTR_NDEV_INDEX
	attrNdevName                   = 0x33 // RDMA_NLDEV_ATTR_NDEV_NAME
	attrDriver                     = 0x34 // RDMA_NLDEV_ATTR_DRIVER
	attrDriverString               = 0x36 // RDMA_NLDEV_ATTR_DRIVER_STRING
	attrDriverPrintType            = 0x37 // RDMA_NLDEV_ATTR_DRIVER_PRINT_TYPE
	attrDriverS32                  = 0x38 // RDMA_NLDEV_ATTR_DRIVER_S32
	attrDriverU32                  = 0x39 // RDMA_NLDEV_ATTR_DRIVER_U32
	attrDriverS64                  = 0x3a // RDMA_NLDEV_ATTR_DRIVER_S64
	attrDriverU64                  = 0x3b // RDMA_NLDEV_ATTR_DRIVER_U64
	attrResPDN                     = 0x3c // RDMA_NLDEV_ATTR_RES_PDN
	attrResCQN                     = 0x3d // RDMA_NLDEV_ATTR_RES_CQN
	attrResMRN                     = 0x3e // RDMA_NLDEV_ATTR_RES_MRN
	attrResCMIDN                   = 0x3f // RDMA_NLDEV_ATTR_RES_CM_IDN
	attrResCtxN                    = 0x40 // RDMA_NLDEV_ATTR_RES_CTXN
	attrLinkType                   = 0x41 // RDMA_NLDEV_ATTR_LINK_TYPE
	attrSysNetNSMode               = 0x42 // RDMA_NLDEV_SYS_ATTR_NETNS_MODE
	attrDevProtocol                = 0x43 // RDMA_NLDEV_ATTR_DEV_PROTOCOL
	attrNetNSFD                    = 0x44 // RDMA_NLDEV_NET_NS_FD
	attrChardevType                = 0x45 // RDMA_NLDEV_ATTR_CHARDEV_TYPE
	attrChardevName                = 0x46 // RDMA_NLDEV_ATTR_CHARDEV_NAME
	attrChardevABI                 = 0x47 // RDMA_NLDEV_ATTR_CHARDEV_ABI
	attrChardev                    = 0x48 // RDMA_NLDEV_ATTR_CHARDEV
	attrUverbsDriverID             = 0x49 // RDMA_NLDEV_ATTR_UVERBS_DRIVER_ID
	attrStatMode                   = 0x4a // RDMA_NLDEV_ATTR_STAT_MODE
	attrStatRes                    = 0x4b // RDMA_NLDEV_ATTR_STAT_RES
	attrStatAutoModeMask           = 0x4c // RDMA_NLDEV_ATTR_STAT_AUTO_MODE_MASK
	attrStatCounter                = 0x4d // RDMA_NLDEV_ATTR_STAT_COUNTER
	attrStatCounterEntry           = 0x4e // RDMA_NLDEV_ATTR_STAT_COUNTER_ENTRY
	attrStatCounterID              = 0x4f // RDMA_NLDEV_ATTR_STAT_COUNTER_ID
	attrStatHWCounters             = 0x50 // RDMA_NLDEV_ATTR_STAT_HWCOUNTERS
	attrStatHWCounterEntry         = 0x51 // RDMA_NLDEV_ATTR_STAT_HWCOUNTER_ENTRY
	attrStatHWCounterEntryName     = 0x52 // RDMA_NLDEV_ATTR_STAT_HWCOUNTER_ENTRY_NAME
	attrStatHWCounterEntryValue    = 0x53 // RDMA_NLDEV_ATTR_STAT_HWCOUNTER_ENTRY_VALUE
	attrDevDim                     = 0x54 // RDMA_NLDEV_ATTR_DEV_DIM
	attrResRaw                     = 0x55 // RDMA_NLDEV_ATTR_RES_RAW
	attrResCtx                     = 0x56 // RDMA_NLDEV_ATTR_RES_CTX
	attrResCtxEntry                = 0x57 // RDMA_NLDEV_ATTR_RES_CTX_ENTRY
	attrResSRQ                     = 0x58 // RDMA_NLDEV_ATTR_RES_SRQ
	attrResSRQEntry                = 0x59 // RDMA_NLDEV_ATTR_RES_SRQ_ENTRY
	attrResSRQN                    = 0x5a // RDMA_NLDEV_ATTR_RES_SRQN
	attrMinRange                   = 0x5b // RDMA_NLDEV_ATTR_MIN_RANGE
	attrMaxRange                   = 0x5c // RDMA_NLDEV_ATTR_MAX_RANGE
	attrSysCopyOnFork              = 0x5d // RDMA_NLDEV_SYS_ATTR_COPY_ON_FORK
	attrStatHWCounterIndex         = 0x5e // RDMA_NLDEV_ATTR_STAT_HWCOUNTER_INDEX
	attrStatHWCounterDynamic       = 0x5f // RDMA_NLDEV_ATTR_STAT_HWCOUNTER_DYNAMIC
	attrSysPrivilegedQKeyMode      = 0x60 // RDMA_NLDEV_SYS_ATTR_PRIVILEGED_QKEY_MODE
	attrDriverDetails              = 0x61 // RDMA_NLDEV_ATTR_DRIVER_DETAILS
	attrResSubtype                 = 0x62 // RDMA_NLDEV_ATTR_RES_SUBTYPE
	attrDevType                    = 0x63 // RDMA_NLDEV_ATTR_DEV_TYPE
	attrParentName                 = 0x64 // RDMA_NLDEV_ATTR_PARENT_NAME
	attrNameAssignType             = 0x65 // RDMA_NLDEV_ATTR_NAME_ASSIGN_TYPE
	attrEventType                  = 0x66 // RDMA_NLDEV_ATTR_EVENT_TYPE
	attrSysMonitorMode             = 0x67 // RDMA_NLDEV_SYS_ATTR_MONITOR_MODE
	attrStatOpcounterEnabled       = 0x68 // RDMA_NLDEV_ATTR_STAT_OPCOUNTER_ENABLED
	attrFRMRPools                  = 0x69 // RDMA_NLDEV_ATTR_FRMR_POOLS
	attrFRMRPoolEntry              = 0x6a // RDMA_NLDEV_ATTR_FRMR_POOL_ENTRY
	attrFRMRPoolKey                = 0x6b // RDMA_NLDEV_ATTR_FRMR_POOL_KEY
	attrFRMRPoolKeyATS             = 0x6c // RDMA_NLDEV_ATTR_FRMR_POOL_KEY_ATS
	attrFRMRPoolKeyAccessFlags     = 0x6d // RDMA_NLDEV_ATTR_FRMR_POOL_KEY_ACCESS_FLAGS
	attrFRMRPoolKeyVendorKey       = 0x6e // RDMA_NLDEV_ATTR_FRMR_POOL_KEY_VENDOR_KEY
	attrFRMRPoolKeyNumDMABlocks    = 0x6f // RDMA_NLDEV_ATTR_FRMR_POOL_KEY_NUM_DMA_BLOCKS
	attrFRMRPoolQueueHandles       = 0x70 // RDMA_NLDEV_ATTR_FRMR_POOL_QUEUE_HANDLES
	attrFRMRPoolMaxInUse           = 0x71 // RDMA_NLDEV_ATTR_FRMR_POOL_MAX_IN_USE
	attrFRMRPoolInUse              = 0x72 // RDMA_NLDEV_ATTR_FRMR_POOL_IN_USE
	attrFRMRPoolsAgingPeriod       = 0x73 // RDMA_NLDEV_ATTR_FRMR_POOLS_AGING_PERIOD
	attrFRMRPoolPinnedHandles      = 0x74 // RDMA_NLDEV_ATTR_FRMR_POOL_PINNED_HANDLES
	attrFRMRPoolKeyKernelVendorKey = 0x75 // RDMA_NLDEV_ATTR_FRMR_POOL_KEY_KERNEL_VENDOR_KEY
	attrResSummaryEntryMax         = 0x76 // RDMA_NLDEV_ATTR_RES_SUMMARY_ENTRY_MAX
)

func newAttributeDecoder(b []byte) (*netlink.AttributeDecoder, error) {
	ad, err := netlink.NewAttributeDecoder(b)
	if err != nil {
		return nil, err
	}

	ad.ByteOrder = binary.NativeEndian

	return ad, nil
}

func newAttributeEncoder() *netlink.AttributeEncoder {
	ae := netlink.NewAttributeEncoder()

	ae.ByteOrder = binary.NativeEndian

	return ae
}

// A request encodes the attributes of a request.
type request interface {
	marshal() ([]byte, error)
}

// A response decodes the attributes of a reply. A dump decodes all its
// replies into the same response.
type response interface {
	unmarshal(data []byte) error
}

// entry constrains P to a pointer to T that decodes itself.
type entry[T any] interface {
	*T
	response
}

// unmarshalEntries appends the entries of type typ of a nested table, such
// as the QPs of RDMA_NLDEV_ATTR_RES_QP, to items.
func unmarshalEntries[T any, P entry[T]](table []byte, typ uint16, items *[]T) error {
	ad, err := newAttributeDecoder(table)
	if err != nil {
		return err
	}

	for ad.Next() {
		if ad.Type() != typ {
			continue
		}

		var e T
		if err := P(&e).unmarshal(ad.Bytes()); err != nil {
			return err
		}

		*items = append(*items, e)
	}

	return ad.Err()
}

// unmarshalTable appends the entries of type typ of the table attribute attr
// of a reply to items.
func unmarshalTable[T any, P entry[T]](data []byte, attr, typ uint16, items *[]T) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}

	for ad.Next() {
		if ad.Type() == attr {
			if err := unmarshalEntries[T, P](ad.Bytes(), typ, items); err != nil {
				return err
			}
		}
	}

	return ad.Err()
}

// putBool encodes b as a u8 attribute, as the kernel's flags are.
func putBool(ae *netlink.AttributeEncoder, typ uint16, b bool) {
	var v uint8
	if b {
		v = 1
	}

	ae.Uint8(typ, v)
}
