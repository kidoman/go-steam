package steam

import (
	"bytes"

	"github.com/golang/glog"
)

type ChallengeRequest struct {
}

func (ChallengeRequest) MarshalBinary() []byte {
	buf := new(bytes.Buffer)

	writeRequestPrefix(buf)
	writeByte(buf, 'U')
	writeRequestPrefix(buf)

	glog.V(2).Infof("steam: challengeRequest buffer: %v", buf.Bytes())
	return buf.Bytes()
}
