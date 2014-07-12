package steam

import (
	"bytes"
)

type responsePacket struct {
	size    int32
	id      int32
	reqType int32
	body    string
}

func newResponsePacket(b []byte) *responsePacket {
	buf := bytes.NewBuffer(b)
	s := readLong(buf)
	id := readLong(buf)
	t := readLong(buf)
	body := readBytes(buf, len(b)-2)

	return &responsePacket{s, id, t, body}
}
