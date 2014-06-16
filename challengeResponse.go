package steam

import (
	"bytes"
	"fmt"

	"github.com/golang/glog"
)

type ChallengeResponse []byte

func (c ChallengeResponse) GetChallange() (challenge []byte) {
	glog.V(2).Infof("steam: getting challenge from %v", c)

	return c[(len(c) - 4):]
}

func (c ChallengeResponse) String() string {
	buf := new(bytes.Buffer)

	writeString(buf, fmt.Sprint("challengeResponse: ["))
	for i := 0; i < len(c); i++ {
		writeString(buf, fmt.Sprintf("%x ", c[i]))
	}
	writeString(buf, fmt.Sprint("]"))

	return buf.String()
}
