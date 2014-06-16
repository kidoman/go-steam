package steam

import (
	"bytes"

	"github.com/golang/glog"
)

type A2SPlayerRequest struct {
	c ChallengeResponse
}

func (a A2SPlayerRequest) MarshalBinary() []byte {
	buf := new(bytes.Buffer)

	writeRequestPrefix(buf)
	writeByte(buf, 'U')
	buf.Write(a.c.GetChallange())

	glog.V(2).Infof("steam: a2SPlayerRequest buffer: %v", buf.Bytes())
	return buf.Bytes()
}
