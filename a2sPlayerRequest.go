package steam

import (
	"github.com/golang/glog"
	"bytes"
)

type A2SPlayerRequest struct {
	challengeRes ChallengeResponse
}

func (a *A2SPlayerRequest) MarshalBinary() ([]byte) {
	buf := new(bytes.Buffer)

	writeRequestPrefix(buf)
	writeByte(buf, 'U')
	buf.Write(a.challengeRes.GetChallangeNumber())

	glog.V(3).Infof("A2SPlayerRequest buffer: %v string: %v", buf.Bytes(), string(buf.Bytes()))
	return buf.Bytes()
}
