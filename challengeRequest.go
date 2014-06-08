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

	glog.V(3).Infof("Challenge request buffer: %v string: %v", buf.Bytes(), string(buf.Bytes()))
	return buf.Bytes()
}
