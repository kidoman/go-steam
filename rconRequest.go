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

type rconRequest struct {
	size    int32
	id      int32
	reqType int32
	body    string
}

func newrconRequest(reqType int32, body string) *rconRequest {
	return &rconRequest{
		size:    int32(len(body) + 10),
		id:      rand.Int31(),
		reqType: reqType,
		body:    body,
	}
}

func (r *rconRequest) constructPacket() []byte {
	buf := new(bytes.Buffer)
	writeLilEndianInt32(buf, r.size)
	writeLilEndianInt32(buf, r.id)
	writeLilEndianInt32(buf, r.reqType)
	buf.WriteString(r.body)
	writeNullTerminator(buf)
	writeNullTerminator(buf)
	return buf.Bytes()
}
