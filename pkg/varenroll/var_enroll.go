package varenroll

import (
	"errors"
	"fmt"
	"os"
	"ovmf/ovmf-enroll/pkg/utils"
	"strconv"
)

const (
	EFI_VARIABLE_GUID                          = "ddcf3616-3275-4164-98b6-fe85707ffe7d"
	EFI_AUTHENTICATED_VARIABLE_BASED_TIME_GUID = "aaf32c78-947b-439a-a180-2e144ec37792"
	EFI_AUTHENTICATED_VARIABLE_GUID            = "515fa686-b06e-4550-9112-382bf1067bfb"
	HEADER_SIZE                                = 28
	NVRAM_GUID                                 = "fff12b8d-7696-4c8b-a985-2747075b4f50"
)

type OVMFctl struct {
	Op         string //TODO: support multiple ops
	InputFile  string
	OutputFile string
	Name       string
	Guid       string
	Attributes string
	DataFile   string
}

func findVarInfo(fdData []byte) (VariableStore, error) {
	format := []string{"16s", "16s", "Q", "4s", "I", "H", "H", "H", "B", "B"}
	offset := 0
	var fv FirmwareVolume
	for ; offset < len(fdData); offset += 128 {
		res, err := utils.UnPack(format, fdData)
		if err != nil {
			return VariableStore{}, err
		}
		fv = FirmwareVolume{
			zeros:        res[0].([]byte),
			guid:         res[1].([]byte),
			size:         res[2].(uint64),
			magic:        res[3].([]byte),
			attributes:   res[4].(uint32),
			hdrlen:       res[5].(uint16),
			checksum:     res[6].(uint16),
			extHdrOffset: res[7].(uint16),
			rsvd:         res[8].(uint8),
			revision:     res[9].(uint8),
		}
		if fv.IsNVRAM() {
			fv.rawData = fdData[offset : offset+int(fv.size)]
			break
		}
	}

	if !fv.IsNVRAM() {
		return VariableStore{}, errors.New("not found NVRAM")
	}

	hdrFormat := []string{"16s", "I", "B", "B", "H", "I"}
	res, err := utils.UnPack(hdrFormat, fv.rawData[fv.hdrlen:fv.hdrlen+HEADER_SIZE])
	if err != nil {
		return VariableStore{}, err
	}

	varStore := VariableStore{
		firmwareVolume: fv,
		offsetFd:       offset + int(fv.hdrlen),
		varSize:        0,
		header:         fv.rawData[fv.hdrlen : fv.hdrlen+HEADER_SIZE],
		rawData:        fv.rawData[fv.hdrlen:],
		signature:      res[0].([]byte),
		size:           res[1].(uint32),
		format:         res[2].(uint8),
		state:          res[3].(uint8),
		rsvd:           res[4].(uint16),
		rsvd1:          res[5].(uint32),
		varList:        []TimeBasedAuthVariable{},
	}

	err = varStore.SyncVarList()
	if err != nil {
		return VariableStore{}, err
	}

	return varStore, nil
}

func VarEnroll(ovmfctl OVMFctl) error {
	fdData, err := os.ReadFile(ovmfctl.InputFile)
	if err != nil {
		return fmt.Errorf("Read OVMF File failed: %s", err)
	}

	vs, err := findVarInfo(fdData)
	if err != nil {
		return err
	}

	vs.PrintVariable()
	
	// parse attributes
	attributes, err := strconv.ParseUint(ovmfctl.Attributes, 10, 32)
	if err != nil {
		return err
	}
	// parse dataFile
	dataFile, err := os.ReadFile(ovmfctl.DataFile)
	if err != nil {
		return fmt.Errorf("Read Data File failed: %s", err)
	}

	err = vs.AddVariable(ovmfctl.Name, ovmfctl.Guid, uint32(attributes), dataFile, uint32(len(dataFile)))
	if err != nil {
		return err
	}

	err = vs.Sync2OVMF(fdData, ovmfctl.OutputFile)
	if err != nil {
		return err
	}
	return nil
}

