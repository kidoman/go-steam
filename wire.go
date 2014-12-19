package steam

import (
	"bytes"
	"encoding/binary"
	"io"
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

func readByte(r io.Reader) byte {
	buf := make([]byte, 1)
	_, err := io.ReadFull(r, buf)
	must(err)
	return buf[0]
}

func readBytes(r io.Reader, n int) (buf []byte) {
	buf = make([]byte, n)
	_, err := io.ReadFull(r, buf)
	must(err)
	return
}

func readShort(r io.Reader) (v int16) {
	must(binary.Read(r, binary.LittleEndian, &v))
	return
}

func readLong(r io.Reader) (v int32) {
	must(binary.Read(r, binary.LittleEndian, &v))
	return
}

func readULong(r io.Reader) (v uint32) {
	must(binary.Read(r, binary.LittleEndian, &v))
	return
}

func readLongLong(r io.Reader) (v int64) {
	must(binary.Read(r, binary.LittleEndian, &v))
	return
}

func readString(r io.Reader) string {
	if buf, ok := r.(*bytes.Buffer); ok {
		// See if we are being passed a bytes.Buffer.
		// Fast path.
		bytes, err := buf.ReadBytes(0)
		must(err)
		return string(bytes)
	}
	var buf bytes.Buffer
	for {
		b := make([]byte, 1)
		_, err := io.ReadFull(r, b)
		must(err)
		buf.WriteByte(b[0])
		if b[0] == 0 {
			break
		}
	}
	return buf.String()
}

func readFloat(r io.Reader) float32 {
	v := readULong(r)
	return math.Float32frombits(v)
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
			panic(errBadData)
		}
		return i
	}
	panic(errBadData)
}

func writeRequestPrefix(buf *bytes.Buffer) {
	buf.Write(requestPrefix)
}

var requestPrefix = []byte{0xFF, 0xFF, 0xFF, 0xFF}

func writeString(buf *bytes.Buffer, v string) {
	buf.WriteString(v)
	buf.WriteByte(0)
}

func writeByte(buf *bytes.Buffer, v byte) {
	buf.WriteByte(v)
}

func writeLong(buf *bytes.Buffer, v int32) {
	must(binary.Write(buf, binary.LittleEndian, v))
}

func writeNull(buf *bytes.Buffer) {
	buf.WriteByte(0)
}
