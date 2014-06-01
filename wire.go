package steam

import (
	"bytes"
	"math"
	"strconv"
)

type parseError string

func (e parseError) Error() string {
	return string(e)
}

var errCouldNotReadData = parseError("steam: could not read data")
var errNotEnoughDataInResponse = parseError("steam: not enough data in response")
var errBadData = parseError("steam: bad data in response")

func triggerError(err parseError) {
	panic(err)
}

func readByte(buf *bytes.Buffer) byte {
	b, err := buf.ReadByte()
	if err != nil {
		triggerError(errCouldNotReadData)
	}
	return b
}

func readBytes(buf *bytes.Buffer, n int) []byte {
	b := buf.Next(n)
	if n != len(b) {
		triggerError(errNotEnoughDataInResponse)
	}
	return b
}

func readShort(buf *bytes.Buffer) int16 {
	var t [2]byte
	n, err := buf.Read(t[:])
	if err != nil {
		triggerError(errCouldNotReadData)
	}
	if n != 2 {
		triggerError(errNotEnoughDataInResponse)
	}
	return int16(uint16(t[0]) + uint16(t[1]<<8))
}

func readLong(buf *bytes.Buffer) int32 {
	var t [4]byte
	n, err := buf.Read(t[:])
	if err != nil {
		triggerError(errCouldNotReadData)
	}
	if n != 4 {
		triggerError(errNotEnoughDataInResponse)
	}
	return int32(uint32(t[0]) + uint32(t[1])<<8 + uint32(t[2])<<16 + uint32(t[3])<<24)
}

func readULong(buf *bytes.Buffer) uint32 {
	var t [4]byte
	n, err := buf.Read(t[:])
	if err != nil {
		triggerError(errCouldNotReadData)
	}
	if n != 4 {
		triggerError(errNotEnoughDataInResponse)
	}
	return uint32(uint32(t[0]) + uint32(t[1])<<8 + uint32(t[2])<<16 + uint32(t[3])<<24)
}

func toInt(v interface{}) int {
	switch v := v.(type) {
	case byte:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			triggerError(errBadData)
		}
		return i
	}

	triggerError(errBadData)

	panic("unreachable")
}

func readLongLong(buf *bytes.Buffer) int64 {
	var t [8]byte
	n, err := buf.Read(t[:])
	if err != nil {
		triggerError(errCouldNotReadData)
	}
	if n != 8 {
		triggerError(errNotEnoughDataInResponse)
	}
	return int64(uint64(t[0]) + uint64(t[1])<<8 + uint64(t[2])<<16 + uint64(t[3])<<24 + uint64(t[4])<<32 + uint64(t[5])<<40 + uint64(t[6])<<48 + uint64(t[7])<<56)
}

func readString(buf *bytes.Buffer) string {
	bytes, err := buf.ReadBytes(0)
	if err != nil {
		triggerError(errCouldNotReadData)
	}
	return string(bytes[:len(bytes)-1])
}

func readFloat(buf *bytes.Buffer) float32 {
	v := readULong(buf)
	return math.Float32frombits(v)
}

var requestPrefix = [4]byte{0xFF, 0xFF, 0xFF, 0xFF}

func writeRequestPrefix(buf *bytes.Buffer) {
	buf.Write(requestPrefix[:])
}

func writeString(buf *bytes.Buffer, v string) {
	buf.WriteString(v)
	buf.WriteByte(0)
}

func writeByte(buf *bytes.Buffer, v byte) {
	buf.WriteByte(v)
}

func writeLong(buf *bytes.Buffer, v int32) {
	bytes := [4]byte{byte(v & 0xFF), byte(v >> 8 & 0xFF), byte(v >> 16 & 0xFF), byte(v >> 24 & 0xFF)}
	buf.Write(bytes[:])
}
