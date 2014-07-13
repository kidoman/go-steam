package steam

import (
	"bytes"

	"github.com/golang/glog"
)

type rconResponse struct {
	size    int32
	id      int32
	reqType int32
	body    string
}

func newRconResponse(b []byte) *rconResponse {
	buf := bytes.NewBuffer(b)
	s := readLong(buf)
	id := readLong(buf)
	t := readLong(buf)
	body := readBytes(buf, int(s-8))

	resp := &rconResponse{s, id, t, string(body)}
	glog.V(2).Infof("steam: rconResponse: %v", resp)
	return resp
}
