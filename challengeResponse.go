package steam

import (
	"fmt"

	"github.com/golang/glog"
)

type ChallengeResponse []byte

func (c ChallengeResponse) GetChallange() (challenge []byte) {
	glog.V(2).Infoln(c)

	return c[(len(c) - 4):]
}

func (c ChallengeResponse) String() string {
	str := fmt.Sprint("challengeResponse: [")
	for i := 0; i < len(c); i++ {
		str += fmt.Sprintf("%x ", c[i])
	}
	str += fmt.Sprint("]")
	return str
}
