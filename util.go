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

// Dial allows overriding the dialer used to get the TCP/UDP connections
// required for communicating with the game server.
var Dial = (&net.Dialer{
	Timeout:   3 * time.Second,
	KeepAlive: 30 * time.Minute,
}).Dial
