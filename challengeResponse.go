package steam

import (
	"fmt"
	"github.com/golang/glog"
)

type ChallengeResponse struct {
	data []byte
}

func (c *ChallengeResponse) GetChallangeNumber() (challenge []byte) {
	glog.V(2).Infoln(c)

	return c.data[(len(c.data)-4):]
}

func (c *ChallengeResponse) String() string {
	str := fmt.Sprint("challengeResponse: [")
	for i := 0; i < len(c.data); i++ {
		 str += fmt.Sprintf("%x ", c.data[i])
	}
	str += fmt.Sprint("]")
	return str
}
