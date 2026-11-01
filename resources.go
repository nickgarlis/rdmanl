//go:build linux

package rdmanl

import "fmt"

// A ResourceCount is how many objects of one kind a device has.
type ResourceCount struct {
	// Kind, such as "qp", "cq", "mr" or "pd"
	Name string
	Curr uint64
	// The most the device supports, or 0 if it has no limit
	Max uint64
}

// A ResourcesReq asks how many objects of each kind a device has.
type ResourcesReq struct {
	DevIndex uint32
	// Whether to count objects only drivers use
	DriverDetails bool
}

func (r ResourcesReq) marshal() ([]byte, error) {
	ae := newAttributeEncoder()
	ae.Uint32(attrDevIndex, r.DevIndex)
	if r.DriverDetails {
		ae.Uint8(attrDriverDetails, 1)
	}
	return ae.Encode()
}

// Resources returns how many objects of each kind a device has, like
// "rdma resource show".
func (c *Conn) Resources(req ResourcesReq) ([]ResourceCount, error) {
	var counts resourceCounts
	if err := c.get(cmdResGet, req, &counts); err != nil {
		return nil, fmt.Errorf("rdmanl: count resources of device %d: %w", req.DevIndex, err)
	}

	return counts, nil
}

// A rawList is the objects of a raw dump.
type rawList struct {
	attr  uint16
	entry uint16
	items []raw
}

func (l *rawList) unmarshal(data []byte) error {
	return unmarshalTable(data, l.attr, l.entry, &l.items)
}

// bytes returns the objects.
func (l *rawList) bytes() [][]byte {
	objs := make([][]byte, len(l.items))
	for i, r := range l.items {
		objs[i] = r
	}

	return objs
}

// A raw is an object in the driver's format.
type raw []byte

func (r *raw) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		if ad.Type() == attrResRaw {
			*r = ad.Bytes()
		}
	}
	return ad.Err()
}

// resourceCounts is the counts of a device's objects.
type resourceCounts []ResourceCount

func (l *resourceCounts) unmarshal(data []byte) error {
	return unmarshalTable(data, attrResSummary, attrResSummaryEntry, (*[]ResourceCount)(l))
}

func (r *ResourceCount) unmarshal(data []byte) error {
	ad, err := newAttributeDecoder(data)
	if err != nil {
		return err
	}
	for ad.Next() {
		switch ad.Type() {
		case attrResSummaryEntryName:
			r.Name = ad.String()
		case attrResSummaryEntryCurr:
			r.Curr = ad.Uint64()
		case attrResSummaryEntryMax:
			r.Max = ad.Uint64()
		}
	}
	return ad.Err()
}
