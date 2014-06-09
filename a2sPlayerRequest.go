package steam

import (
	"github.com/golang/glog"
	"bytes"
)

type A2SPlayerRequest struct {
}

func (A2SPlayerRequest) MarshalBinaryFromChallenge(c ChallengeResponse) ([]byte) {
	buf := new(bytes.Buffer)

	writeRequestPrefix(buf)
	writeByte(buf, 'U')
	buf.Write(c.GetChallangeNumber())

	glog.V(2).Infof("a2SPlayerRequest buffer: %v", buf.Bytes())
	return buf.Bytes()
}
