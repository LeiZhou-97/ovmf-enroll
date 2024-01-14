package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UnPack the binary byte slice according to the input format
func UnPack(format []string, rawData []byte) ([]interface{}, error) {
	length, err := calculateLength(format)
	if err != nil {
		return nil, err
	} else if length > len(rawData) {
		return nil, errors.New("the format length is larger than rawData length")
	}

	res := []interface{}{}

	for _, e := range format {
		switch e {
		case "B":
			res = append(res, uint8(rawData[0]))
			rawData = rawData[1:]
		case "H":
			res = append(res, binary.LittleEndian.Uint16(rawData[:2]))
			rawData = rawData[2:]
		case "I", "L":
			res = append(res, binary.LittleEndian.Uint32(rawData[:4]))
			rawData = rawData[4:]
		case "Q":
			res = append(res, binary.LittleEndian.Uint64(rawData[:8]))
			rawData = rawData[8:]
		default:
			if strings.Contains(e, "s") {
				n, _ := strconv.Atoi(strings.TrimRight(e, "s"))
				res = append(res, rawData[:n])
				rawData = rawData[n:]
			} else {
				return nil, errors.New("the input format string contains unrecognized character")
			}
		}
	}

	return res, nil
}

func Pack(format []string, rawObj []interface{}) ([]byte, error) {

	res := []byte{}

	for i, e := range format {
		buf := bytes.NewBuffer([]byte{})
		switch e {
		case "B":
			tmp, ok := rawObj[i].(byte)
			if !ok {
				return nil, errors.New("format B assert error")
			}
			res = append(res, tmp)
		case "H":
			tmp, ok := rawObj[i].(uint16)
			if !ok {
				return nil, errors.New("format H assert error")
			}
			binary.Write(buf, binary.LittleEndian, tmp)
			res = append(res, buf.Bytes()...)
		case "I", "L":
			tmp, ok := rawObj[i].(uint32)
			if !ok {
				return nil, errors.New("format I or L assert error")
			}
			binary.Write(buf, binary.LittleEndian, tmp)
			res = append(res, buf.Bytes()...)
		case "Q":
			tmp, ok := rawObj[i].(uint64)
			if !ok {
				return nil, errors.New("format Q assert error")
			}
			binary.Write(buf, binary.LittleEndian, tmp)
			res = append(res, buf.Bytes()...)
		default:
			if strings.Contains(e, "s") {
				tmp, ok := rawObj[i].([]byte)
				if !ok {
					return nil, errors.New("Type s assert error")
				}
				// n, _ := strconv.Atoi(strings.TrimRight(e, "s"))
				res = append(res, tmp...)
			} else {
				return nil, errors.New("Type s assert error")
			}
		}
	}

	return res, nil
}

func calculateLength(format []string) (int, error) {
	length := 0
	for _, e := range format {
		switch e {
		case "B":
			length++
		case "H":
			length += 2
		case "I", "L":
			length += 4
		case "Q":
			length += 8
		default:
			if strings.Contains(e, "s") {
				n, err := strconv.Atoi(strings.TrimRight(e, "s"))
				if err != nil {
					return -1, err
				}
				length += n
			} else {
				return -1, errors.New("the input format string contains unrecognized character")
			}
		}
	}
	return length, nil
}

func Bytes2Guid(bytes []byte) (uuid.UUID, error) {
	guid := make([]byte, len(bytes))
	copy(guid, bytes)
	slices.Reverse(guid[:4])
	slices.Reverse(guid[4:6])
	slices.Reverse(guid[6:8])

	u, err := uuid.FromBytes(guid)
	if err != nil {
		return uuid.UUID{}, err
	}
	return u, nil
}

func Str2Blob(str string) ([]byte, error) {
	var res []byte
	for _, e := range str {
		tmp, err := Pack([]string{"H"}, []interface{}{uint16(e)})
		res = append(res, tmp...)
		if err != nil {
			return res, err
		}
	}

	res = append(res, []byte{0, 0}...)

	return res, nil
}

func Str2Guid(str string) ([]byte, error) {
	u, err := uuid.Parse(str)
	if err != nil {
		return nil, err
	}
	b, err := u.MarshalBinary()
	if err != nil {
		return nil, err
	}
	slices.Reverse(b[:4])
	slices.Reverse(b[4:6])
	slices.Reverse(b[6:8])

	return b, nil
}

type EFITime struct {
	year       uint16
	month      uint8
	day        uint8
	hour       uint8
	minute     uint8
	second     uint8
	pad1       uint8
	nanosecond uint32
	timezone   uint16
	daylight   uint8
	pad2       uint8
}

func (e *EFITime) Now() EFITime {
	t := time.Now()
	efiTime := EFITime{
		year: uint16(t.Year()),
		month: uint8(t.Month()),
		day: uint8(t.Day()),
		hour: uint8(t.Hour()),
		minute: uint8(t.Minute()),
		second: uint8(t.Second()),
		pad1: 0,
		nanosecond: 0,
		timezone: 0,
		daylight: 0,
		pad2: 0,
	}
	return efiTime
}

func (e *EFITime) MarshalBinary() ([]byte, error) {
	format := []string{"H", "B", "B", "B", "B", "B", "B", "I", "H", "B", "B"}
	b, err := Pack(format, []interface{}{
		e.year,
		e.month,
		e.day,
		e.hour,
		e.minute,
		e.second,
		e.pad1,
		e.nanosecond,
		e.timezone,
		e.daylight,
		e.pad2,
	})

	if err != nil {
		return []byte{}, err
	}
	return b, nil
}
