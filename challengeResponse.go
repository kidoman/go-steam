package steam

import (
	"github.com/golang/glog"
)

type ChallengeResponse struct {
	data []byte
}

func (c *ChallengeResponse) GetChallangeNumber() (challenge []byte) {
	glog.V(3).Infof("getting challenge number from %v", c.data)

	return c.data[(len(c.data)-4):]
}
