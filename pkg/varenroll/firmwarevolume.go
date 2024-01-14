package varenroll

import (
	"bytes"
	"log"
	"ovmf/ovmf-enroll/pkg/utils"
)

// Describes the features and layout of the firmware volume.
// See PI Spec 3.2.1
//
//	struct EFI_FIRMWARE_VOLUME_HEADER {
//	    UINT8: Zeros[16]
//	    UCHAR: FileSystemGUID[16]
//	    UINT64: Length
//	    UINT32: Signature (_FVH)
//	    UINT32: Attribute mask
//	    UINT16: Header Length
//	    UINT16: Checksum
//	    UINT16: ExtHeaderOffset
//	    UINT8: Reserved[1]
//	    UINT8: Revision
//	    [<BlockMap>]+, <BlockMap(0,0)>
//	};
type FirmwareVolume struct {
	name         string
	zeros        []byte
	guid         []byte
	size         uint64
	magic        []byte
	attributes   uint32
	hdrlen       uint16
	checksum     uint16
	extHdrOffset uint16
	rsvd         byte
	revision     uint8
	rawData      []byte
}

func (fv *FirmwareVolume) IsNVRAM() bool {
	if !bytes.Equal(fv.magic, []byte("_FVH")) {
		return false
	}

	// Parse byte slice into guid format
	u, err := utils.Bytes2Guid(fv.guid)
	if err != nil {
		log.Fatalf("convert []byte to guid error: %s", err)
		return false
	}
	if u.String() != NVRAM_GUID {
		return false
	}
	return true
}
