package steam

import (
	"net"
	"time"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var Dial = (&net.Dialer{
	Timeout:   3 * time.Second,
	KeepAlive: 30 * time.Minute,
}).Dial
