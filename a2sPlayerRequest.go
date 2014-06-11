package steam

import (
	"bytes"

	"github.com/golang/glog"
)

type A2SPlayerRequest struct {
}

func (A2SPlayerRequest) MarshalBinaryFromChallenge(c ChallengeResponse) []byte {
	buf := new(bytes.Buffer)

	writeRequestPrefix(buf)
	writeByte(buf, 'U')
	buf.Write(c.GetChallange())

	glog.V(2).Infof("steam: a2SPlayerRequest buffer: %v", buf.Bytes())
	return buf.Bytes()
}
