package varenroll

import (
	"encoding/binary"
	"os"
	"ovmf/ovmf-enroll/pkg/utils"
)

// Describe the layout of Variable Store
//
//	typedef struct {
//	  EFI_GUID  Signature;
//	  // Size of entire variable store,
//	  // including size of variable store header but not including the size of FvHeader.
//	  UINT32  Size;
//	  // Variable region format state.
//	  UINT8   Format;
//	  // Variable region healthy state.
//	  UINT8   State;
//	  UINT16  Reserved;
//	  UINT32  Reserved1;
//	} VARIABLE_STORE_HEADER;
type VariableStore struct {
	signature      []byte
	size           uint32
	format         uint8
	state          uint8
	rsvd           uint16
	rsvd1          uint32
	firmwareVolume FirmwareVolume
	offsetFd       int
	varSize        int
	header         []byte
	rawData        []byte
	varList        []TimeBasedAuthVariable
}

func (vs *VariableStore) SyncVarList() error {
	// the size of TimeBasedAuthVariable is 60
	varHdrSize := 60

	i := HEADER_SIZE
	for i < int(vs.size) {
		// check if is the valid variable
		sig := binary.LittleEndian.Uint16(vs.rawData[i : i+2])
		if sig != 0x55AA {
			break
		}

		format := []string{"H", "B", "B", "I", "Q", "16s", "I", "I", "I", "16s"}
		data := vs.rawData[i : i+varHdrSize]
		res, err := utils.UnPack(format, data)
		if err != nil {
			return err
		}
		timeBasedAuthVariable := TimeBasedAuthVariable{
			startId:    res[0].(uint16),
			state:      res[1].(uint8),
			rsvd:       res[2].(uint8),
			attributes: res[3].(uint32),
			count:      res[4].(uint64),
			timeStamp:  res[5].([]byte),
			pkIdx:      res[6].(uint32),
			nameSize:   res[7].(uint32),
			dataSize:   res[8].(uint32),
			vendorGuid: res[9].([]byte),
		}
		// fullSize = nameSize + dataSize + varHdrSize
		timeBasedAuthVariable.fullSize = timeBasedAuthVariable.nameSize + timeBasedAuthVariable.dataSize + uint32(varHdrSize)
		timeBasedAuthVariable.rawData = vs.rawData[i : i+int(timeBasedAuthVariable.fullSize)]
		timeBasedAuthVariable.nameBlob = timeBasedAuthVariable.rawData[60 : 60+timeBasedAuthVariable.nameSize]
		timeBasedAuthVariable.data = timeBasedAuthVariable.rawData[60+timeBasedAuthVariable.nameSize:]

		if guid, err := utils.Bytes2Guid(timeBasedAuthVariable.vendorGuid); err != nil {
			return err
		} else {
			timeBasedAuthVariable.vendorGuidStr = guid.String()
		}

		i += int(timeBasedAuthVariable.fullSize)
		// align start bit by 4
		i = ((i + 3) >> 2) << 2

		vs.varList = append(vs.varList, timeBasedAuthVariable)
	}

	vs.varSize = i - HEADER_SIZE

	return nil
}

// Sync2OVMF syncs the new variable list to output OVMF file 
func (vs *VariableStore) Sync2OVMF(fdData []byte, outputFile string) error {

	start := vs.offsetFd + HEADER_SIZE
	size := vs.size - HEADER_SIZE

	buf := make([]byte, start)
	copy(buf, fdData[:start])

	for _, e := range vs.varList {
		b, err := e.MarshalBinary()
		if err != nil {
			return err
		}
		// align by 4 bit and append with 0xff
		if pad := (((len(b) + 3) >> 2) << 2) - len(b); pad > 0 {
			for i := 0; i < pad; i++ {
				b = append(b, 0xff)
			}
		}
		buf = append(buf, b...)
	}

	for i := len(buf); i < start+int(size); i++ {
		buf = append(buf, 0xff)
	}

	restData := make([]byte, len(fdData)-(start+int(size)))
	copy(restData, fdData[start+int(size):])
	buf = append(buf, restData...)

	err := os.WriteFile(outputFile, buf, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VariableStore) AddVariable(name string, vendorGuidStr string, attributes uint32, buffer []byte, size uint32) error {
	var efiTime utils.EFITime
	t := efiTime.Now()
	b, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	timeBasedAuthVar := TimeBasedAuthVariable{
		startId:       0x55AA,
		state:         0x3f,
		timeStamp:     b,
		data:          buffer,
		dataSize:      size,
		attributes:    attributes,
		name:          name,
		vendorGuidStr: vendorGuidStr,
		hdrSize:       60,
	}

	if res, err := utils.Str2Blob(timeBasedAuthVar.name); err == nil {
		timeBasedAuthVar.nameBlob = res
		timeBasedAuthVar.nameSize = uint32(len(timeBasedAuthVar.nameBlob))
	} else {
		return err
	}

	if res, err := utils.Str2Guid(timeBasedAuthVar.vendorGuidStr); err == nil {
		timeBasedAuthVar.vendorGuid = res
	}

	timeBasedAuthVar.fullSize = timeBasedAuthVar.nameSize + timeBasedAuthVar.dataSize + timeBasedAuthVar.hdrSize

	vs.varList = append(vs.varList, timeBasedAuthVar)

	return nil
}

func (vs *VariableStore) PrintVariable() {
	for _, e := range vs.varList {
		e.Dump()	
	}
}
