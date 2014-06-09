package steam

import (
	"github.com/golang/glog"
	"bytes"
)

type ChallengeRequest struct {
}

func (ChallengeRequest) MarshalBinary() ([]byte) {
	buf := new(bytes.Buffer)

	writeRequestPrefix(buf)
	writeByte(buf, 'U')
	writeRequestPrefix(buf)

	glog.V(2).Infof("challengeRequest buffer: %v", buf.Bytes())
	return buf.Bytes()
}
