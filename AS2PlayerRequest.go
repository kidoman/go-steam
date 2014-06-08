package steam

import (
	"github.com/golang/glog"
	"bytes"
)

type AS2PlayerRequest struct {
	challengeRes ChallengeResponse
}

func (a *AS2PlayerRequest) MarshalBinary() ([]byte) {
	buf := new(bytes.Buffer)

	writeRequestPrefix(buf)
	writeByte(buf, 'U')
	buf.Write(a.challengeRes.GetChallangeNumber())

	glog.V(3).Infof("AS2PlayerRequest buffer: %v string: %v", buf.Bytes(), string(buf.Bytes()))
	return buf.Bytes()
}
