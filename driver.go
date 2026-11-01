//go:build linux

package rdmanl

// PrintType tells how to show the value of a Driver.
//
// https://github.com/torvalds/linux/blob/v7.3/include/uapi/rdma/rdma_netlink.h#L318
type PrintType uint8

const (
	PrintTypeUnspec PrintType = 0
	PrintTypeHex    PrintType = 1
)

// A Driver is a driver-specific attribute of an object, such as the
// hardware number of an mlx5 QP: a name, and a value of one of the types.
//
// https://github.com/torvalds/linux/blob/v7.3/include/uapi/rdma/rdma_netlink.h#L482
type Driver struct {
	Name      string
	PrintType *PrintType
	String    *string
	S32       *int32
	U32       *uint32
	S64       *int64
	U64       *uint64
}

// drivers is the RDMA_NLDEV_ATTR_DRIVER table: a flat list in which
// each attribute is a name string, optionally a print type, then a value.
type drivers []Driver

func (l *drivers) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}

	var cur *Driver
	for ad.Next() {
		if cur == nil {
			if ad.Type() == attrDriverString {
				cur = &Driver{Name: ad.String()}
			}
			continue
		}

		switch ad.Type() {
		case attrDriverPrintType:
			cur.PrintType = new(PrintType(ad.Uint8()))
			continue
		case attrDriverString:
			cur.String = new(ad.String())
		case attrDriverS32:
			cur.S32 = new(int32(ad.Uint32()))
		case attrDriverU32:
			cur.U32 = new(ad.Uint32())
		case attrDriverS64:
			cur.S64 = new(int64(ad.Uint64()))
		case attrDriverU64:
			cur.U64 = new(ad.Uint64())
		default:
			continue
		}

		*l = append(*l, *cur)
		cur = nil
	}

	return ad.Err()
}
