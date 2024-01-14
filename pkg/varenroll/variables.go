package varenroll

import (
	"fmt"
	"ovmf/ovmf-enroll/pkg/utils"
)

// Represents the Time based authenticated Variable Header
//
//	typedef struct {
//	  UINT16      StartId;  // 0x55AA
//	  UINT8       State;    // 0x3f
//	  UINT8       Reserved;
//	  UINT32      Attributes;
//	  UINT64      MonotonicCount;
//	  EFI_TIME    TimeStamp;    // 16 bytes
//	  UINT32      PubKeyIndex;
//	  UINT32      NameSize;
//	  UINT32      DataSize;
//	  EFI_GUID    VendorGuid;
//	} VARIABLE_HEADER_TIME_BASED_AUTH;
type TimeBasedAuthVariable struct {
	startId       uint16
	state         uint8
	rsvd          uint8
	attributes    uint32
	count         uint64
	timeStamp     []byte
	pkIdx         uint32
	nameSize      uint32
	dataSize      uint32
	fullSize      uint32
	data          []byte
	vendorGuid    []byte
	rawData       []byte
	name          string
	nameBlob      []byte
	vendorGuidStr string
	hdrSize       uint32
}

// MarshalBinary converts the TimeBasedAuthVariable struct to binary
func (v *TimeBasedAuthVariable) MarshalBinary() ([]byte, error) {
	format := []string{"H", "B", "B", "I", "Q", "16s", "I", "I", "I", "16s"}
	b, err := utils.Pack(format, []interface{}{
		v.startId,
		v.state,
		v.rsvd,
		v.attributes,
		v.count,
		v.timeStamp,
		v.pkIdx,
		v.nameSize,
		v.dataSize,
		v.vendorGuid,
	})
	if err != nil {
		return []byte{}, err
	}

	b = append(b, v.nameBlob...)
	b = append(b, v.data...)

	return b, nil
}

func (v *TimeBasedAuthVariable) Dump() {
	fmt.Printf("   name             :%s\n", v.nameBlob)
	fmt.Printf("   vendor_guid      :%s\n", v.vendorGuidStr)
	fmt.Printf("   full_size        :%d\n", v.fullSize)
	fmt.Printf("   attributes       :%d\n", v.attributes)
	fmt.Printf("   state            :%d\n", v.state)
	fmt.Printf("   MonotonicCount   :%d\n", v.count)
	fmt.Printf("   PubKeyIndex      :%d\n", v.pkIdx)
	fmt.Printf("   TimeStamp        :%s\n", v.timeStamp)
	fmt.Printf("   Data             :%s\n", v.data)
}
