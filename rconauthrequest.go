package steam

import (
	"bytes"

	"math/rand"
)

const (
	SERVERDATA_AUTH           = 3
	SERVERDATA_EXECCOMMAND    = 2
	SERVERDATA_AUTH_RESPONSE  = 2
	SERVERDATA_RESPONSE_VALUE = 0
)

type rconauthrequest struct {
	size    int32
	id      int32
	reqType int32
	body    string
}

func newRconAuthRequest(passwd string) *rconauthrequest {
	return &rconauthrequest{
		size:    len(passwd) + 8,
		id:      rand.Int31(),
		reqType: SERVERDATA_AUTH,
		body:    passwd,
	}
}

func (r *rconauthrequest) constructPacket() []byte {
	buf := bytes.Buffer
	writeLilEndianInt32(buf, r.size)
	writeLilEndianInt32(buf, r.id)
	writeLilEndianInt32(buf, r.reqType)
	buf.WriteString(r.body)
	writeNullTerminator(buf)
	writeNullTerminator(buf)
	return buf.Bytes()
}
